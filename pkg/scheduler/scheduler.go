package scheduler

import (
	"context"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	rp "github.com/tech-engine/goscrapy/internal/resource_pool"
	"github.com/tech-engine/goscrapy/internal/types"
	"github.com/tech-engine/goscrapy/pkg/core"
	"github.com/tech-engine/goscrapy/pkg/logger"
	ts "github.com/tech-engine/goscrapy/pkg/telemetry/stats"
)

type scheduler struct {
	opts
	executor          IExecutor
	schedulerWorkPool *rp.Pooler[schedulerWork]
	requestPool       *rp.Pooler[request]
	workerQueue       WorkerQueue
	workQueue         WorkQueue
	stopping          atomic.Bool
	logger            core.ILogger
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
		schedulerWorkPool: rp.NewPooler(rp.WithSize[schedulerWork](opts.reqResPoolSize)),
		requestPool:       rp.NewPooler(rp.WithSize[request](opts.reqResPoolSize)),
		workerQueue:       make(WorkerQueue, opts.numWorkers),
		workQueue:         make(WorkQueue, opts.workQueueSize),
		logger:            logger.GetLogger(), // default to global logger
	}
}

func (s *scheduler) WithExecutor(executor IExecutor) {
	s.executor = executor
}

func (s *scheduler) WithLogger(l core.ILogger) {
	s.logger = l
	s.executor.WithLogger(l)
}

// Handles creating workers and listening on the work queue
func (s *scheduler) Start(ctx context.Context) error {

	if ctx.Err() != nil {
		return ctx.Err()
	}

	var (
		i  uint16
		wg sync.WaitGroup
	)

	defer wg.Wait()
	wg.Add(int(s.opts.numWorkers))

	// worker lifecyle context
	wCtx, wCancel := context.WithCancel(ctx)

	for i = 0; i < s.opts.numWorkers; i++ {
		go func() {
			defer wg.Done()

			var recorder ts.StatRecorder
			if s.opts.statsProducer != nil {
				recorder = s.opts.statsProducer.NewWorkerCollector()
			}

			worker := NewWorker(i+1, s.executor, s.workerQueue, s.schedulerWorkPool, s.requestPool, s.opts.reqResPoolSize, recorder)

			// blocking
			_ = worker.Start(wCtx)
		}()
	}

	// cancel worker context upon scheduler exit
	defer wCancel()

	for {
		select {
		case work := <-s.workQueue:
			select {
			case worker := <-s.workerQueue:
				worker <- work
			case <-ctx.Done():
				// context cancellation from top(engine).
				// we should try to put the work back or handle it.
				// in graceful shutdown, we will handle this in the drain loop.
				s.workQueue <- work
				s.stopping.Store(true)
				goto drain
			}
		case <-ctx.Done():
			s.stopping.Store(true)
			goto drain
		}
	}

drain:
	// draining the work queue
	for {
		select {
		case work := <-s.workQueue:
			select {
			case worker := <-s.workerQueue:
				worker <- work
			case <-time.After(100 * time.Millisecond):
				// if we can't get a worker within 100ms during drain,
				// it might mean workers are stuck or busy.
				s.workQueue <- work
			}
		default:
			// queue is empty
			return ctx.Err()
		}
	}
}

func (s *scheduler) Schedule(req core.IRequestReader, next core.ResponseCallback) {
	if s.stopping.Load() {
		return
	}

	work := s.schedulerWorkPool.Acquire()

	if work == nil {
		work = &schedulerWork{}
	}

	work.request = req
	work.next = next

	s.workQueue <- work
}

func (s *scheduler) NewRequest(ctx context.Context) core.IRequestRW {
	req := s.requestPool.Acquire()
	if req == nil {
		req = &request{
			method: "GET",
			header: make(http.Header),
		}
	}
	req.ctx = ctx
	return req
}
