package worker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/tech-engine/goscrapy/pkg/core"
	"github.com/tech-engine/goscrapy/pkg/engine"
	"github.com/tech-engine/goscrapy/pkg/logger"
	ts "github.com/tech-engine/goscrapy/pkg/telemetry/stats"
)

type workTask struct {
	req          *core.Request
	callbackName string
	taskHandle   core.TaskHandle
}

func (wt *workTask) Reset() {
	wt.req = nil
	wt.callbackName = ""
	wt.taskHandle = nil
}

type Worker struct {
	ID               uint16
	executor         IExecutor
	workerTaskBuffer <-chan *workTask
	workerTaskPool   *sync.Pool
	results          chan<- engine.IResult
	responsePool     *sync.Pool
	stats            ts.IStatsRecorder
	logger           core.ILogger
	tracker          core.IActivityTracker
	onTaskDone       func(time.Duration)
}

func NewWorker(id uint16, executor IExecutor, workerTaskBuffer <-chan *workTask, results chan<- engine.IResult, responsePool *sync.Pool, taskPool *sync.Pool, onTaskDone func(time.Duration)) *Worker {
	return &Worker{
		ID:               id,
		executor:         executor,
		workerTaskBuffer: workerTaskBuffer,
		workerTaskPool:   taskPool,
		results:          results,
		responsePool:     responsePool,
		onTaskDone:       onTaskDone,
		logger:           logger.EnsureLogger(nil).WithName(fmt.Sprintf("Worker-%d", id)),
	}
}

func (w *Worker) Start(ctx context.Context) error {
	var wg sync.WaitGroup
	defer wg.Wait()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case task, ok := <-w.workerTaskBuffer:
			if !ok || task == nil || task.req == nil {
				return nil
			}

			wg.Add(1)
			start := time.Now()

			res := w.execute(ctx, task)

			// return task to pool
			task.Reset()
			w.workerTaskPool.Put(task)

			select {
			case w.results <- res:
			case <-ctx.Done():
				wg.Done()
				return ctx.Err()
			}

			if w.onTaskDone != nil {
				w.onTaskDone(time.Since(start))
			}
			wg.Done()
		}
	}
}

func (w *Worker) execute(ctx context.Context, task *workTask) engine.IResult {
	resp := w.responsePool.Get().(*response)

	// context will be cancelled in ReleaseResult after body is drained
	execCtx, cancel := context.WithCancel(ctx)

	// use request deadline if provided
	if task.req.Ctx != nil {
		if d, ok := task.req.Ctx.Deadline(); ok {
			execCtx, _ = context.WithDeadline(execCtx, d)
		}
	}

	// set our exec context
	task.req.Ctx = execCtx

	if w.stats != nil {
		task.req.Ctx = ts.WithRecorder(execCtx, w.stats)
	}

	if w.tracker != nil {
		w.tracker.Inc()
		defer w.tracker.Dec()
	}

	err := w.executor.Execute(task.req, resp)

	// we transfer request meta to response so callbacks can access it
	if task.req.Meta_ != nil {
		resp.WriteMeta(task.req.Meta_)
	}

	return &result{
		request:      task.req,
		response:     resp,
		callbackName: task.callbackName,
		taskHandle:   task.taskHandle,
		err:          err,
		cancel:       cancel,
	}
}
