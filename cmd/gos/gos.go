package gos

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tech-engine/goscrapy/pkg/core"
	"github.com/tech-engine/goscrapy/pkg/engine"
	"github.com/tech-engine/goscrapy/pkg/executor"
	httpAdapter "github.com/tech-engine/goscrapy/pkg/executor_adapters/http_native"
	"github.com/tech-engine/goscrapy/pkg/logger"
	"github.com/tech-engine/goscrapy/pkg/middlewaremanager"
	pipelinemanager "github.com/tech-engine/goscrapy/pkg/pipeline_manager"
	"github.com/tech-engine/goscrapy/pkg/scheduler"
	ts "github.com/tech-engine/goscrapy/pkg/telemetry/stats"

)

func NewApp[OUT any]() *app[OUT] {

	httpClient := &http.Client{
		Timeout:   30 * time.Second,
		Transport: http.DefaultTransport.(*http.Transport).Clone(),
	}

	adapter := httpAdapter.NewAdapter(
		httpAdapter.WithClient(httpClient),
	)
	executor := executor.New(adapter)
	scheduler := scheduler.New(executor)
	mm := middlewaremanager.New(httpClient)
	pm := pipelinemanager.New[OUT]()

	eng := engine.New(scheduler, pm).
		WithLogger(logger.NewLogger())

	app := &app[OUT]{
		Core:              core.New(eng),
		Engine:            eng,
		MiddlewareManager: mm,
		PipelineManager:   pm,
		Scheduler:         scheduler,
	}

	return app
}

func (gos *app[OUT]) Setup(
	middlewares []middlewaremanager.Middleware,
	pipelines []pipelinemanager.IPipeline[OUT],
	onShutdown ...func(),
) *app[OUT] {
	gos.logger.Infof("Initializing engine with %d middlewares and %d pipelines", len(middlewares), len(pipelines))
	gos.MiddlewareManager.Add(middlewares...)
	gos.PipelineManager.Add(pipelines...)
	for _, fn := range onShutdown {
		gos.Engine.WithOnShutdown(fn)
	}
	return gos
}

func (gos *app[OUT]) WithStatsRecorderFactory(factory ts.IStatsRecorderFactory) *app[OUT] {
	if gos.Scheduler != nil && factory != nil {
		gos.Scheduler.WithStatsRecorderFactory(factory)
	}
	return gos
}

func (gos *app[OUT]) WithLogger(loggerIn core.IConfigurableLogger) *app[OUT] {
	loggerIn = logger.EnsureLogger(loggerIn).(core.IConfigurableLogger)
	gos.logger = loggerIn.WithName("GOS")
	gos.Engine.WithLogger(loggerIn)
	return gos
}

func (gos *app[OUT]) Telemetry() *ts.TelemetryHub {
	if gos.hub == nil {
		gos.hub = ts.NewTelemetryHub()
	}
	return gos.hub
}

func (gos *app[OUT]) Logger() core.ILogger {
	return gos.logger
}



func (gos *app[OUT]) Start(ctx context.Context) error {
	if gos.hub != nil {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()
		go gos.hub.Start(ctx)
	}
	return gos.Engine.Start(ctx)
}

// Wait for completion or termination
func (gos *app[OUT]) Wait(cancel context.CancelFunc, errCh <-chan error) error {
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
