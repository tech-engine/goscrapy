package scheduler

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tech-engine/goscrapy/pkg/core"
)

type dummyExecutor struct{}

func (e *dummyExecutor) Execute(core.IRequestReader, core.IResponseWriter) error {
	return nil
}

func (e *dummyExecutor) WithLogger(logger core.ILogger) IExecutor { return e }

func TestWorker(t *testing.T) {
	// create a worker
	workerQueue := make(WorkerQueue, 1)
	executor := &dummyExecutor{}
	worker := NewWorker(1, executor, workerQueue, nil, nil, nil, nil)

	// start the worker
	ctx, cancel := context.WithCancel(context.Background())
	go worker.Start(ctx)

	// the worker should have added itself to the worker queue
	w := <-workerQueue
	assert.NotNil(t, w)

	// stop the worker
	cancel()
}
