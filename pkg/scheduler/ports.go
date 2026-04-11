package scheduler

import (
	"github.com/tech-engine/goscrapy/pkg/core"
)

// An executor must implement the IExecutor interface to be used by the scheduler.*Scheduler
type IExecutor interface {
	Execute(core.IRequestReader, core.IResponseWriter) error
	WithLogger(core.ILogger)
}
