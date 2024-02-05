package engine

import (
	"context"
	"sync"

	"github.com/tech-engine/goscrapy/pkg/core"
)

type Engine[OUT any] struct {
	wg              sync.WaitGroup
	scheduler       IScheduler
	pipelineManager IPipelineManager[OUT]
	middlewares     []Middleware
	outputCh        chan core.IOutput[OUT]
}

func New[OUT any](schd IScheduler, pm IPipelineManager[OUT]) *Engine[OUT] {

	engine := &Engine[OUT]{
		scheduler:       schd,
		pipelineManager: pm,
		middlewares:     make([]Middleware, 0),
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

// func (m *Engine[OUT]) chainMiddlewares() http.RoundTripper {

// 	// add all the middlewares
// 	roundTripper := http.DefaultTransport
// 	for _, middleware := range m.middlewares {
// 		roundTripper = middleware(roundTripper)
// 	}

// 	return roundTripper

// }

// func (m *Engine[OUT]) AddMiddlewares(middlewares ...Middleware) {
// 	m.middlewares = append(m.middlewares, middlewares...)
// }

// func (m *Manager[OUT]) AddPipeline(p Pipeline[IN, any, OUT, Output[OUT]], err error) *pipeline[IN, any, OUT, Output[OUT]] {

// 	pipeline := &pipeline[IN, any, OUT, Output[OUT]]{
// 		p: p,
// 	}

// 	if err != nil {
// 		log.Panicf("Core:AddPipeline %s", err.Error())
// 	}

// 	m.pipelines.add(pipeline)
// 	return pipeline
// }
