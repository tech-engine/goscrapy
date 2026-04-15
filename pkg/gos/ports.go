// GOS is the app build using the goscrapy components.
// This where the components are stitch together to form a complete GOS application
// This is the package that majority is not all spider will be using.
// Unless you may want to tweak the individual components.
package gos

import (
	"context"
	"net/http"
	"time"

	"github.com/tech-engine/goscrapy/pkg/core"
	"github.com/tech-engine/goscrapy/pkg/engine"
	"github.com/tech-engine/goscrapy/pkg/executor"
	"github.com/tech-engine/goscrapy/pkg/middlewaremanager"
	pipelinemanager "github.com/tech-engine/goscrapy/pkg/pipeline_manager"
	"github.com/tech-engine/goscrapy/pkg/scheduler"
	ts "github.com/tech-engine/goscrapy/pkg/telemetry/stats"
)

// Any custom spider created using GoScrapy Framework must implement ICoreSpider[OUT any] interface
type ICoreSpider[OUT any] interface {
	// Parse schedules a request and passed the processed response to the callback
	Parse(req core.IRequestReader, cb core.ResponseCallback)
	// Request creates a new request
	Request(context.Context) core.IRequestRW
	Yield(core.IOutput[OUT])
	Logger() core.ILogger
	// Wait blocks until the spider finishes its work or receives a termination signal (Ctrl+C).
	// If the optional 'autoExit' parameter is set to true, the framework will
	// automatically initiate a graceful shutdown once it detects that all
	// scraping tasks and pipeline operations are complete.
	Wait(...bool) error
}

// Separate interface created for configuration purposes

// engine.*Engine[OUT] accepts a pipeline manager that implements engine.IPipelineManager[OUT]
// interface which doesn't have the Add function as engine.IPipelineManager[OUT]
// is not responsible for adding pipelines.
// But pipelinemanager.*PipelineManager[OUT] does exposes an Add function for external configuration
// purpose and to access it we have created the IPipelineManagerAdder[OUT] interface.
type IPipelineManagerAdder[OUT any] interface {
	engine.IPipelineManager[OUT]
	Add(...pipelinemanager.IPipeline[OUT])
}

// core.*Core[OUT] accepts an engine that implements core.IEngine[OUT] interface which
// doesn't have the WithScheduler function as core.IEngine[OUT] is not responsible for
// setting Scheduler. But engine.*Engine[OUT] does exposes a WithScheduler function for external
// configuration purposes and to access it we have created the IEngineConfigurer[OUT] interface.
// Same is the case for WithPipelineManager function.
type IEngineConfigurer[OUT any] interface {
	core.IEngine[OUT]
	WithScheduler(engine.IScheduler)
	WithPipelineManager(engine.IPipelineManager[OUT])
	WithOnShutdown(...func())
	WithShutdownTimeout(time.Duration)
	WithLogger(core.ILogger) *engine.Engine[OUT]
	ActiveCount() int64
}

// engine.*Engine[OUT] accepts a scheduler that implements engine.IScheduler[OUT] interface which
// doesn't have the WithExecutor function as engine.IScheduler[OUT] is not responsible for
// setting an Executor. But engine.IScheduler does exposes a WithExecutor function for external
// configuration purposes and to access it we have created the ISchedulerConfigurer[OUT] interface.
type ISchedulerConfigurer[OUT any] interface {
	engine.IScheduler
	WithExecutor(scheduler.IExecutor)
	WithStatsRecorderFactory(ts.IStatsRecorderFactory)
}

// scheduler.*Scheduler accepts a executor that implements scheduler.IExecutor interface which
// doesn't have the WithAdapter function as scheduler.IExecutor is not responsible for
// setting an adapter. But scheduler.*Scheduler does exposes a WithAdapter function for external
// configuration purposes and to access it we have created the IExecutorConfigurer[OUT] interface.
type IExecutorConfigurer[OUT any] interface {
	scheduler.IExecutor
	WithAdapter(executor.IExecutorAdapter)
}

// executor.*Executor accepts a adapter that implements executor.IExecutorAdapter interface which
// doesn't have the WithClient function as executor.IExecutorAdapter is not responsible for
// setting a http client. But executoradapter.*HTTPAdapter does exposes a WithClient function for external
// configuration purposes and to access it we have created the IExecutorAdapterConfigurer[OUT] interface.
type IExecutorAdapterConfigurer interface {
	executor.IExecutorAdapter
	WithClient(*http.Client)
}

type IMiddlewareManager interface {
	HTTPClient() *http.Client
	Add(...middlewaremanager.Middleware)
}
