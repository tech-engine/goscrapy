package engine

import (
	"context"

	"github.com/tech-engine/goscrapy/pkg/core"
)

type IPipelineManager[OUT any] interface {
	Start(context.Context) error
	Push(core.IOutput[OUT])
	Add(...IPipeline[OUT])
	Stop()
}

type IPipeline[OUT any] interface {
	Open(context.Context) error
	ProcessItem(IPipelineItem, core.IOutput[OUT]) error
	Close()
}

type IPipelineItem interface {
	Set(key string, value any) error
	Get(key string) (any, bool)
	Del(key string)
	Clear()
	Keys() []any
}

type Resetter interface {
	Reset()
}

type IResult interface {
	Request() *core.Request
	Response() core.IResponseReader
	CallbackName() string
	TaskHandle() TaskHandle
	Error() error
	Release()
}


type TaskHandle = core.TaskHandle


type IScheduler interface {
	Start(context.Context) error
	Schedule(req *core.Request, cbName string) error
	NextRequest(ctx context.Context) (*core.Request, string, TaskHandle, error)
	Ack(handle TaskHandle) error
	Nack(handle TaskHandle) error
}

type IWorkerPool interface {
	Start(context.Context) error
	Submit(req *core.Request, cbName string, handle TaskHandle) error
	Results() <-chan IResult
	ReleaseResult(IResult)
}


type ICallbackRegistry interface {
	Register(string, core.ResponseCallback)
	Resolve(string) (core.ResponseCallback, bool)
	Deregister(string)
}
