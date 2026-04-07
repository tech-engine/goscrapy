package engine

import (
	"context"

	"time"

	"github.com/tech-engine/goscrapy/internal/types"
	"github.com/tech-engine/goscrapy/pkg/core"
	"golang.org/x/sync/errgroup"
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
	}

	return engine
}

// WithOnShutdown registers shutdown handlers to be executed on exit
func WithOnShutdown(funcs ...func()) types.OptFunc[opts] {
	return func(o *opts) {
		o.onShutdown = append(o.onShutdown, funcs...)
	}
}

// start the core
func (m *Engine[OUT]) Start(ctx context.Context) error {

	// run all shutdown hooks before returning
	defer func() {
		for _, fn := range m.opts.onShutdown {
			fn()
		}
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
	// we use a timeout to prevent hanging forever.
	if m.opts.shutdownTimeout > 0 {
		_, cancel := context.WithTimeout(context.Background(), m.opts.shutdownTimeout)
		defer cancel()
	}

	return err
}

func (m *Engine[OUT]) Schedule(req core.IRequestReader, cb core.ResponseCallback) {
	m.scheduler.Schedule(req, cb)
}

func (m *Engine[OUT]) Yield(out core.IOutput[OUT]) {
	m.pipelineManager.Push(out)
}

func (m *Engine[OUT]) NewRequest(ctx context.Context) core.IRequestRW {
	return m.scheduler.NewRequest(ctx)
}

func (m *Engine[OUT]) WithScheduler(schd IScheduler) {
	m.scheduler = schd
}

func (m *Engine[OUT]) WithPipelineManager(pm IPipelineManager[OUT]) {
	m.pipelineManager = pm
}

func (m *Engine[OUT]) WithShutdownTimeout(timeout time.Duration) {
	m.opts.shutdownTimeout = timeout
}

func (m *Engine[OUT]) WithOnShutdown(funcs ...func()) {
	m.opts.onShutdown = append(m.opts.onShutdown, funcs...)
}
