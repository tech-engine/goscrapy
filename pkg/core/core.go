package core

import (
	"context"
	"math/rand"
	"sync"
	"time"

	executer "github.com/tech-engine/goscrapy/internal/executer/http"
	rp "github.com/tech-engine/goscrapy/internal/resource_pool"
	"github.com/tech-engine/goscrapy/internal/utils"
	restyadapter "github.com/tech-engine/goscrapy/pkg/executer_adapters/http/resty"
	metadata "github.com/tech-engine/goscrapy/pkg/meta_data"
)

func New[IN Job, OUT any](ctx context.Context, scraper Scraper[IN, OUT]) *manager[IN, OUT] {

	manager := &manager[IN, OUT]{
		mu:           &sync.RWMutex{},
		ctx:          ctx,
		scraper:      scraper,
		executer:     executer.NewExecuter(restyadapter.NewRestyHTTPClientAdapter()),
		requestPool:  rp.NewPooler[Request](rp.WithSize[Request](1e6)),
		responsePool: rp.NewPooler[Response](rp.WithSize[Response](1e6)),
		Jobs:         make(map[string]ManagerJob[IN]),
		Pipelines:    NewPipelineManager[IN, any, OUT, Output[IN, OUT]](),
	}

	manager.scraper.SetDelegator(&ScraperDelegation[IN, OUT]{
		m: manager,
	})

	go manager.Start(ctx)
	return manager
}

func (m *manager[IN, OUT]) Start(ctx context.Context) {
	go m.Pipelines.start(ctx)
	go m.scraper.Start(ctx)
	m.ProcessOutput()
	m.Pipelines.stop()
}

func (m *manager[IN, OUT]) AddJob(job ManagerJob[IN]) *manager[IN, OUT] {
	// we store it
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Jobs[job.ScraperJob.GetId()] = job
	go m.Run(job)
	return m
}

// run runs continuously in an infinite loop
func (m *manager[IN, OUT]) Run(job ManagerJob[IN]) {
	rand.Seed(time.Now().UTC().UnixNano())
	ticker := utils.NewRandomTicker(2*time.Second, 6*time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-job.ctx.Done():
			// remove job from map once we receive a terminal signal on our sigTerm channel
			m.removeJob(job.ScraperJob)
			return
		case <-ticker.C:
			m.scraper.PushJob(job.ScraperJob)
		}
	}
}

// ProcessOutput runs continuously
func (m *manager[IN, OUT]) ProcessOutput() {
	for {
		select {
		case <-m.ctx.Done():
			return
		default:
			data := m.scraper.PullResult()
			if data == nil {
				// save the job and maybe try again
				continue
			}

			m.mu.Lock()
			managerJob, ok := m.Jobs[data.UpdatedJob().GetId()]
			m.mu.Unlock()

			// if we couldn't find our job in managerJobs map, we move on
			if !ok {
				continue
			}

			if data.Error() != nil {
				// if error we signal job termination
				managerJob.cancel()
				continue
			}

			m.Pipelines.do(data, metadata.MetaData{
				"JOB_NAME": managerJob.name,
				"JOB_ID":   data.UpdatedJob().GetId(),
			})
		}
	}
}

// removeJob will release the scraper job and also removes it for our map
func (m *manager[IN, OUT]) removeJob(job Job) *manager[IN, OUT] {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.Jobs, job.GetId())
	return m
}

// removes all jobs
func (m *manager[IN, OUT]) removeAllJobs() *manager[IN, OUT] {
	for _, job := range m.Jobs {
		m.removeJob(job.ScraperJob)
	}
	return m
}

func (m *manager[IN, OUT]) NewJob(ctx context.Context, name string) ManagerJob[IN] {
	job := m.scraper.NewJob()
	return newManagerJob(ctx, name, job)
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

	go func(_ctx context.Context, _cb ResponseCallback) {
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

	}(ctx, cb)
}
