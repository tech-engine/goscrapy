package scheduler

import (
	"context"
	"net/http"
	"sync"
	"sync/atomic"

	rp "github.com/tech-engine/goscrapy/internal/resource_pool"
	"github.com/tech-engine/goscrapy/internal/types"
	"github.com/tech-engine/goscrapy/pkg/core"
	"github.com/tech-engine/goscrapy/pkg/engine"
	"github.com/tech-engine/goscrapy/pkg/logger"
	ts "github.com/tech-engine/goscrapy/pkg/telemetry/stats"
)

type scheduler struct {
	opts
	executor          IExecutor
	schedulerWorkPool *rp.Pooler[schedulerWork]
	requestPool       *rp.Pooler[request]
	responsePool      *rp.Pooler[response]
	workerQueue       WorkerQueue
	workQueue         WorkQueue
	stopping          atomic.Bool
	logger            core.ILogger
}

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
		responsePool:      rp.NewPooler(rp.WithSize[response](opts.reqResPoolSize)),
		workerQueue:       make(WorkerQueue, opts.numWorkers),
		workQueue:         make(WorkQueue, opts.workQueueSize),
		logger:            logger.EnsureLogger(nil).WithName("Scheduler"),
	}
}

func (s *scheduler) WithExecutor(executor IExecutor) {
	s.executor = executor
}

func (s *scheduler) WithLogger(loggerIn core.ILogger) engine.IScheduler {
	loggerIn = logger.EnsureLogger(loggerIn)
	s.logger = loggerIn.WithName("Scheduler")
	s.executor.WithLogger(loggerIn)
	return s
}

func (s *scheduler) WithStatsRecorderFactory(f ts.IStatsRecorderFactory) {
	s.opts.statsFactory = f
}

func (s *scheduler) Start(ctx context.Context) error {
	s.logger.Infof("Starting scheduler with %d workers", s.opts.numWorkers)

	var wg sync.WaitGroup
	// worker lifecycle context
	wCtx, wCancel := context.WithCancel(context.Background())

	wg.Add(int(s.opts.numWorkers))

	defer func() {
		wCancel()
		wg.Wait()
		s.logger.Info("stopped")
	}()

	var recorders []ts.IStatsRecorder
	if s.opts.statsFactory != nil {
		recorders = make([]ts.IStatsRecorder, s.opts.numWorkers)
	}

	for i := uint16(0); i < s.opts.numWorkers; i++ {
		var recorder ts.IStatsRecorder
		if s.opts.statsFactory != nil {
			recorder = s.opts.statsFactory.NewStatsRecorder()
			recorders[i] = recorder
		}

		worker := NewWorker(i+1, s.executor, s.workerQueue, s.schedulerWorkPool, s.requestPool, s.responsePool, recorder)
		worker.WithLogger(s.logger)
		go func() {
			defer wg.Done()
			// blocking call
			_ = worker.Start(wCtx)
		}()
	}

	for {
		select {
		case <-ctx.Done():
			s.stopping.Store(true)
			s.logger.Infof("received context cancellation: %v", ctx.Err())
			return nil
		case work := <-s.workQueue:
			select {
			case worker := <-s.workerQueue:
				worker <- work
			case <-ctx.Done():
				s.stopping.Store(true)
				s.logger.Infof("received context cancellation during work dispatch: %v", ctx.Err())
				return nil
			}
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
