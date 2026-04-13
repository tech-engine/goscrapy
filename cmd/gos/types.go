package gos

import (
	"net/http"

	"github.com/tech-engine/goscrapy/pkg/core"
	ts "github.com/tech-engine/goscrapy/pkg/telemetry/stats"
)

type app[OUT any] struct {
	*core.Core[OUT]
	Engine            IEngineConfigurer[OUT]
	PipelineManager   IPipelineManagerAdder[OUT]
	Scheduler         ISchedulerConfigurer[OUT]
	Executor          IExecutorConfigurer[OUT]
	ExecutorAdapter   IExecutorAdapterConfigurer
	MiddlewareManager IMiddlewareManager
	httpClient        *http.Client
	logger            core.ILogger
	hub               *ts.TelemetryHub
}
