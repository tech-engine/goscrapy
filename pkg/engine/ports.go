package engine

import (
	"context"

	"github.com/tech-engine/goscrapy/pkg/core"
)

type IPipelineManager[OUT any] interface {
	Start(context.Context) error
	Push(core.IOutput[OUT])
	// WithLogger(core.ILogger) IPipelineManager[OUT]
	// WithActivityTracker(core.IActivityTracker) IPipelineManager[OUT]
}

type Resetter interface {
	Reset()
}

type IScheduler interface {
	Start(context.Context) error
	Schedule(*core.Request, string)
	// Schedule(*core.Request, core.ResponseCallback)
	NextRequest() (*core.Request, string, error)
	// WithLogger(core.ILogger) IScheduler
	// WithActivityTracker(core.IActivityTracker) IScheduler
}

type ICallbackRegistry interface {
	Register(string, core.ResponseCallback)
	Resolve(string) (core.ResponseCallback, bool)
	Deregister(string)
}
