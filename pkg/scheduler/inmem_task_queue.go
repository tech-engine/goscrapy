package scheduler

import (
	"context"

	"github.com/tech-engine/goscrapy/pkg/engine"
)

type inMemTaskQueue struct {
	tasks chan QueuedTask
}

func newInMemoryTaskQueue(size uint64) ITaskQueue {
	return &inMemTaskQueue{
		tasks: make(chan QueuedTask, size),
	}
}

func (q *inMemTaskQueue) Push(ctx context.Context, task QueuedTask) error {
	select {
	case q.tasks <- task:
		return nil
	default:
		return ErrTaskQueueFull
	}
}

func (q *inMemTaskQueue) Pull(ctx context.Context) (QueuedTask, engine.TaskHandle, error) {
	select {
	case <-ctx.Done():
		return QueuedTask{}, nil, ctx.Err()
	case task, ok := <-q.tasks:
		if !ok {
			return QueuedTask{}, nil, ErrTaskQueueClosed
		}
		// Return a dummy handle for the in-memory queue since Ack/Nack are no-ops
		return task, "inmem_handle", nil
	}
}

func (q *inMemTaskQueue) Ack(ctx context.Context, handle engine.TaskHandle) error {
	return nil
}

func (q *inMemTaskQueue) Nack(ctx context.Context, handle engine.TaskHandle) error {
	return nil
}
