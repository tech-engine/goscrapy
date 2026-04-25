package worker

import (
	"context"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"github.com/tech-engine/goscrapy/pkg/core"
	"github.com/tech-engine/goscrapy/pkg/engine"
	"github.com/tech-engine/goscrapy/pkg/logger"
)

type Config struct {
	Executor   IExecutor
	Autoscaler struct {
		MaxWorkers    uint32
		MinWorkers    uint32
		ScalingFactor float32
		ScalingWindow time.Duration
		EMAAlpha      float32
	}
	Logger core.ILogger
}

type workerPool struct {
	executor         IExecutor
	results          chan engine.IResult
	workerTaskBuffer chan *workTask
	workerTaskPool   sync.Pool
	autoscaler       *autoscaler
	activeWorkers    atomic.Int32
	lastWorkerID     atomic.Uint32
	responsePool     sync.Pool
	logger           core.ILogger
	wg               sync.WaitGroup
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

	p := &workerPool{
		executor:         config.Executor,
		results:          make(chan engine.IResult, 1000),
		workerTaskBuffer: make(chan *workTask, 1000),
		logger:           config.Logger,
	}

	autoscalerConfig := &AutoscalerConfig{
		MaxWorkers:    config.Autoscaler.MaxWorkers,
		MinWorkers:    config.Autoscaler.MinWorkers,
		ScalingFactor: config.Autoscaler.ScalingFactor,
		ScalingWindow: config.Autoscaler.ScalingWindow,
		EMAAlpha:      config.Autoscaler.EMAAlpha,
		currentWorkerCntFn: func() int32 {
			return p.activeWorkers.Load()
		},
		spawnWorkerFn: func(ctx context.Context) {
			p.spawnWorker(ctx)
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

	w := NewWorker(id, p.executor, p.workerTaskBuffer, p.results, &p.responsePool, &p.workerTaskPool, p.autoscaler.OnTaskDone, p.autoscaler.ShouldExit)

	p.wg.Add(1)
	go func() {
		defer p.wg.Done()
		defer p.activeWorkers.Add(-1)
		_ = w.Start(ctx)
	}()
}

func (p *workerPool) Submit(req *core.Request, callbackName string, handle core.TaskHandle) error {
	p.autoscaler.OnTaskArrival()

	task := p.workerTaskPool.Get().(*workTask)
	task.req = req
	task.callbackName = callbackName
	task.taskHandle = handle

	select {
	case p.workerTaskBuffer <- task:
		return nil
	default:
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
			io.Copy(io.Discard, resp.body)
			resp.body.Close()
		}
		resp.Reset()
		p.responsePool.Put(resp)
	}

	// now cancel the context (connection stays alive for reuse)
	res.Release()
}
