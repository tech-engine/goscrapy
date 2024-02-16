package corespider

import (
	"context"
	"net/http"

	"github.com/tech-engine/goscrapy/pkg/core"
	"github.com/tech-engine/goscrapy/pkg/engine"
	"github.com/tech-engine/goscrapy/pkg/executor"
	httpnative "github.com/tech-engine/goscrapy/pkg/executor_adapters/http_native"
	"github.com/tech-engine/goscrapy/pkg/middlewaremanager"
	pipelinemanager "github.com/tech-engine/goscrapy/pkg/pipeline_manager"
	"github.com/tech-engine/goscrapy/pkg/scheduler"
)

func New[OUT any]() *CoreSpiderBuilder[OUT] {

	c := &CoreSpiderBuilder[OUT]{}

	c.HttpClient = &http.Client{}

	c.MiddlewareManager = middlewaremanager.New(c.HttpClient)

	c.ExecutorAdapter = httpnative.NewHTTPClientAdapter(
		c.MiddlewareManager.HTTPClient(),
	)

	c.Executor = executor.New(c.ExecutorAdapter)

	c.Scheduler = scheduler.New(c.Executor)

	c.PipelineManager = pipelinemanager.New[OUT]()

	c.Engine = engine.New(c.Scheduler, c.PipelineManager)

	c.Core = core.New[OUT](c.Engine)

	return c
}

func (c *CoreSpiderBuilder[OUT]) Start(ctx context.Context) error {
	return c.Engine.Start(ctx)
}
