package worker

import (
	"context"
	"io"
	"os"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/tech-engine/goscrapy/pkg/core"
	"github.com/tech-engine/goscrapy/pkg/engine"
	"github.com/tech-engine/goscrapy/pkg/logger"
	"github.com/tech-engine/goscrapy/pkg/signal"
	ts "github.com/tech-engine/goscrapy/pkg/telemetry/stats"
)

var discardBufPool = sync.Pool{
	New: func() any {
		b := make([]byte, 32*1024)
		return &b
	},
}

type Config struct {
	Executor      IExecutor
	Results       chan engine.IResult
	QueueSize     int
	Autoscaler    *AutoscalerConfig
	Logger        core.ILogger
	EnableMetrics bool
	Signals       signal.RequestBus
}

// WorkerPoolSnapshot is returned by the worker pool collector.
type WorkerPoolSnapshot struct {
	ActiveWorkers  int32
	QueueDepth     int
	QueueCapacity  int
	TasksSubmitted uint64
	TasksDropped   uint64
}

// poolMetrics groups telemetry only fields for the worker pool.
type poolMetrics struct {
	enabled        bool
	tasksSubmitted atomic.Uint64
	tasksDropped   atomic.Uint64
}

func (m *poolMetrics) recordSubmit() {
	if m.enabled {
		m.tasksSubmitted.Add(1)
	}
}

func (m *poolMetrics) recordDrop() {
	if m.enabled {
		m.tasksDropped.Add(1)
	}
}

func (m *poolMetrics) snapshot(activeWorkers int32, queueDepth, queueCap int) WorkerPoolSnapshot {
	return WorkerPoolSnapshot{
		ActiveWorkers:  activeWorkers,
		QueueDepth:     queueDepth,
		QueueCapacity:  queueCap,
		TasksSubmitted: m.tasksSubmitted.Load(),
		TasksDropped:   m.tasksDropped.Load(),
	}
}

type workerPool struct {
	executor         IExecutor
	results          chan engine.IResult
	workerTaskBuffer chan *workTask
	workerTaskPool   sync.Pool
	resultPool       sync.Pool
	autoscaler       *autoscaler
	activeWorkers    atomic.Int32
	lastWorkerID     atomic.Uint32
	responsePool     sync.Pool
	logger           core.ILogger
	wg               sync.WaitGroup
	metrics          poolMetrics
	signals          signal.RequestBus
}

func (p *workerPool) Name() string { return "WorkerPool" }

func (p *workerPool) Snapshot() ts.ComponentSnapshot {
	return p.metrics.snapshot(
		p.activeWorkers.Load(),
		len(p.workerTaskBuffer),
		cap(p.workerTaskBuffer),
	)
}

func NewPool(config *Config) (engine.IWorkerPool, error) {
	if config == nil {
		config = &Config{}
	}

	if config.Logger == nil {
		config.Logger = logger.EnsureLogger(nil).WithName("WorkerPool")
	}

	if config.Executor == nil {
		return nil, ErrExecutorRequired
	}

	// defaults
	results := config.Results
	if results == nil {
		results = make(chan engine.IResult, 1000)
	}

	queueSize := config.QueueSize
	if queueSize <= 0 {
		queueSize = 1000
	}

	p := &workerPool{
		executor:         config.Executor,
		results:          results,
		workerTaskBuffer: make(chan *workTask, queueSize),
		logger:           logger.EnsureLogger(config.Logger).WithName("WorkerPool"),
		metrics:          poolMetrics{enabled: config.EnableMetrics},
		signals:          config.Signals,
	}

	// read from env vars for tuning
	var maxWorkers uint32
	var minWorkers uint32
	var scalingFactor float32

	if config.Autoscaler != nil {
		maxWorkers = config.Autoscaler.MaxWorkers
		minWorkers = config.Autoscaler.MinWorkers
		scalingFactor = config.Autoscaler.ScalingFactor
	}

	if maxWorkers == 0 {
		if m := os.Getenv("AUTOSCALER_MAX_WORKERS"); m != "" {
			if v, err := strconv.ParseUint(m, 10, 32); err == nil {
				maxWorkers = uint32(v)
			}
		}
		// Fallback to adaptive default
		if maxWorkers == 0 {
			maxWorkers = uint32(runtime.NumCPU() * 8)
			if maxWorkers < 16 {
				maxWorkers = 16
			}
		}
	}

	if minWorkers == 0 {
		if m := os.Getenv("AUTOSCALER_MIN_WORKERS"); m != "" {
			if v, err := strconv.ParseUint(m, 10, 32); err == nil {
				minWorkers = uint32(v)
			}
		} else {
			minWorkers = maxWorkers / 2
		}
	}

	if scalingFactor <= 0 {
		if v := os.Getenv("AUTOSCALER_SCALING_FACTOR"); v != "" {
			if v, err := strconv.ParseFloat(v, 32); err == nil && v > 0 {
				scalingFactor = float32(v)
			}
		}
	}
	if scalingFactor <= 0 {
		scalingFactor = 1.0
	}

	scalingWindow := 5 * time.Second
	if config.Autoscaler != nil && config.Autoscaler.ScalingWindow > 0 {
		scalingWindow = config.Autoscaler.ScalingWindow
	}

	emaAlpha := float32(0.3)
	if config.Autoscaler != nil && config.Autoscaler.EMAAlpha > 0 {
		emaAlpha = config.Autoscaler.EMAAlpha
	}

	autoscalerConfig := &AutoscalerConfig{
		MaxWorkers:    maxWorkers,
		MinWorkers:    minWorkers,
		ScalingFactor: scalingFactor,
		ScalingWindow: scalingWindow,
		EMAAlpha:      emaAlpha,
		currentWorkerCntFn: func() int32 {
			return p.activeWorkers.Load()
		},
		spawnWorkerFn: func(ctx context.Context) {
			p.spawnWorker(ctx)
		},
		despawnWorkerFn: func() {
			p.despawnWorker()
		},
		logger: config.Logger,
	}

	p.autoscaler = newAutoscaler(autoscalerConfig)

	p.responsePool.New = func() any {
		return &response{}
	}

	p.workerTaskPool.New = func() any {
		return &workTask{}
	}

	p.resultPool.New = func() any {
		return &result{}
	}
	return p, nil
}

func (p *workerPool) Start(ctx context.Context) error {
	p.autoscaler.SetDesired(uint32(p.autoscaler.minWorkers))
	for i := uint16(0); i < p.autoscaler.minWorkers; i++ {
		p.spawnWorker(ctx)
	}

	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		p.autoscaler.Start(ctx)
	}()

	// wait for shutdown
	<-ctx.Done()

	// wait for all workers to finish
	p.logger.Debug("Waiting for workers to finish...")
	p.wg.Wait()
	p.logger.Debug("Worker pool shut down.")

	return nil
}

func (p *workerPool) spawnWorker(ctx context.Context) {
	id := uint16(p.lastWorkerID.Add(1))
	p.activeWorkers.Add(1)

	w := NewWorker(id, p.executor, p.workerTaskBuffer, p.results, &p.responsePool, &p.workerTaskPool, &p.resultPool, p)

	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		defer p.activeWorkers.Add(-1)
		_ = w.Start(ctx)
	}()
}

// despawnWorker sends a poison pill (nil) to the task buffer, causing one
// idle worker to exit its loop and terminate
func (p *workerPool) despawnWorker() {
	select {
	case p.workerTaskBuffer <- nil:
	default:
	}
}

func (p *workerPool) Submit(req *core.Request, callbackName string, handle core.TaskHandle) error {
	p.autoscaler.OnTaskArrival()

	task := p.workerTaskPool.Get().(*workTask)
	task.req = req
	task.callbackName = callbackName
	task.taskHandle = handle

	select {
	case p.workerTaskBuffer <- task:
		p.metrics.recordSubmit()
		if p.signals != nil {
			p.signals.EmitRequestScheduled(context.Background(), req)
		}
		return nil
	default:
		p.metrics.recordDrop()
		if p.signals != nil {
			p.signals.EmitRequestDropped(context.Background(), req, ErrWorkerPoolFull)
		}
		task.req = nil
		task.taskHandle = nil
		p.workerTaskPool.Put(task)
		return ErrWorkerPoolFull
	}
}

func (p *workerPool) Results() <-chan engine.IResult {
	return p.results
}

func (p *workerPool) ReleaseResult(res engine.IResult) {
	if res == nil {
		return
	}

	// drain and close body before cancel to ensure connection reuse
	if resp, ok := res.Response().(*response); ok {
		if resp.body != nil {
			buf := discardBufPool.Get().(*[]byte)
			io.CopyBuffer(io.Discard, resp.body, *buf)
			discardBufPool.Put(buf)
			resp.body.Close()
		}
		resp.Reset()
		p.responsePool.Put(resp)
	}

	// now cancel the context (connection stays alive for reuse)
	res.Release()

	// return result to pool
	if r, ok := res.(*result); ok {
		r.Reset()
		p.resultPool.Put(r)
	}
}
