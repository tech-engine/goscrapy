package executor

import (
	"net/http"

	"github.com/tech-engine/goscrapy/pkg/core"
	"github.com/tech-engine/goscrapy/pkg/engine"
)

type IExecutorAdapter interface {
	Do(engine.IResponseWriter, *http.Request) error
	WithClient(*http.Client)
	WithLogger(core.ILogger)
}
