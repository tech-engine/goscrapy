package scheduler

import (
	"github.com/tech-engine/goscrapy/pkg/core"
	"github.com/tech-engine/goscrapy/pkg/engine"
)

// An executor must implement the IExecutor interface to be used by the scheduler.*Scheduler
type IExecutor interface {
	Execute(core.IRequestReader, engine.IResponseWriter) error
}
