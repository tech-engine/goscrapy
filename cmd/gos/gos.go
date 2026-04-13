package gos

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/tech-engine/goscrapy/internal/types"
	"github.com/tech-engine/goscrapy/pkg/core"
	"github.com/tech-engine/goscrapy/pkg/engine"
	"github.com/tech-engine/goscrapy/pkg/executor"
	httpnative "github.com/tech-engine/goscrapy/pkg/executor_adapters/http_native"
	"github.com/tech-engine/goscrapy/pkg/logger"
	"github.com/tech-engine/goscrapy/pkg/middlewaremanager"
	pipelinemanager "github.com/tech-engine/goscrapy/pkg/pipeline_manager"
	"github.com/tech-engine/goscrapy/pkg/scheduler"
	ts "github.com/tech-engine/goscrapy/pkg/telemetry/stats"
)

func New[OUT any]() *gosBuilder[OUT] {

	httpClient := &http.Client{}
	mm := middlewaremanager.New(httpClient)
	pm := pipelinemanager.New[OUT]()

	adapter := httpnative.NewHTTPClientAdapter(httpClient)
	executor := executor.New(adapter)
	scheduler := scheduler.New(executor)

	eng := engine.New(scheduler, pm)

	//propagate logger to all components (recursive)
	eng.WithLogger(logger.NewLogger())

	builder := &gosBuilder[OUT]{
		Engine:            eng,
		MiddlewareManager: mm,
		PipelineManager:   pm,
		Scheduler:         scheduler,
	}

	return builder
}

func (gos *gosBuilder[OUT]) Setup(
	middlewares []middlewaremanager.Middleware,
	pipelines []pipelinemanager.IPipeline[OUT],
	onShutdown ...func(),
) *gosBuilder[OUT] {
	gos.logger.Infof("Initializing engine with %d middlewares and %d pipelines", len(middlewares), len(pipelines))
	gos.MiddlewareManager.Add(middlewares...)
	gos.PipelineManager.Add(pipelines...)
	for _, fn := range onShutdown {
		gos.Engine.WithOnShutdown(fn)
	}
	return gos
}

func (gos *gosBuilder[OUT]) WithStatsRecorderFactory(factory ts.IStatsRecorderFactory) *gosBuilder[OUT] {
	if gos.Scheduler != nil && factory != nil {
		gos.Scheduler.WithStatsRecorderFactory(factory)
	}
	return gos
}

func (gos *gosBuilder[OUT]) WithLogger(loggerIn core.IConfigurableLogger) *gosBuilder[OUT] {
	loggerIn = logger.EnsureLogger(loggerIn).(core.IConfigurableLogger)
	gos.logger = loggerIn.WithName("GOS")
	gos.Engine.WithLogger(loggerIn)
	return gos
}

func (gos *gosBuilder[OUT]) Start(ctx context.Context) error {
	return gos.Engine.Start(ctx)
}

// Wait for completion or termination
func (gos *gosBuilder[OUT]) Wait(cancel context.CancelFunc, errCh <-chan error) error {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errCh:
		return err
	case sig := <-sigCh:
		gos.logger.Infof("Received termination signal: %v", sig)
		cancel()
		return <-errCh
	}
}

func NewBroadcaster(s ts.ISnapshotter, options ...types.OptFunc[ts.BroadcasterOpts]) *ts.Broadcaster {
	return ts.NewBroadcaster(s, options...)
}
