package core

import (
	"context"
	"sync"

	executer "github.com/tech-engine/goscrapy/internal/executer/http"
	rp "github.com/tech-engine/goscrapy/internal/resource_pool"
	restyadapter "github.com/tech-engine/goscrapy/pkg/executer_adapters/http/resty"
)

func New[IN Job, OUT any](ctx context.Context, scraper Scraper[IN, OUT]) *manager[IN, OUT] {

	manager := &manager[IN, OUT]{
		ctx:          ctx,
		scraper:      scraper,
		executer:     executer.NewExecuter(restyadapter.NewRestyHTTPClientAdapter()),
		requestPool:  rp.NewPooler[Request](rp.WithSize[Request](1e6)),
		responsePool: rp.NewPooler[Response](rp.WithSize[Response](1e6)),
		Pipelines:    NewPipelineManager[IN, any, OUT, Output[IN, OUT]](),
	}

	manager.scraper.SetDelegator(&ScraperDelegation[IN, OUT]{
		m: manager,
	})

	return manager
}

// start the core
func (m *manager[IN, OUT]) Start(ctx context.Context) error {

	// first start the pipelines
	if err := m.Pipelines.start(ctx); err != nil {
		return err
	}

	m.wg.Add(1)
	go m.ProcessOutput()

	return nil
}

// wait for all goroutines to exit
func (m *manager[IN, OUT]) Wait() {
	m.wg.Wait()
}

func (m *manager[IN, OUT]) Run(job IN) {
	m.wg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		m.scraper.StartRequest(m.ctx, job)
	}(&m.wg)
}

// ProcessOutput runs continuously
func (m *manager[IN, OUT]) ProcessOutput() {
	defer m.wg.Done()
	for {
		select {
		case <-m.ctx.Done():
			// execute pipelines' close hooks - blocking
			m.Pipelines.stop()
			// close spider's output channel
			m.scraper.Close()
			return
		default:
			data := m.scraper.PullResult()

			if data == nil {
				continue
			}

			m.wg.Add(1)
			// if we have data we push to pipelines
			go func(wg *sync.WaitGroup) {
				defer wg.Done()
				m.Pipelines.do(data, nil)
			}(&m.wg)
		}
	}
}

func (m *manager[IN, OUT]) NewJob(id string) IN {
	return m.scraper.NewJob(id)
}

func (m *manager[IN, OUT]) reqResCleanUp(req *Request, res *Response) {
	if req != nil {
		req.Reset()
		m.requestPool.Release(req)
	}

	if res != nil {
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
