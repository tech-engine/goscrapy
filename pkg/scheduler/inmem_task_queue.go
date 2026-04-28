package scheduler

import (
	"context"

	"github.com/tech-engine/goscrapy/pkg/engine"
)

type inMemTaskQueue struct {
	taskCh chan *QueuedTask
}

// Default: inmem implementation of ITaskQueue
func newInMemoryTaskQueue(size uint64) ITaskQueue {
	return &inMemTaskQueue{
		taskCh: make(chan *QueuedTask, size),
	}
}

func (q *inMemTaskQueue) Push(ctx context.Context, task *QueuedTask) error {
	select {
	case q.taskCh <- task:
		return nil
	default:
		return ErrTaskQueueFull
	}
}

func (q *inMemTaskQueue) Pull(ctx context.Context) (*QueuedTask, engine.TaskHandle, error) {
	select {
	case <-ctx.Done():
		return nil, nil, ctx.Err()
	case task, ok := <-q.taskCh:
		if !ok {
			return nil, nil, ErrTaskQueueClosed
		}

		return task, task, nil
	}
}

func (q *inMemTaskQueue) Ack(ctx context.Context, handle engine.TaskHandle) error {
	return nil
}

func (q *inMemTaskQueue) Nack(ctx context.Context, handle engine.TaskHandle) error {
	return nil
}
