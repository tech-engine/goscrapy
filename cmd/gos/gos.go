package gos

import (
	"context"
	"net/http"

	"errors"
	"os"
	"os/signal"
	"syscall"

	"github.com/tech-engine/goscrapy/pkg/core"
	"github.com/tech-engine/goscrapy/pkg/engine"
	"github.com/tech-engine/goscrapy/pkg/executor"
	httpnative "github.com/tech-engine/goscrapy/pkg/executor_adapters/http_native"
	"github.com/tech-engine/goscrapy/pkg/middlewaremanager"
	pipelinemanager "github.com/tech-engine/goscrapy/pkg/pipeline_manager"
	"github.com/tech-engine/goscrapy/pkg/scheduler"
)

func New[OUT any]() *gosBuilder[OUT] {
	c := &gosBuilder[OUT]{
		httpClient: DefaultClient(),
	}

	c.MiddlewareManager = middlewaremanager.New(c.httpClient)

	c.ExecutorAdapter = httpnative.NewHTTPClientAdapter(c.MiddlewareManager.HTTPClient(), 0)

	c.Executor = executor.New(c.ExecutorAdapter)

	c.Scheduler = scheduler.New(c.Executor)

	c.PipelineManager = pipelinemanager.New[OUT]()

	c.Engine = engine.New(c.Scheduler, c.PipelineManager)

	c.Core = core.New(c.Engine)
	return c
}

func (c *gosBuilder[OUT]) WithClient(cli *http.Client) *gosBuilder[OUT] {
	c.httpClient = cli
	return c
}

func (c *gosBuilder[OUT]) Start(ctx context.Context) error {
	return c.Engine.Start(ctx)
}

func (c *gosBuilder[OUT]) Setup(
	middlewares []middlewaremanager.Middleware,
	pipelines []pipelinemanager.IPipeline[OUT],
	onShutdown ...func(),
) *gosBuilder[OUT] {
	c.MiddlewareManager.Add(middlewares...)
	c.PipelineManager.Add(pipelines...)
	for _, fn := range onShutdown {
		c.Engine.WithOnShutdown(fn)
	}
	return c
}

// Wait for completion or termination
func Wait(cancel context.CancelFunc, errCh <-chan error) error {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		return err
	case <-sigCh:
		cancel()
		return <-errCh
	}
}
