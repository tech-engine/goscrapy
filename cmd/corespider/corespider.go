package corespider

import (
	"context"
	"log"
	"net/http"

	"github.com/tech-engine/goscrapy/pkg/core"
	"github.com/tech-engine/goscrapy/pkg/engine"
	"github.com/tech-engine/goscrapy/pkg/executor"
	httpnative "github.com/tech-engine/goscrapy/pkg/executor_adapters/http_native"
	pipelinemanager "github.com/tech-engine/goscrapy/pkg/pipeline_manager"
	"github.com/tech-engine/goscrapy/pkg/scheduler"
)

func New[OUT any]() *CoreSpiderBuilder[OUT] {

	c := &CoreSpiderBuilder[OUT]{}

	c.httpClient = &http.Client{}

	c.executorAdapter = httpnative.NewHTTPClientAdapter(c.httpClient)

	c.executor = executor.New(c.executorAdapter)

	c.scheduler = scheduler.New(c.executor)

	c.pipelineManager = pipelinemanager.New[OUT]()

	c.engine = engine.New(c.scheduler, c.pipelineManager)

	c.Core = core.New[OUT](c.engine)

	return c
}

func (c *CoreSpiderBuilder[OUT]) Engine() IEngineConfigurer[OUT] {
	return c.engine
}

func (c *CoreSpiderBuilder[OUT]) WithEngine(engine IEngineConfigurer[OUT]) *CoreSpiderBuilder[OUT] {

	if engine == nil {
		log.Fatal("corespider.go:WithEngine(engine): engine cannot be nil")
	}

	c.engine = engine

	return c
}

func (c *CoreSpiderBuilder[OUT]) Scheduler() ISchedulerConfigurer[OUT] {
	return c.scheduler
}

func (c *CoreSpiderBuilder[OUT]) WithScheduler(scheduler ISchedulerConfigurer[OUT]) *CoreSpiderBuilder[OUT] {
	if scheduler == nil {
		log.Fatal("corespider.go:WithScheduler(scheduler): scheduler cannot be nil")
	}

	c.scheduler = scheduler

	return c
}

// Returns the current pipelinemanager.IPipelineManagerAdder[OUT]
func (c *CoreSpiderBuilder[OUT]) PipelineManager() IPipelineManagerAdder[OUT] {
	return c.pipelineManager
}

// Let's us set our own custom pipelinemanager.IPipelineManagerAdder[OUT]
func (c *CoreSpiderBuilder[OUT]) WithPipelineManager(p IPipelineManagerAdder[OUT]) *CoreSpiderBuilder[OUT] {
	c.pipelineManager = p
	return c
}

func (c *CoreSpiderBuilder[OUT]) Executer() IExecutorConfigurer[OUT] {
	return c.executor
}

func (c *CoreSpiderBuilder[OUT]) WithExecuter(executor IExecutorConfigurer[OUT]) *CoreSpiderBuilder[OUT] {
	if executor == nil {
		log.Fatal("corespider.go:WithExecuter(executor): executor cannot be nil")
	}

	c.executor = executor

	return c
}

func (c *CoreSpiderBuilder[OUT]) ExecutorAdapter() IExecutorAdapterConfigurer {
	return c.executorAdapter
}

func (c *CoreSpiderBuilder[OUT]) WithExecutorAdapter(adapter IExecutorAdapterConfigurer) *CoreSpiderBuilder[OUT] {

	if adapter == nil {
		log.Fatal("corespider.go:WithExecutorAdapter(adapter): adapter cannot be nil")
	}

	c.executorAdapter = adapter

	return c
}

func (c *CoreSpiderBuilder[OUT]) Start(ctx context.Context) error {
	return c.engine.Start(ctx)
}
