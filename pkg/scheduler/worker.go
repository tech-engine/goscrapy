package scheduler

import (
	"context"
	"io"
	"sync"

	rp "github.com/tech-engine/goscrapy/internal/resource_pool"
	"github.com/tech-engine/goscrapy/pkg/core"
	ts "github.com/tech-engine/goscrapy/pkg/telemetry/stats"
)

// Worker will handle the execution of a Work unit
type Worker struct {
	ID                uint16
	executor          IExecutor
	workerQueue       WorkerQueue
	workQueue         WorkQueue
	schedulerWorkPool *rp.Pooler[schedulerWork]
	responsePool      *rp.Pooler[response]
	requestPool       *rp.Pooler[request]
	stats             ts.StatRecorder
}

func NewWorker(id uint16, executor IExecutor, workerQueue WorkerQueue, schedulerWorkPool *rp.Pooler[schedulerWork], requestPool *rp.Pooler[request], responsePool *rp.Pooler[response], stats ts.StatRecorder) *Worker {

	return &Worker{
		ID:                id,
		workerQueue:       workerQueue,
		executor:          executor,
		workQueue:         make(WorkQueue),
		schedulerWorkPool: schedulerWorkPool,
		requestPool:       requestPool,
		responsePool:      responsePool,
		stats:             stats,
	}
}

// Handles listen for any incoming work in workQueue
func (w *Worker) Start(ctx context.Context) error {
	var err error

	if err = ctx.Err(); err != nil {
		return err
	}

	var wg sync.WaitGroup

	// we wait for all worker jobs to be completed finished/fail afer context cancellation
	defer wg.Wait()

	for {

		// make this worker available again
		w.workerQueue <- w.workQueue

		select {
		case work := <-w.workQueue:

			if err = ctx.Err(); err != nil {
				return err
			}

			wg.Add(1)

			// we don't want the workers to crash, so we ignore the error from execute
			_ = w.execute(ctx, work)
			wg.Done()

		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (w *Worker) execute(ctx context.Context, work *schedulerWork) error {
	res := w.responsePool.Acquire()
	if res == nil {
		res = &response{}
	}

	// merge framework lifecycle and req context
	reqCtx := work.request.ReadContext()
	if reqCtx == nil {
		reqCtx = context.Background()
	}

	// use engine lifecycle as parent
	baseCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// link req cancel/timeout
	if reqCtx != context.Background() {
		// inherit deadline
		if d, ok := reqCtx.Deadline(); ok {
			var dCancel context.CancelFunc
			baseCtx, dCancel = context.WithDeadline(baseCtx, d)
			defer dCancel()
		}

		// abort if req is cancelled. captured baseCtx is stable here.
		go func() {
			select {
			case <-reqCtx.Done():
				if reqCtx.Err() != context.DeadlineExceeded {
					cancel()
				}
			case <-baseCtx.Done():
			}
		}()
	}

	// merge values from both context chains
	fullCtx := &mergedContext{Context: baseCtx, reqCtx: reqCtx}

	// For later reference if I forgot for any reason, which is very likely.
	// inject recorder + unified context
	// although work.request is of type core.IRequestReader, but under the hood it is actually as
	// core.IRequestWriter because Scheduler.NewRequest returns a core.IRequestRW, which is what
	// work.request is actually.
	reqWriter := work.request.(core.IRequestWriter)
	reqWriter.Context(fullCtx)

	if w.stats != nil {
		reqWriter.Context(ts.WithRecorder(fullCtx, w.stats))
	}

	err := w.executor.Execute(work.request, res)

	if err == nil {
		next := work.next
		if next != nil {
			res.WriteMeta(work.request.ReadMeta())

			// build persistent callback context isolated from exec cancellation
			cbCtx := &callbackContext{
				Context:      reqCtx,
				frameworkCtx: ctx,
			}
			cbValCtx := context.WithValue(cbCtx, "WORKER_ID", w.ID)

			// revert request context for safety
			// the network phase is over, this request is now bound to the long-lived session context again
			reqWriter.Context(cbValCtx)

			// sync native http request context for consistency if available
			if r := res.Request(); r != nil {
				res.WriteRequest(r.WithContext(cbValCtx))
			}

			// pass persistent context to spider callback
			next(cbValCtx, res)
		}
	}

	w.resetAndRelease(work)
	if res.body != nil {
		io.Copy(io.Discard, res.body)
		res.body.Close()
	}
	res.Reset()
	w.responsePool.Release(res)

	return err
}

// helper to join two context value chains
// prioritize lifecycle values
// fallback to spider values
type mergedContext struct {
	context.Context
	reqCtx context.Context
}

func (m *mergedContext) Value(key any) any {
	if val := m.Context.Value(key); val != nil {
		return val
	}
	return m.reqCtx.Value(key)
}

// helper to join two context value chains for callbacks
// prioritize spider values
// fallback to lifecycle values
type callbackContext struct {
	context.Context
	frameworkCtx context.Context
}

func (c *callbackContext) Value(key any) any {
	if val := c.Context.Value(key); val != nil {
		return val
	}
	return c.frameworkCtx.Value(key)
}

func (w *Worker) resetAndRelease(work *schedulerWork) {
	// release *request to pool
	req, ok := work.request.(*request)

	if !ok {
		return
	}

	req.Reset()

	w.requestPool.Release(req)

	// release *schedulerWork to pool
	work.Reset()

	w.schedulerWorkPool.Release(work)
}
