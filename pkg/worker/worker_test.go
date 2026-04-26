package worker

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tech-engine/goscrapy/pkg/core"
	"github.com/tech-engine/goscrapy/pkg/engine"
)

type dummyExecutor struct{}

func (e *dummyExecutor) Execute(req *core.Request, res core.IResponseWriter) error {
	res.WriteStatusCode(200)
	return nil
}

func TestWorker_Lifecycle(t *testing.T) {
	executor := &dummyExecutor{}
	taskChan := make(chan *workTask, 1)
	results := make(chan engine.IResult, 1)

	respPool := &sync.Pool{New: func() any { return &response{} }}
	workPool := &sync.Pool{New: func() any { return &workTask{} }}

	w := NewWorker(
		1,
		executor,
		taskChan,
		results,
		respPool,
		workPool,
		func(d time.Duration) {},
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		_ = w.Start(ctx)
	}()

	// submit a task
	task := &workTask{
		req:          &core.Request{},
		callbackName: "test_cb",
	}
	taskChan <- task

	// wait for result
	select {
	case res := <-results:
		assert.NotNil(t, res)
		assert.Equal(t, "test_cb", res.CallbackName())
		assert.Equal(t, 200, res.Response().StatusCode())
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for result")
	}
}
