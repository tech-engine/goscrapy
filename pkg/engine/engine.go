package engine

import (
	"context"
	"os"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"

	"reflect"
	"time"

	"github.com/tech-engine/goscrapy/pkg/core"
	"github.com/tech-engine/goscrapy/pkg/logger"
	"github.com/tech-engine/goscrapy/pkg/signal"

	"golang.org/x/sync/errgroup"
)

type Config[OUT any] struct {
	Scheduler        IScheduler
	WorkerPool       IWorkerPool
	PipelineManager  IPipelineManager[OUT]
	CallbackRegistry ICallbackRegistry
	Logger           core.ILogger
	Signals          *signal.Bus[OUT]
	ResultHandlers   uint
	shutdownTimeout  time.Duration
}

type Engine[OUT any] struct {
	scheduler        IScheduler
	workerPool       IWorkerPool
	pipelineManager  IPipelineManager[OUT]
	callbackRegistry ICallbackRegistry
	shutdownTimeout  time.Duration
	resultHandlers   uint
	logger           core.ILogger
	activeCount      atomic.Int64
	started          atomic.Bool
	signals          *signal.Bus[OUT]
	cbNameCache      sync.Map
}

func New[OUT any](config *Config[OUT]) (*Engine[OUT], error) {
	if config == nil {
		config = &Config[OUT]{}
	}

	if config.Scheduler == nil {
		return nil, ErrSchedulerMissing
	}

	if config.PipelineManager == nil {
		return nil, ErrPipelineManagerMissing
	}

	if config.WorkerPool == nil {
		return nil, ErrWorkerPoolMissing
	}

	if config.CallbackRegistry == nil {
		config.CallbackRegistry = NewCallbackRegistry()
	}

	if config.shutdownTimeout == 0 {
		config.shutdownTimeout = 10 * time.Second
	}

	if config.Logger == nil {
		config.Logger = logger.EnsureLogger(config.Logger).WithName("Engine")
	}

	if config.Signals == nil {
		config.Signals = signal.New[OUT]()
	}

	engine := &Engine[OUT]{
		scheduler:        config.Scheduler,
		workerPool:       config.WorkerPool,
		pipelineManager:  config.PipelineManager,
		callbackRegistry: config.CallbackRegistry,
		shutdownTimeout:  config.shutdownTimeout,
		resultHandlers:   config.ResultHandlers,
		logger:           config.Logger,
		signals:          config.Signals,
	}

	if engine.resultHandlers == 0 {
		if v := os.Getenv("ENGINE_RESULT_HANDLERS"); v != "" {
			if i, err := strconv.ParseUint(v, 10, 32); err == nil && i > 0 {
				engine.resultHandlers = uint(i)
			}
		}
	}

	if engine.resultHandlers == 0 {
		engine.resultHandlers = uint(runtime.NumCPU())
		if engine.resultHandlers < 4 {
			engine.resultHandlers = 4
		}
	}

	engine.logger.Debugf("Engine created at %p", engine)

	// wire up activity tracking to signals
	engine.signals.OnItemScraped(func(ctx context.Context, item OUT) { engine.Dec() })
	engine.signals.OnItemDropped(func(ctx context.Context, item OUT, err error) { engine.Dec() })
	engine.signals.OnItemError(func(ctx context.Context, item OUT, err error) { engine.Dec() })

	return engine, nil
}

func (m *Engine[OUT]) Start(ctx context.Context) error {
	if m.started.Swap(true) {
		return ErrAlreadyStarted
	}

	m.logger.Infof("Engine starting...")
	m.signals.EmitEngineStarted(ctx)

	// run all shutdown hooks before returning
	defer func() {
		m.logger.Infof("Shutting down engine...")
		m.signals.EmitSpiderClosed(ctx)
		m.signals.EmitEngineStopped(ctx)
		m.logger.Infof("shutdown complete.")
		m.started.Store(false)
	}()

	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error { return m.pipelineManager.Start(gCtx) })
	g.Go(func() error { return m.scheduler.Start(gCtx) })
	g.Go(func() error { return m.workerPool.Start(gCtx) })

	// track activity during open to prevent premature idle
	m.activeCount.Add(1)
	m.signals.EmitSpiderOpened(gCtx)
	m.Dec()

	// result handler pool
	for i := uint(0); i < m.resultHandlers; i++ {
		g.Go(func() error {
			for {
				select {
				case <-gCtx.Done():
					return nil
				case res, ok := <-m.workerPool.Results():
					if !ok {
						return nil
					}
					m.handleResult(gCtx, res)
				}
			}
		})
	}

	// pull work from scheduler and submit to worker pool
	g.Go(func() error {
		for {
			req, cbName, handle, err := m.scheduler.NextRequest(gCtx)
			if err != nil {
				return nil // either context cancelled or scheduler closed
			}

			if req == nil {
				continue
			}

			if err := m.workerPool.Submit(req, cbName, handle); err != nil {
				m.logger.Errorf("failed to submit task: %v", err)
				m.activeCount.Add(-1)
				m.scheduler.Nack(handle)
			}
		}
	})

	err := g.Wait()

	// stop pipelines after workers and scheduler finish
	m.pipelineManager.Stop()

	return err
}

func (m *Engine[OUT]) handleResult(ctx context.Context, res IResult) {
	defer func() {
		m.activeCount.Add(-1)

		// notify idle if all tasks are done
		m.checkIdle(ctx)

		// acknowledge the task
		m.scheduler.Ack(res.TaskHandle())
		// release the result, cancels context and returns response to pool
		m.workerPool.ReleaseResult(res)
	}()

	if res.Error() != nil {
		m.logger.Errorf("request failed: %v", res.Error())
		if m.signals != nil {
			m.signals.EmitSpiderError(ctx, res.Error())
		}
		return
	}

	cb, ok := m.callbackRegistry.Resolve(res.CallbackName())
	if !ok {
		m.logger.Errorf("callback not found: %s", res.CallbackName())
		return
	}

	// run the callback
	cb(ctx, res.Response())
}

func (m *Engine[OUT]) Schedule(req *core.Request, cb core.ResponseCallback) {
	m.activeCount.Add(1)

	// we get or cache callback name
	ptr := reflect.ValueOf(cb).Pointer()
	name, ok := m.cbNameCache.Load(ptr)
	if !ok {
		resolved := runtime.FuncForPC(ptr).Name()
		name, _ = m.cbNameCache.LoadOrStore(ptr, resolved)
		m.callbackRegistry.Register(resolved, cb)
	}

	// schedule the request
	if err := m.scheduler.Schedule(req, name.(string)); err != nil {
		m.logger.Errorf("failed to schedule request: %v", err)
		m.activeCount.Add(-1)
	}
}

// scans the spider for callback methods and registers them
func (m *Engine[OUT]) RegisterSpider(spider any) error {
	v := reflect.ValueOf(spider)
	t := v.Type()

	m.logger.Debugf("Discovering callbacks for spider: %T", spider)

	count := 0
	cbType := reflect.TypeOf((*core.ResponseCallback)(nil)).Elem()

	for i := 0; i < t.NumMethod(); i++ {
		method := v.Method(i)
		mType := method.Type()

		if mType.ConvertibleTo(cbType) {
			cb := method.Convert(cbType).Interface().(core.ResponseCallback)
			name := t.Method(i).Name
			m.callbackRegistry.Register(name, cb)
			m.logger.Debugf("  -> registered callback: %s", name)
			count++
		}
	}

	// auto discover spider signals
	if method, ok := t.MethodByName("Open"); ok {
		if fn, ok := v.Method(method.Index).Interface().(func(context.Context)); ok {
			m.signals.OnSpiderOpened(fn)
			m.logger.Debugf("  -> auto-discovered signal: Open")
		}
	}
	if method, ok := t.MethodByName("Idle"); ok {
		if fn, ok := v.Method(method.Index).Interface().(func(context.Context)); ok {
			m.signals.OnSpiderIdle(fn)
			m.logger.Debugf("  -> auto-discovered signal: Idle")
		}
	}
	if method, ok := t.MethodByName("Close"); ok {
		if fn, ok := v.Method(method.Index).Interface().(func(context.Context)); ok {
			m.signals.OnSpiderClosed(fn)
			m.logger.Debugf("  -> auto-discovered signal: Close")
		}
	}
	if method, ok := t.MethodByName("Error"); ok {
		if fn, ok := v.Method(method.Index).Interface().(func(context.Context, error)); ok {
			m.signals.OnSpiderError(fn)
			m.logger.Debugf("  -> auto-discovered signal: Error")
		}
	}

	if count == 0 {
		return ErrNoCallbacksFound
	}

	return nil
}

func (m *Engine[OUT]) Yield(v core.IOutput[OUT]) {
	m.activeCount.Add(1)
	m.pipelineManager.Push(v)
}

func (m *Engine[OUT]) Inc() {
	m.activeCount.Add(1)
}

func (m *Engine[OUT]) Dec() {
	m.activeCount.Add(-1)
	m.checkIdle(context.Background())
}

func (m *Engine[OUT]) checkIdle(ctx context.Context) {
	if m.activeCount.Load() == 0 && m.started.Load() {
		m.signals.EmitSpiderIdle(ctx)
	}
}

// ActiveCount returns current number of active tasks
func (m *Engine[OUT]) ActiveCount() int64 { return m.activeCount.Load() }

// IsStarted returns true if the engine has started
func (m *Engine[OUT]) IsStarted() bool { return m.started.Load() }

func (m *Engine[OUT]) WithLogger(loggerIn core.ILogger) core.IEngine[OUT] {
	loggerIn = logger.EnsureLogger(loggerIn)
	m.logger = loggerIn.WithName("Engine")
	return m
}
