package scheduler

import (
	"context"
	"net/http"
	"sync"

	rp "github.com/tech-engine/goscrapy/internal/resource_pool"
	"github.com/tech-engine/goscrapy/internal/types"
	"github.com/tech-engine/goscrapy/pkg/core"
)

type scheduler struct {
	opts
	executor          IExecutor
	schedulerWorkPool *rp.Pooler[schedulerWork]
	requestPool       *rp.Pooler[request]
	workerQueue       WorkerQueue
	workQueue         WorkQueue
}

// NewScheduler creates a new scheduler.
func New(executor IExecutor, optFuncs ...types.OptFunc[opts]) *scheduler {

	// set default options
	opts := defaultOpts()

	// set custom options
	for _, fn := range optFuncs {
		fn(&opts)
	}

	return &scheduler{
		opts:              opts,
		executor:          executor,
		schedulerWorkPool: rp.NewPooler[schedulerWork](rp.WithSize[schedulerWork](opts.reqResPoolSize)),
		requestPool:       rp.NewPooler[request](rp.WithSize[request](opts.reqResPoolSize)),
		workerQueue:       make(WorkerQueue, opts.numWorkers),
		workQueue:         make(WorkQueue, opts.workQueueSize),
	}
}

func (s *scheduler) WithExecutor(executor IExecutor) {
	s.executor = executor
}

// Handles creating workers and listening on the work queue
func (s *scheduler) Start(ctx context.Context) error {

	if ctx.Err() != nil {
		return ctx.Err()
	}

	var (
		i   uint16
		err error
		wg  sync.WaitGroup
	)

	defer wg.Wait()
	wg.Add(int(s.opts.numWorkers))

	// this is to make sure that we close the scheduler and after that close all the workers
	wCtx, wCancel := context.WithCancel(context.Background())

	for i = 0; i < s.opts.numWorkers; i++ {
		go func(i uint16) {
			defer wg.Done()
			worker := NewWorker(i+1, s.executor, s.workerQueue, s.schedulerWorkPool, s.requestPool, s.opts.reqResPoolSize)

			// blocking
			_ = worker.Start(wCtx)
		}(i)
	}

	// below will trigger context cancellation for the worker after scheduler is done.
	defer wCancel()

	for {
		select {
		case work := <-s.workQueue:

			// the below check ensures our scheduler don't pick any worker once context has been cancelled
			if err = ctx.Err(); err != nil {
				return err
			}

			wg.Add(1)
			go s.push(&wg, work)
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (s *scheduler) Schedule(req core.IRequestReader, next core.ResponseCallback) {

	work := s.schedulerWorkPool.Acquire()

	if work == nil {
		work = &schedulerWork{}
	}

	work.request = req
	work.next = next

	s.workQueue <- work
}

func (s *scheduler) NewRequest() core.IRequestRW {
	req := s.requestPool.Acquire()
	if req == nil {
		req = &request{
			method: "GET",
			header: make(http.Header),
		}
	}
	return req
}

// push a *schedulerWork unit to a worker
func (s *scheduler) push(wg *sync.WaitGroup, work *schedulerWork) {
	defer wg.Done()

	// pull a worker and push a task in the worker's queue
	worker := <-s.workerQueue
	worker <- work
}
