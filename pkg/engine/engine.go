package engine

import (
	"context"

	"errors"
	"sync/atomic"
	"time"

	"github.com/tech-engine/goscrapy/internal/types"
	"github.com/tech-engine/goscrapy/pkg/core"
	"github.com/tech-engine/goscrapy/pkg/logger"
	"golang.org/x/sync/errgroup"
)

var (
	ErrAlreadyStarted = errors.New("engine already started")
)

type opts struct {
	shutdownTimeout time.Duration
	onShutdown      []func()
}

func defaultOpts() opts {
	return opts{
		shutdownTimeout: 10 * time.Second,
	}
}

type Engine[OUT any] struct {
	opts
	scheduler       IScheduler
	pipelineManager IPipelineManager[OUT]
	logger          core.ILogger
	activeCount     atomic.Int64
	started         atomic.Bool
}

func New[OUT any](schd IScheduler, pm IPipelineManager[OUT], optFuncs ...types.OptFunc[opts]) *Engine[OUT] {

	opts := defaultOpts()

	for _, fn := range optFuncs {
		fn(&opts)
	}

	engine := &Engine[OUT]{
		opts:            opts,
		scheduler:       schd,
		pipelineManager: pm,
		logger:          logger.EnsureLogger(nil).WithName("Engine"),
	}

	return engine
}

func (m *Engine[OUT]) WithShutdownTimeout(timeout time.Duration) {
	m.opts.shutdownTimeout = timeout
}

// WithOnShutdown registers shutdown handlers to be executed on exit
func (m *Engine[OUT]) WithOnShutdown(funcs ...func()) {
	m.onShutdown = append(m.onShutdown, funcs...)
}

func (m *Engine[OUT]) WithScheduler(schd IScheduler) {
	m.scheduler = schd
}

func (m *Engine[OUT]) WithPipelineManager(pm IPipelineManager[OUT]) {
	m.pipelineManager = pm
}

func (m *Engine[OUT]) Start(ctx context.Context) error {
	if m.started.Swap(true) {
		return ErrAlreadyStarted
	}

	m.logger.Infof("Engine starting...")

	// Wire up activity tracking
	m.scheduler.WithActivityTracker(m)
	m.pipelineManager.WithActivityTracker(m)

	// run all shutdown hooks before returning
	defer func() {
		m.logger.Infof("Shutting down engine...")
		for _, fn := range m.onShutdown {
			fn()
		}
		m.logger.Infof("shutdown complete.")
		m.started.Store(false)
	}()

	g, gCtx := errgroup.WithContext(ctx)

	// pmCtx is used to stop the pipeline manager once scheduler has finished
	pmCtx, pmCancel := context.WithCancel(context.Background())

	g.Go(func() error {
		// Signals pipeline manager to shut down only after scheduler completes
		defer pmCancel()
		return m.scheduler.Start(gCtx)
	})

	g.Go(func() error {
		return m.pipelineManager.Start(pmCtx)
	})

	err := g.Wait()

	// after finishing scheduler and pipeline manager we wait for queued items to finish.
	// we use a timeout
	if m.opts.shutdownTimeout > 0 {
		_, cancel := context.WithTimeout(context.Background(), m.opts.shutdownTimeout)
		defer cancel()
	}

	return err
}

func (m *Engine[OUT]) Schedule(req *core.Request, next core.ResponseCallback) {
	m.activeCount.Add(1)
	m.scheduler.Schedule(req, next)
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
}

func (m *Engine[OUT]) ActiveCount() int64 {
	return m.activeCount.Load()
}

func (m *Engine[OUT]) WithLogger(loggerIn core.ILogger) *Engine[OUT] {
	loggerIn = logger.EnsureLogger(loggerIn)
	m.logger = loggerIn.WithName("Engine")
	m.scheduler.WithLogger(loggerIn)
	m.pipelineManager.WithLogger(loggerIn)

	return m
}
