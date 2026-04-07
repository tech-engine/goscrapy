package engine

import (
	"context"

	"github.com/tech-engine/goscrapy/pkg/core"
	"golang.org/x/sync/errgroup"
)

type Engine[OUT any] struct {
	scheduler       IScheduler
	pipelineManager IPipelineManager[OUT]
	outputCh        chan core.IOutput[OUT]
}

func New[OUT any](schd IScheduler, pm IPipelineManager[OUT]) *Engine[OUT] {

	engine := &Engine[OUT]{
		scheduler:       schd,
		pipelineManager: pm,
	}

	return engine
}

// start the core
func (m *Engine[OUT]) Start(ctx context.Context) error {

	g, gCtx := errgroup.WithContext(ctx)

	pmCtx, pmCancel := context.WithCancel(context.Background())

	g.Go(func() error {
		// Signals pipeline manager to shut down only after scheduler completes
		defer pmCancel()
		return m.scheduler.Start(gCtx)
	})

	g.Go(func() error {
		return m.pipelineManager.Start(pmCtx)
	})

	return g.Wait()
}

func (m *Engine[OUT]) Schedule(req core.IRequestReader, cb core.ResponseCallback) {
	m.scheduler.Schedule(req, cb)
}

func (m *Engine[OUT]) Yield(out core.IOutput[OUT]) {
	m.pipelineManager.Push(out)
}

func (m *Engine[OUT]) NewRequest() core.IRequestRW {
	return m.scheduler.NewRequest()
}

func (m *Engine[OUT]) WithScheduler(schd IScheduler) {
	m.scheduler = schd
}

func (m *Engine[OUT]) WithPipelineManager(pm IPipelineManager[OUT]) {
	m.pipelineManager = pm
}
