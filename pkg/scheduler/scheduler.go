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
	tracker           core.IActivityTracker

	currentWorkerCnt atomic.Int32
	lastWorkerID     atomic.Uint32
	workerWG         sync.WaitGroup

	// adaptive scaling, nil when disabled
	autoscaler *autoscaler
}

func New(executor IExecutor, optFuncs ...types.OptFunc[opts]) *scheduler {

	// set default options
	opts := defaultOpts()

	// set custom options
	for _, fn := range optFuncs {
		fn(&opts)
	}

	// when adaptive scaling is on, size the workerQueue for the upper bound
	// so dynamically spawned workers can always register themselves
	queueCap := opts.numWorkers
	if opts.adaptive != nil && opts.adaptive.MaxWorkers > queueCap {
		queueCap = opts.adaptive.MaxWorkers
	}

	s := &scheduler{
		opts:              opts,
		executor:          executor,
		schedulerWorkPool: rp.NewPooler(rp.WithSize[schedulerWork](opts.reqResPoolSize)),
		requestPool:       rp.NewPooler(rp.WithSize[request](opts.reqResPoolSize)),
		responsePool:      rp.NewPooler(rp.WithSize[response](opts.reqResPoolSize)),
		workerQueue:       make(WorkerQueue, queueCap),
		workQueue:         make(WorkQueue, opts.workQueueSize),
		logger:            logger.EnsureLogger(nil).WithName("Scheduler"),
	}

	if opts.adaptive != nil {
		s.autoscaler = newAutoscaler(autoscalerConfig{
			minWorkers:         opts.adaptive.MinWorkers,
			maxWorkers:         opts.adaptive.MaxWorkers,
			scalingFactor:      opts.adaptive.ScalingFactor,
			scalingWindow:      opts.adaptive.ScalingWindow,
			emaAlpha:           opts.adaptive.EMAAlpha,
			currentWorkerCntFn: s.currentWorkerCnt.Load,
			spawnWorkerFn:      s.spawnWorker,
			workerQueue:        s.workerQueue,
			logger:             s.logger,
		})
	}

	return s
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

func (s *scheduler) WithActivityTracker(tracker core.IActivityTracker) engine.IScheduler {
	s.tracker = tracker
	return s
}

func (s *scheduler) Start(ctx context.Context) error {
	s.logger.Infof("Starting scheduler with %d workers", s.opts.numWorkers)

	// worker lifecycle context, cancelled in defer to signal all workers
	wCtx, wCancel := context.WithCancel(ctx)

	defer func() {
		wCancel()
		s.workerWG.Wait()
		s.logger.Info("stopped")
	}()

	if s.autoscaler != nil {
		s.autoscaler.SetDesired(uint32(s.opts.numWorkers))
	}

	for i := uint16(0); i < s.opts.numWorkers; i++ {
		s.spawnWorker(wCtx)
	}

	if s.autoscaler != nil {
		go s.autoscaler.Start(wCtx)
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

	if s.autoscaler != nil {
		s.autoscaler.OnTaskArrival()
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

func (s *scheduler) spawnWorker(ctx context.Context) {
	s.workerWG.Add(1)
	s.currentWorkerCnt.Add(1)
	id := uint16(s.lastWorkerID.Add(1))

	var recorder ts.IStatsRecorder
	if s.opts.statsFactory != nil {
		recorder = s.opts.statsFactory.NewStatsRecorder()
	}

	// wire autoscaler callbacks if enabled
	var onTaskDone func(time.Duration)
	var shouldExit func() bool

	if s.autoscaler != nil {
		onTaskDone = s.autoscaler.OnTaskDone
		shouldExit = s.autoscaler.ShouldExit
	}

	worker := NewWorker(id, s.executor, s.workerQueue, s.schedulerWorkPool, s.requestPool, s.responsePool, recorder, onTaskDone, shouldExit)

	worker.WithLogger(s.logger)
	if s.tracker != nil {
		worker.WithActivityTracker(s.tracker)
	}

	go func() {
		defer s.workerWG.Done()
		defer s.currentWorkerCnt.Add(-1)
		_ = worker.Start(ctx)
	}()
}

// For the telemetry hub
type SchedulerSnapshot struct {
	CurrentWorkerCnt uint16  `json:"current_workers"`
	DesiredWorkers   uint16  `json:"desired_workers"`
	TaskArrivalRate  float64 `json:"task_arrival_rate_ema"`
	TaskServiceTime  float64 `json:"task_service_time_ema"`
}

// implements IStatsCollector
func (s *scheduler) Name() string {
	return "Scheduler"
}

func (s *scheduler) Snapshot() ts.ComponentSnapshot {
	snap := SchedulerSnapshot{
		CurrentWorkerCnt: uint16(s.currentWorkerCnt.Load()),
	}
	if s.autoscaler != nil {
		snap.DesiredWorkers = uint16(s.autoscaler.desiredWorkerCnt.Load())
		snap.TaskArrivalRate = s.autoscaler.lambdaEMA.Value()
		snap.TaskServiceTime = s.autoscaler.serviceTimeEMA.Value()
	}
	return snap
}
