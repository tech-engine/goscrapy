package worker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/tech-engine/goscrapy/pkg/core"
	"github.com/tech-engine/goscrapy/pkg/engine"
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
	ID             uint16
	executor       IExecutor
	workerTaskCh   <-chan *workTask
	results        chan<- engine.IResult
	workerTaskPool *sync.Pool
	responsePool   *sync.Pool
	resultPool     *sync.Pool
	tracker        core.IActivityTracker
	logger         core.ILogger
	pool           *workerPool
}

func NewWorker(id uint16, executor IExecutor, workerTaskCh <-chan *workTask, results chan<- engine.IResult, responsePool *sync.Pool, taskPool *sync.Pool, resultPool *sync.Pool, pool *workerPool) *Worker {
	return &Worker{
		ID:             id,
		executor:       executor,
		workerTaskCh:   workerTaskCh,
		results:        results,
		responsePool:   responsePool,
		workerTaskPool: taskPool,
		resultPool:     resultPool,
		pool:           pool,
		logger:         pool.logger.WithName(fmt.Sprintf("Worker-%d", id)),
	}
}

func (w *Worker) Start(ctx context.Context) error {
	// create per worker context
	workerLocalCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// we revert to select
	for {
		select {
		case <-workerLocalCtx.Done():
			w.logger.Debugf("worker id: %d context cancelled", w.ID)
			return workerLocalCtx.Err()
		case task, ok := <-w.workerTaskCh:
			if !ok {
				w.logger.Debugf("worker id: %d task channel closed", w.ID)
				return nil
			}
			if task == nil {
			// poision signal
				w.logger.Debugf("got poision signal, killing worker id: %d", w.ID)
				return nil
			}

			if task.req == nil {
				w.logger.Warn("got nil task")
				continue
			}

			res := w.execute(workerLocalCtx, task)

			// return task to pool
			task.Reset()
			w.workerTaskPool.Put(task)

			if w.results != nil {
				select {
				case <-workerLocalCtx.Done():
					res.Release()
					return workerLocalCtx.Err()
				case w.results <- res:
				}
			}
		}
	}
}

func (w *Worker) execute(ctx context.Context, task *workTask) engine.IResult {
	resp := w.responsePool.Get().(*response)

	// merge contexts to handle timeouts and graceful shutdown
	execCtx, cleanup := mergeContexts(ctx, task.req.Ctx)

	task.req.Ctx = execCtx

	if w.tracker != nil {
		w.tracker.Inc()
		defer w.tracker.Dec()
	}

	start := time.Now()
	err := w.executor.Execute(task.req, resp)

	if w.pool.autoscaler != nil {
		w.pool.autoscaler.OnTaskDone(time.Since(start))
	}

	// we transfer request meta to response so callbacks can access it
	if task.req.Meta_ != nil {
		resp.WriteMeta(task.req.Meta_)
	}

	if w.pool.signals != nil {
		if err != nil {
			w.pool.signals.EmitRequestError(context.Background(), task.req, err)
		} else {
			w.pool.signals.EmitResponseReceived(context.Background(), resp)
		}
	}

	res := w.resultPool.Get().(*result)
	res.request = task.req
	res.response = resp
	res.callbackName = task.callbackName
	res.taskHandle = task.taskHandle
	res.err = err
	res.cancel = cleanup

	return res
}
