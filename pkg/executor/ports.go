package executor

import (
	"net/http"

	"github.com/tech-engine/goscrapy/pkg/core"
)

type IExecutorAdapter interface {
	Do(core.IResponseWriter, *http.Request) error
	WithLogger(core.ILogger) IExecutorAdapter
}
