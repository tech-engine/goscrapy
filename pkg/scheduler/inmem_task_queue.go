package scheduler

import (
	"context"
	"sync"

	"github.com/tech-engine/goscrapy/pkg/engine"
)

type inMemTaskQueue struct {
	taskCh chan *inMemTask
	pool   sync.Pool
}

type inMemTask struct {
	task *QueuedTask
}

// Default: inmem implementation of ITaskQueue
func newInMemoryTaskQueue(size uint64) ITaskQueue {
	return &inMemTaskQueue{
		taskCh: make(chan *inMemTask, size),
		pool: sync.Pool{New: func() any {
			return &inMemTask{}
		}},
	}
}

func (q *inMemTaskQueue) Push(ctx context.Context, task *QueuedTask) error {
	inMemTask, ok := q.pool.Get().(*inMemTask)

	if !ok {
		return ErrFailedGetTask
	}

	inMemTask.task = task
	select {
	case q.taskCh <- inMemTask:
		return nil
	default:
		inMemTask.task = nil
		q.pool.Put(inMemTask)

		return ErrTaskQueueFull
	}
}

func (q *inMemTaskQueue) Pull(ctx context.Context) (*QueuedTask, engine.TaskHandle, error) {
	select {
	case <-ctx.Done():
		return nil, nil, ctx.Err()
	case item, ok := <-q.taskCh:
		if !ok {
			return nil, nil, ErrTaskQueueClosed
		}

		task := item.task
		item.task = nil
		q.pool.Put(item)

		return task, task, nil
	}
}

func (q *inMemTaskQueue) Ack(ctx context.Context, handle engine.TaskHandle) error {
	return nil
}

func (q *inMemTaskQueue) Nack(ctx context.Context, handle engine.TaskHandle) error {
	return nil
}
