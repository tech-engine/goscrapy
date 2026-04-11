package engine

import (
	"context"
	"github.com/tech-engine/goscrapy/pkg/core"
)

type IPipelineManager[OUT any] interface {
	Start(context.Context) error
	Push(core.IOutput[OUT])
	WithLogger(core.ILogger)
}

type Resetter interface {
	Reset()
}

type IScheduler interface {
	Start(context.Context) error
	Schedule(core.IRequestReader, core.ResponseCallback)
	NewRequest(context.Context) core.IRequestRW
	WithLogger(core.ILogger)
}
