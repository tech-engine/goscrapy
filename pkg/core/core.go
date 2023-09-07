package core

import (
	"context"
	"net/http"
	"strings"
	"sync"

	executer "github.com/tech-engine/goscrapy/internal/executer/http"
	rp "github.com/tech-engine/goscrapy/internal/resource_pool"
	httpadapter "github.com/tech-engine/goscrapy/pkg/executer_adapters/http/native"
)

func New[IN Job, OUT any](ctx context.Context, spider Spider[IN, OUT]) Manager[IN, OUT] {

	manager := &manager[IN, OUT]{
		ctx:          ctx,
		spider:       spider,
		executer:     nil,
		requestPool:  rp.NewPooler[Request](rp.WithSize[Request](1e6)),
		responsePool: rp.NewPooler[Response](rp.WithSize[Response](1e6)),
		pipelines:    NewPipelineManager[IN, any, OUT, Output[IN, OUT]](),
		middlewares:  make([]Middleware, 0),
		outputCh:     make(chan Output[IN, OUT]),
	}

	manager.spider.SetDelegator(&SpiderDelegation[IN, OUT]{
		m: manager,
	})

	return manager
}

// start the core
func (m *manager[IN, OUT]) Start(ctx context.Context) error {

	/*
		We chain all middlewares and create an http client and then an executer.
	*/
	m.executer = executer.NewExecuter(httpadapter.NewHTTPClientAdapter(&http.Client{
		Transport: m.chainMiddlewares(),
	}))

	// first start the pipelines
	if err := m.pipelines.start(ctx); err != nil {
		return err
	}

	m.wg.Add(1)
	go m.processOutput()

	return nil
}

// wait for all goroutines to exit
func (m *manager[IN, OUT]) Wait() {
	m.wg.Wait()
}

func (m *manager[IN, OUT]) Run(job IN) {
	if strings.TrimSpace(job.Id()) == "" {
		return
	}

	m.wg.Add(1)

	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		m.spider.StartRequest(m.ctx, job)
	}(&m.wg)
}

func (m *manager[IN, OUT]) close() {
	// execute pipelines' close hooks - blocking
	m.pipelines.stop()
	// close spider's output channel
	m.spider.Close(m.ctx)
	close(m.outputCh)
}

// ProcessOutput runs continuously
func (m *manager[IN, OUT]) processOutput() {
	defer m.wg.Done()
	for {
		select {
		case <-m.ctx.Done():
			m.close()
			return
		case data := <-m.outputCh:

			if data == nil {
				continue
			}

			m.wg.Add(1)
			// if we have data we push to pipelines
			go func(wg *sync.WaitGroup) {
				defer wg.Done()
				m.pipelines.do(data, nil)
			}(&m.wg)
		}
	}
}

func (m *manager[IN, OUT]) NewJob(id string) IN {
	return m.spider.NewJob(id)
}

func (m *manager[IN, OUT]) reqResCleanUp(req *Request, res *Response) {
	if req != nil {
		if req.body != nil {
			req.body.Close()
		}
		req.Reset()
		m.requestPool.Release(req)
	}

	if res != nil {
		if res.body != nil {
			res.body.Close()
		}
		res.Reset()
		m.responsePool.Release(res)
	}

}

func (m *manager[IN, OUT]) exRequest(ctx context.Context, req *Request, cb ResponseCallback) {

	if cb == nil {
		m.reqResCleanUp(req, nil)
		return
	}

	m.wg.Add(1)
	go func(wg *sync.WaitGroup, _ctx context.Context, _cb ResponseCallback) {
		defer wg.Done()
		res := m.responsePool.Acquire()

		if res == nil {
			res = &Response{}
		}

		// clean res and release it to the pool
		defer m.reqResCleanUp(req, res)

		// execute request and store response in res
		if err := m.executer.Execute(_ctx, req, res); err != nil {
			return
		}

		_cb(
			context.WithValue(_ctx, "META_DATA", req.MetaData()),
			res,
		)

	}(&m.wg, ctx, cb)
}

func (m *manager[IN, OUT]) chainMiddlewares() http.RoundTripper {

	// add all the middlewares
	roundTripper := http.DefaultTransport
	for _, middleware := range m.middlewares {
		roundTripper = middleware(roundTripper)
	}

	return roundTripper

}

func (m *manager[IN, OUT]) AddMiddlewares(middlewares ...Middleware) {
	m.middlewares = append(m.middlewares, middlewares...)
}

func (m *manager[IN, OUT]) AddPipelines(pipeline Pipeline[IN, any, OUT, Output[IN, OUT]], err error, required ...bool) {
	m.pipelines.add(pipeline, err, required...)
}
