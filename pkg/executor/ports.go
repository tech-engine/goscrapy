package executor

import (
	"net/http"

	"github.com/tech-engine/goscrapy/pkg/engine"
)

type IExecutorAdapter interface {
	Do(engine.IResponseWriter, *http.Request) error
	Acquire() *http.Request
	WithClient(*http.Client)
}
