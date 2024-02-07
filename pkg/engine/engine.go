package engine

import (
	"context"
	"sync"

	"github.com/tech-engine/goscrapy/pkg/core"
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

	var (
		wg    sync.WaitGroup
		errCh = make(chan error, 2)
	)

	wg.Add(2)

	pmCtx, pmCancel := context.WithCancel(context.Background())

	go func() {

		defer wg.Done()
		defer pmCancel()

		errCh <- m.scheduler.Start(ctx)

	}()

	go func() {

		defer wg.Done()

		errCh <- m.pipelineManager.Start(pmCtx)

	}()

	wg.Wait()

	close(errCh)

	for err := range errCh {
		if err != nil {
			return err
		}
	}
	return nil
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
