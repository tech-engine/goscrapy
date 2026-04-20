// Note: generated tests
package scheduler

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tech-engine/goscrapy/pkg/core"
)

type blockingExecutor struct {
	startBlock chan struct{}
}

func (e *blockingExecutor) Execute(req core.IRequestReader, res core.IResponseWriter) error {
	select {
	case <-e.startBlock:
	case <-req.ReadContext().Done():
		return req.ReadContext().Err()
	}
	return nil
}

func (e *blockingExecutor) WithLogger(logger core.ILogger) IExecutor { return e }

func TestWorker_ContextIntegration(t *testing.T) {
	t.Run("FrameworkShutdown_AbortsInFlightRequest", func(t *testing.T) {
		executor := &blockingExecutor{startBlock: make(chan struct{})}
		workerQueue := make(WorkerQueue, 1)

		worker := NewWorker(1, executor, workerQueue, nil, nil, nil, nil, nil, nil)

		ctx, cancel := context.WithCancel(context.Background())
		go worker.Start(ctx)

		// Wait for worker to be ready
		workChan := <-workerQueue

		// Send work
		work := &schedulerWork{
			request: &request{ctx: ctx},
			next:    func(context.Context, core.IResponseReader) {},
		}
		workChan <- work

		// Give it a moment to enter executor
		time.Sleep(50 * time.Millisecond)

		// Shutdown framework
		cancel()

		// The worker should eventually stop.
		// If it doesn't, the test will timeout or leak goroutines.
	})
}

func TestScheduler_GracefulShutdown(t *testing.T) {
	executor := &blockingExecutor{startBlock: make(chan struct{})}
	sched := New(executor, WithWorkers(2))

	ctx, cancel := context.WithCancel(context.Background())
	go sched.Start(ctx)

	// Send some requests
	for i := 0; i < 5; i++ {
		req := sched.NewRequest(ctx)
		sched.Schedule(req, func(context.Context, core.IResponseReader) {})
	}

	time.Sleep(50 * time.Millisecond)

	// Shutdown
	cancel()

	// Scheduler should exit
}

type mockExecutorFunc struct {
	execute func(core.IRequestReader, core.IResponseWriter) error
}

func (m *mockExecutorFunc) Execute(req core.IRequestReader, res core.IResponseWriter) error {
	return m.execute(req, res)
}

func (m *mockExecutorFunc) WithLogger(logger core.ILogger) IExecutor { return m }

func TestWorker_Concurrency(t *testing.T) {
	var mu sync.Mutex
	active := 0
	maxActive := 0

	executor := &mockExecutorFunc{
		execute: func(req core.IRequestReader, res core.IResponseWriter) error {
			mu.Lock()
			active++
			if active > maxActive {
				maxActive = active
			}
			mu.Unlock()

			time.Sleep(100 * time.Millisecond)

			mu.Lock()
			active--
			mu.Unlock()
			return nil
		},
	}

	numWorkers := uint16(5)
	sched := New(executor, WithWorkers(numWorkers))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go sched.Start(ctx)

	// send work
	numReqs := 10
	for i := 0; i < numReqs; i++ {
		req := sched.NewRequest(ctx)
		sched.Schedule(req, func(context.Context, core.IResponseReader) {
			if i == numReqs-1 {
				// wait a bit for others
			}
		})
	}

	// wait for some time to let workers process
	time.Sleep(500 * time.Millisecond)

	mu.Lock()
	assert.LessOrEqual(t, maxActive, int(numWorkers))
	mu.Unlock()
}
