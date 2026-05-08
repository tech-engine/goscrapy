package worker

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tech-engine/goscrapy/pkg/core"
)

type poolTestExecutor struct {
	tasksProcessed atomic.Int32
	sleep          time.Duration
}

func (e *poolTestExecutor) Execute(req *core.Request, res core.IResponseWriter) error {
	if e.sleep > 0 {
		time.Sleep(e.sleep)
	}
	e.tasksProcessed.Add(1)
	res.WriteStatusCode(200)
	return nil
}

func TestPool_Lifecycle(t *testing.T) {
	minWorkers := uint32(3)
	executor := &poolTestExecutor{}

	config := &Config{
		Executor: executor,
		Autoscaler: &AutoscalerConfig{
			MinWorkers: minWorkers,
			MaxWorkers: 10,
		},
	}

	wp, err := NewPool(config)
	assert.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error, 1)
	go func() {
		done <- wp.Start(ctx)
	}()

	// verify it starts with desired minimum workers
	// wait a bit for workers to spawn
	time.Sleep(50 * time.Millisecond)
	assert.Equal(t, int32(minWorkers), wp.(*workerPool).activeWorkers.Load())

	// verify lifecycle - stop gracefully
	// this should trigger the stop method of worker pool
	cancel()

	select {
	case err := <-done:
		assert.NoError(t, err)
	case <-time.After(2 * time.Second):
		t.Fatal("Pool did not shut down gracefully")
	}

	// verify results channel is closed
	_, ok := <-wp.Results()
	assert.False(t, ok, "Results channel should be closed after shutdown")

	// verify all workers finished
	assert.Equal(t, int32(0), wp.(*workerPool).activeWorkers.Load())
}

func TestPool_ExecuteAndResults(t *testing.T) {
	executor := &poolTestExecutor{sleep: 10 * time.Millisecond}
	config := &Config{
		Executor: executor,
	}

	wp, err := NewPool(config)
	assert.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go wp.Start(ctx)

	// submit tasks and receive results
	const taskCount = 5
	for range taskCount {
		err := wp.Submit(context.Background(), &core.Request{}, "test_cb", nil)
		assert.NoError(t, err)
	}

	resultsCounts := 0
	for range taskCount {
		select {
		case res := <-wp.Results():
			assert.NotNil(t, res)
			assert.Equal(t, 200, res.Response().StatusCode())
			resultsCounts++
			wp.ReleaseResult(res)
		case <-time.After(1 * time.Second):
			t.Fatal("Timed out waiting for result")
		}
	}

	assert.Equal(t, taskCount, resultsCounts)
	assert.Equal(t, int32(taskCount), executor.tasksProcessed.Load())
}

func TestPool_GracefulDrain(t *testing.T) {
	// verify workers finish executing tasks before exiting
	executor := &poolTestExecutor{sleep: 100 * time.Millisecond}
	config := &Config{
		Executor: executor,
	}

	wp, err := NewPool(config)
	assert.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		_ = wp.Start(ctx)
		close(done)
	}()

	// submit a task that takes time
	err = wp.Submit(context.Background(), &core.Request{}, "slow_task", nil)
	assert.NoError(t, err)

	// wait for it to start
	time.Sleep(20 * time.Millisecond)

	// shutdown immediately
	cancel()

	// verify pool exits gracefully and waits for workers to finish
	select {
	case <-done:
		// verify the slow task actually finished
		assert.Equal(t, int32(1), executor.tasksProcessed.Load())
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Pool did not wait for executing task to finish")
	}
}

func TestPool_SubmitAfterStop(t *testing.T) {
	wp, _ := NewPool(&Config{Executor: &poolTestExecutor{}})
	wp.(*workerPool).isActive.Store(false)

	err := wp.Submit(context.Background(), &core.Request{}, "cb", nil)
	assert.Error(t, err)
	assert.Equal(t, ErrPoolClosed, err)
}

func TestPool_SubmitCanceledCtx(t *testing.T) {
	wp, _ := NewPool(&Config{Executor: &poolTestExecutor{}})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := wp.Submit(ctx, &core.Request{}, "cb", nil)
	assert.Error(t, err)
	assert.Equal(t, context.Canceled, err)
}

func TestPool_EnvConfigOverride(t *testing.T) {
	t.Setenv("AUTOSCALER_MAX_WORKERS", "10")
	t.Setenv("AUTOSCALER_MIN_WORKERS", "20")
	t.Setenv("POOL_MAX_DRAIN_BYTES", "65536")

	wp, _ := NewPool(&Config{Executor: &poolTestExecutor{}})
	pool := wp.(*workerPool)

	// maxWorkers should be clamped to minWorkers
	assert.Equal(t, uint16(20), pool.autoscaler.minWorkers)
	assert.Equal(t, uint16(20), pool.autoscaler.maxWorkers)
	assert.Equal(t, int64(65536), pool.maxDrainBytes)
}

func TestPool_Metrics(t *testing.T) {
	wp, _ := NewPool(&Config{Executor: &poolTestExecutor{}, EnableMetrics: true})
	pool := wp.(*workerPool)

	for range 5 {
		pool.metrics.recordSubmit()
	}
	for range 2 {
		pool.metrics.recordDrop()
	}

	snap := pool.Snapshot().(WorkerPoolSnapshot)
	assert.Equal(t, uint64(5), snap.TasksSubmitted)
	assert.Equal(t, uint64(2), snap.TasksDropped)
}
