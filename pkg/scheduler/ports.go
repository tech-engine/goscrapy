package scheduler

import (
	"context"

	"github.com/tech-engine/goscrapy/pkg/engine"
)

type ITaskQueue interface {
	Push(context.Context, *QueuedTask) error
	Pull(context.Context) (*QueuedTask, engine.TaskHandle, error)
	Ack(context.Context, engine.TaskHandle) error
	Nack(context.Context, engine.TaskHandle) error
}
