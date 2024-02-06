package corespider

import (
	"net/http"

	"github.com/tech-engine/goscrapy/pkg/core"
)

type CoreSpiderBuilder[OUT any] struct {
	*core.Core[OUT]
	engine            IEngineConfigurer[OUT]
	pipelineManager   IPipelineManagerAdder[OUT]
	scheduler         ISchedulerConfigurer[OUT]
	executor          IExecutorConfigurer[OUT]
	executorAdapter   IExecutorAdapterConfigurer
	middlewareManager IMiddlewareManager
	httpClient        *http.Client
}
