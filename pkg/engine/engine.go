package engine

import (
	"context"
	"errors"
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
		engine.resultHandlers = uint(max(runtime.NumCPU(), 4))
	}

	engine.logger.Debugf("Engine created at %p", engine)

	// wire up activity tracking to signals
	engine.signals.OnItemScraped(func(ctx context.Context, item OUT) { engine.Dec() })
	engine.signals.OnItemDropped(func(ctx context.Context, item OUT, err error) { engine.Dec() })
	engine.signals.OnItemError(func(ctx context.Context, item OUT, err error) { engine.Dec() })

	return engine, nil
}

func (e *Engine[OUT]) Start(ctx context.Context) error {
	if e.started.Swap(true) {
		return ErrAlreadyStarted
	}

	e.logger.Infof("Engine starting...")
	e.signals.EmitEngineStarted(ctx)

	// run all shutdown hooks before returning
	defer func() {
		e.logger.Infof("Shutting down engine...")
		e.signals.EmitSpiderClosed(ctx)
		e.signals.EmitEngineStopped(ctx)
		e.logger.Infof("shutdown complete.")
		e.started.Store(false)
	}()

	g, gCtx := errgroup.WithContext(ctx)

	g.Go(func() error { return e.pipelineManager.Start(gCtx) })
	g.Go(func() error { return e.scheduler.Start(gCtx) })
	g.Go(func() error { return e.workerPool.Start(gCtx) })

	// track activity during open to prevent premature idle
	e.activeCount.Add(1)
	e.signals.EmitSpiderOpened(gCtx)
	e.Dec()

	// result handler pool
	for range e.resultHandlers {
		g.Go(func() error {
			for res := range e.workerPool.Results() {
				e.handleResult(gCtx, res)
			}
			return nil
			// for {
			// 	select {
			// 	case <-gCtx.Done():
			// 		return nil
			// 	case res, ok := <-m.workerPool.Results():
			// 		if !ok {
			// 			return nil
			// 		}
			// 		m.handleResult(gCtx, res)
			// 	}
			// }
		})
	}

	// pull work from scheduler and submit to worker pool
	g.Go(func() error {
		for {
			select {
			case <-gCtx.Done(): // we immediately stop pulling new tasks
				return nil
			default:
				req, cbName, handle, err := e.scheduler.NextRequest(gCtx)

				if err != nil {
					// suppress warning if we are shutting down
					if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
						continue
					}
					e.logger.Warnf("failed to pull task: %v", err)
					continue
				}

				if req == nil {
					e.logger.Warn("got nil request from scheduler, retrying...")
					continue
				}

				e.activeCount.Add(1)
				if err := e.workerPool.Submit(gCtx, req, cbName, handle); err != nil {
					if !errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded) {
						e.logger.Errorf("failed to submit task: %v", err)
					}
					e.activeCount.Add(-1)
					e.scheduler.Nack(handle)
				}
			}
		}
	})

	err := g.Wait()

	// drain any pending leftover results
	// 	for {
	// 		select {
	// 		case res := <-m.workerPool.Results():
	// 			m.workerPool.ReleaseResult(res)
	// 		default:
	// 			goto drained
	// 		}
	// 	}
	// drained:

	// no need to be called from engine, pipeline manager calls its automatically now
	// stop pipelines after workers and scheduler finish
	// m.pipelineManager.Stop()

	return err
}

func (e *Engine[OUT]) handleResult(ctx context.Context, res IResult) {
	defer func() {
		e.activeCount.Add(-1)

		// notify idle if all tasks are done
		e.checkIdle(ctx)

		// acknowledge the task
		e.scheduler.Ack(res.TaskHandle())
		// release the result, cancels context and returns response to pool
		e.workerPool.ReleaseResult(res)
	}()

	if res.Error() != nil {
		// suppress expected shutdown errors
		if !errors.Is(res.Error(), context.Canceled) {
			e.logger.Errorf("request failed: %v", res.Error())
		}
		if e.signals != nil {
			e.signals.EmitSpiderError(ctx, res.Error())
		}
		return
	}

	cb, ok := e.callbackRegistry.Resolve(res.CallbackName())
	if !ok {
		e.logger.Errorf("callback not found: %s", res.CallbackName())
		return
	}

	// run the callback
	cb(ctx, res.Response())
}

func (e *Engine[OUT]) Schedule(req *core.Request, cb core.ResponseCallback) {
	if !e.scheduler.IsActive() {
		return
	}
	e.activeCount.Add(1)

	// we get or cache callback name
	ptr := reflect.ValueOf(cb).Pointer()
	name, ok := e.cbNameCache.Load(ptr)
	if !ok {
		resolved := runtime.FuncForPC(ptr).Name()
		name, _ = e.cbNameCache.LoadOrStore(ptr, resolved)
		e.callbackRegistry.Register(resolved, cb)
	}

	// schedule the request
	if err := e.scheduler.Schedule(req, name.(string)); err != nil {
		e.logger.Warnf("schedule rejected (further errors suppressed): %v", err)
		e.activeCount.Add(-1)
	}
}

// scans the spider for callback methods and registers them
func (e *Engine[OUT]) RegisterSpider(spider any) error {
	v := reflect.ValueOf(spider)
	t := v.Type()

	e.logger.Debugf("Discovering callbacks for spider: %T", spider)

	count := 0
	cbType := reflect.TypeOf((*core.ResponseCallback)(nil)).Elem()

	for i := 0; i < t.NumMethod(); i++ {
		method := v.Method(i)
		mType := method.Type()

		if mType.ConvertibleTo(cbType) {
			cb := method.Convert(cbType).Interface().(core.ResponseCallback)
			name := t.Method(i).Name
			e.callbackRegistry.Register(name, cb)
			e.logger.Debugf("  -> registered callback: %s", name)
			count++
		}
	}

	// auto discover spider signals
	if method, ok := t.MethodByName("Open"); ok {
		if fn, ok := v.Method(method.Index).Interface().(func(context.Context)); ok {
			e.signals.OnSpiderOpened(fn)
			e.logger.Debugf("  -> auto-discovered signal: Open")
		}
	}
	if method, ok := t.MethodByName("Idle"); ok {
		if fn, ok := v.Method(method.Index).Interface().(func(context.Context)); ok {
			e.signals.OnSpiderIdle(fn)
			e.logger.Debugf("  -> auto-discovered signal: Idle")
		}
	}
	if method, ok := t.MethodByName("Close"); ok {
		if fn, ok := v.Method(method.Index).Interface().(func(context.Context)); ok {
			e.signals.OnSpiderClosed(fn)
			e.logger.Debugf("  -> auto-discovered signal: Close")
		}
	}
	if method, ok := t.MethodByName("Error"); ok {
		if fn, ok := v.Method(method.Index).Interface().(func(context.Context, error)); ok {
			e.signals.OnSpiderError(fn)
			e.logger.Debugf("  -> auto-discovered signal: Error")
		}
	}

	if count == 0 {
		return ErrNoCallbacksFound
	}

	return nil
}

func (e *Engine[OUT]) Yield(v core.IOutput[OUT]) {
	e.activeCount.Add(1)
	e.pipelineManager.Push(v)
}

func (e *Engine[OUT]) Inc() {
	e.activeCount.Add(1)
}

func (e *Engine[OUT]) Dec() {
	e.activeCount.Add(-1)
	e.checkIdle(context.Background())
}

func (e *Engine[OUT]) checkIdle(ctx context.Context) {
	if e.activeCount.Load() == 0 && e.started.Load() {
		e.signals.EmitSpiderIdle(ctx)
	}
}

// ActiveCount returns current number of active tasks
func (e *Engine[OUT]) ActiveCount() int64 { return e.activeCount.Load() }

// IsStarted returns true if the engine has started
func (e *Engine[OUT]) IsStarted() bool { return e.started.Load() }

func (e *Engine[OUT]) WithLogger(loggerIn core.ILogger) core.IEngine[OUT] {
	loggerIn = logger.EnsureLogger(loggerIn)
	e.logger = loggerIn.WithName("Engine")
	return e
}
