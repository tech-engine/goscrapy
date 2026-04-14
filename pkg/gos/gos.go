package gos

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/tech-engine/goscrapy/internal/types"
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
	httpClient := DefaultClient()

	adapter := httpAdapter.NewAdapter(
		httpAdapter.WithClient(httpClient),
	)
	executor := executor.New(adapter)
	scheduler := scheduler.New(executor)
	mm := middlewaremanager.New(httpClient)
	pm := pipelinemanager.New[OUT]()

	l := logger.NewLogger()
	eng := engine.New(scheduler, pm).
		WithLogger(l)

	app := &app[OUT]{
		Core:              core.New(eng),
		Engine:            eng,
		MiddlewareManager: mm,
		PipelineManager:   pm,
		Scheduler:         scheduler,
		Executor:          executor,
		logger:            l.WithName("GOS"),
		hub:               nil,
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

func (gos *app[OUT]) WithTelemetry(hub *ts.TelemetryHub, optFuncs ...types.OptFunc[ts.TelemetryHubOpts]) *app[OUT] {
	if hub == nil {
		gos.hub = ts.NewTelemetryHub(optFuncs...)
		return gos
	}
	gos.hub = hub
	return gos
}

func (gos *app[OUT]) Logger() core.ILogger {
	return gos.logger
}

func (gos *app[OUT]) Start(ctx context.Context) error {
	// Create internal cancellable context based on provided context
	// This allows the framework to signal its own shutdown
	gos.cancelableSignal = newCancelableSignal(ctx)
	defer gos.cancelableSignal.cancel()

	if gos.hub != nil {
		go gos.hub.Start(gos.cancelableSignal.ctx)
	}
	gos.lastErr = gos.Engine.Start(gos.cancelableSignal.ctx)
	return gos.lastErr
}

// Wait for completion or termination. If autoExit is true, the engine will
// shut down automatically when all work is finished.
func (gos *app[OUT]) Wait(autoExit ...bool) error {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	if len(autoExit) > 0 && autoExit[0] {
		// Start auto-exit monitor
		go func() {
			// Wait for initial start
			time.Sleep(500 * time.Millisecond)
			ticker := time.NewTicker(200 * time.Millisecond)
			defer func() {
				if gos.cancelableSignal != nil {
					gos.cancelableSignal.cancel()
				}
				ticker.Stop()
			}()

			for range ticker.C {
				if gos.Engine.ActiveCount() <= 0 {
					gos.logger.Info("✅ Scraping complete. Automatic shutdown initiated.")
					return
				}
			}
		}()
	} else {
		gos.logger.Info("🕷️  GoScrapy spider is running. Press Ctrl+C to stop.")
	}

	select {
	case <-gos.cancelableSignal.ctx.Done():
		return gos.lastErr
	case sig := <-sigCh:
		gos.logger.Infof("Received termination signal: %v", sig)
		if gos.cancelableSignal != nil {
			gos.cancelableSignal.cancel()
		}
		// Wait for engine to finish cleanup and set lastErr
		<-gos.cancelableSignal.ctx.Done()
		return gos.lastErr
	}
}
