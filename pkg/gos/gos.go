package gos

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/tech-engine/goscrapy/internal/request"
	"github.com/tech-engine/goscrapy/pkg/core"
	"github.com/tech-engine/goscrapy/pkg/engine"
	"github.com/tech-engine/goscrapy/pkg/executor"
	httpAdapter "github.com/tech-engine/goscrapy/pkg/executor_adapters/http_native"
	"github.com/tech-engine/goscrapy/pkg/logger"
	"github.com/tech-engine/goscrapy/pkg/middlewaremanager"
	pipelinemanager "github.com/tech-engine/goscrapy/pkg/pipeline_manager"
	"github.com/tech-engine/goscrapy/pkg/scheduler"
	ts "github.com/tech-engine/goscrapy/pkg/telemetry/stats"
	"github.com/tech-engine/goscrapy/pkg/worker"
)

type Config struct {
	Client *http.Client
	Logger core.ILogger
}

func DefaultConfig() *Config {
	return &Config{
		Client: DefaultClient(),
	}
}

type app[OUT any] struct {
	*core.Core[OUT]
	Engine            core.IEngine[OUT]
	PipelineManager   engine.IPipelineManager[OUT]
	Scheduler         engine.IScheduler
	WorkerPool        engine.IWorkerPool
	Executor          worker.IExecutor
	MiddlewareManager IMiddlewareManager
	httpClient        *http.Client
	logger            core.ILogger
	hub               *ts.TelemetryHub
	cancelableSignal  *cancelableSignal
	lastErr           error
}

func New[OUT any](config *Config) (*app[OUT], error) {
	if config == nil {
		config = DefaultConfig()
	}

	if config.Logger == nil {
		config.Logger = logger.NewLogger()
	}

	l := config.Logger

	// create our http adapter
	adapter, err := httpAdapter.NewAdapter(&httpAdapter.Config{
		Client: config.Client,
		Logger: l,
	})

	if err != nil {
		return nil, err
	}

	// create our executor
	exec, err := executor.New(&executor.Config{
		Adapter: adapter,
		Logger:  l.WithName("Executor"),
	})
	if err != nil {
		return nil, err
	}

	// read from env vars for tuning
	concurrency := uint32(16)
	if c := os.Getenv("SCHEDULER_CONCURRENCY"); c != "" {
		if v, err := strconv.Atoi(c); err == nil {
			concurrency = uint32(v)
		}
	}

	queueSize := uint64(1000)
	if q := os.Getenv("PIPELINEMANAGER_OUTPUT_QUEUE_BUF_SIZE"); q != "" {
		if v, err := strconv.ParseUint(q, 10, 64); err == nil && v > 0 {
			queueSize = v
		}
	}

	// create our scheduler
	sched, err := scheduler.New(&scheduler.Config{
		WorkQueueSize: queueSize,
		Logger:        l.WithName("Scheduler"),
	})
	if err != nil {
		return nil, err
	}

	// create our worker pool
	pool, err := worker.NewPool(&worker.Config{
		Executor: exec,
		Autoscaler: struct {
			MaxWorkers    uint32
			MinWorkers    uint32
			ScalingFactor float32
			ScalingWindow time.Duration
			EMAAlpha      float32
		}{
			MinWorkers: concurrency / 2,
			MaxWorkers: concurrency,
		},
		Logger: l.WithName("WorkerPool"),
	})

	if err != nil {
		return nil, err
	}

	// create our pipeline manager
	pmCfg := pipelinemanager.DefaultConfig()
	pmCfg.OutputQueueBuffSize = queueSize
	pmCfg.Logger = l.WithName("PipelineManager")
	pm := pipelinemanager.New[OUT](pmCfg)

	// create our engine
	eng, err := engine.New(&engine.Config[OUT]{
		Scheduler:       sched,
		WorkerPool:      pool,
		PipelineManager: pm,
	})
	if err != nil {
		return nil, err
	}
	eng.WithLogger(l.WithName("Engine"))

	app := &app[OUT]{
		Core:              core.New(eng, request.NewPool()),
		Engine:            eng,
		MiddlewareManager: middlewaremanager.New(config.Client),
		PipelineManager:   pm,
		Scheduler:         sched,
		WorkerPool:        pool,
		Executor:          exec,
		logger:            l.WithName("GOS"),
		cancelableSignal:  newCancelableSignal(context.Background()),
	}

	return app, nil
}

func (gos *app[OUT]) WithMiddlewares(middlewares ...middlewaremanager.Middleware) *app[OUT] {
	gos.MiddlewareManager.Add(middlewares...)
	return gos
}

func (gos *app[OUT]) WithPipelines(pipelines ...engine.IPipeline[OUT]) *app[OUT] {
	gos.PipelineManager.Add(pipelines...)
	return gos
}

func (gos *app[OUT]) WithOnEngineShutdown(onShutdown ...func()) *app[OUT] {
	for _, fn := range onShutdown {
		gos.Engine.WithOnShutdown(fn)
	}
	return gos
}

func (gos *app[OUT]) WithLogger(loggerIn core.IConfigurableLogger) *app[OUT] {
	loggerIn = logger.EnsureLogger(loggerIn).(core.IConfigurableLogger)
	gos.logger = loggerIn.WithName("GOS")
	gos.Engine.WithLogger(gos.logger.WithName("Engine"))
	return gos
}

func (gos *app[OUT]) RegisterSpider(spider any) {
	gos.Engine.RegisterSpider(spider)
}

func (gos *app[OUT]) WithTelemetry(hub *ts.TelemetryHub, config *ts.TelemetryHubConfig) *app[OUT] {
	if hub == nil {
		gos.hub = ts.NewTelemetryHub(config)
		return gos
	}
	gos.hub = hub
	return gos
}

func (gos *app[OUT]) Logger() core.ILogger {
	return gos.logger
}

func (gos *app[OUT]) Start(ctx context.Context) error {
	// Bridge the external context to the internal lifecycle
	// We do this instead of creating a child context to prevent
	// a race condition between Start and Wait.
	stop := context.AfterFunc(ctx, func() {
		gos.cancelableSignal.cancel()
	})
	defer stop()

	if gos.hub != nil {
		// auto register workerpool components as collectors
		if coll, ok := gos.WorkerPool.(ts.IStatsCollector); ok {
			gos.hub.AddCollector(coll)
		}
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
		// start auto exit monitor
		go func() {
			// wait for initial start
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
