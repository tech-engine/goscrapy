package corespider

import (
	"net/http"

	"github.com/tech-engine/goscrapy/pkg/core"
)

type CoreSpiderBuilder[OUT any] struct {
	*core.Core[OUT]
	Engine            IEngineConfigurer[OUT]
	PipelineManager   IPipelineManagerAdder[OUT]
	Scheduler         ISchedulerConfigurer[OUT]
	Executor          IExecutorConfigurer[OUT]
	ExecutorAdapter   IExecutorAdapterConfigurer
	MiddlewareManager IMiddlewareManager
	HttpClient        *http.Client
}
