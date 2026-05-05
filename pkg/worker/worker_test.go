package worker

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tech-engine/goscrapy/pkg/core"
	"github.com/tech-engine/goscrapy/pkg/engine"
	"github.com/tech-engine/goscrapy/pkg/logger"
)

type workerTestExecutor struct{}

func (e *workerTestExecutor) Execute(req *core.Request, res core.IResponseWriter) error {
	res.WriteStatusCode(200)
	return nil
}

func TestWorker_ExecuteAndResults(t *testing.T) {
	executor := &workerTestExecutor{}
	workerTaskCh := make(chan *workTask, 1)
	results := make(chan engine.IResult, 1)

	respPool := &sync.Pool{New: func() any { return &response{} }}
	taskPool := &sync.Pool{New: func() any { return &workTask{} }}
	resultPool := &sync.Pool{New: func() any { return &result{} }}

	wp := &workerPool{
		logger: logger.NewLogger(),
	}

	w := NewWorker(1, executor, workerTaskCh, results, respPool, taskPool, resultPool, wp)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan struct{})
	go func() {
		_ = w.Start(ctx)
		close(done)
	}()

	// submit a task
	task := &workTask{
		req:          &core.Request{},
		callbackName: "test_cb",
	}
	workerTaskCh <- task

	// verify result in results channel
	select {
	case res := <-results:
		assert.NotNil(t, res)
		assert.Equal(t, "test_cb", res.CallbackName())
		assert.Equal(t, 200, res.Response().StatusCode())
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for result")
	}

	// verify lifecycle - stop by closing channel
	close(workerTaskCh)
	select {
	case <-done:
		// success
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for worker to stop after channel close")
	}
}

func TestWorker_PoisonSignal(t *testing.T) {
	executor := &workerTestExecutor{}
	workerTaskCh := make(chan *workTask, 1)
	results := make(chan engine.IResult, 1)

	respPool := &sync.Pool{New: func() any { return &response{} }}
	taskPool := &sync.Pool{New: func() any { return &workTask{} }}
	resultPool := &sync.Pool{New: func() any { return &result{} }}

	wp := &workerPool{
		logger: logger.NewLogger(),
	}

	w := NewWorker(1, executor, workerTaskCh, results, respPool, taskPool, resultPool, wp)

	done := make(chan struct{})
	go func() {
		_ = w.Start(context.Background())
		close(done)
	}()

	// verify poison task is working
	workerTaskCh <- nil

	select {
	case <-done:
		// success
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for worker to stop after poison signal")
	}
}
