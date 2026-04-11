// Note: AI generated tests
package scheduler

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	rp "github.com/tech-engine/goscrapy/internal/resource_pool"
	"github.com/tech-engine/goscrapy/pkg/core"
)

type blockingExecutor struct {
	started chan struct{}
}

func (e *blockingExecutor) Execute(req core.IRequestReader, res core.IResponseWriter) error {
	close(e.started)
	ctx := req.ReadContext()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(10 * time.Second):
		return fmt.Errorf("timeout waiting for cancellation")
	}
}

func (e *blockingExecutor) WithLogger(logger core.ILogger) {}

func TestWorker_ContextIntegration(t *testing.T) {
	t.Run("FrameworkShutdown_AbortsInFlightRequest", func(t *testing.T) {
		executor := &blockingExecutor{started: make(chan struct{})}
		worker := NewWorker(1, executor, make(WorkerQueue, 1),
			rp.NewPooler(rp.WithSize[schedulerWork](1)),
			rp.NewPooler(rp.WithSize[request](1)),
			rp.NewPooler(rp.WithSize[response](1)), nil)

		// Framework lifecycle context
		fCtx, fCancel := context.WithCancel(context.Background())
		
		// Request context (no specific timeout)
		req := &request{method: "GET", header: make(http.Header), ctx: context.Background()}
		
		work := &schedulerWork{
			request: req,
			next: func(ctx context.Context, resp core.IResponseReader) {},
		}

		errCh := make(chan error, 1)
		go func() {
			errCh <- worker.execute(fCtx, work)
		}()

		// Wait for executor to start
		<-executor.started

		// Simulate global shutdown
		fCancel()

		select {
		case err := <-errCh:
			assert.ErrorIs(t, err, context.Canceled, "request should be aborted when framework stops")
		case <-time.After(2 * time.Second):
			t.Fatal("worker did not abort request after framework cancellation")
		}
	})

	t.Run("RequestTimeout_DoesNotStopWorker", func(t *testing.T) {
		executor := &blockingExecutor{started: make(chan struct{})}
		worker := NewWorker(1, executor, make(WorkerQueue, 1),
			rp.NewPooler(rp.WithSize[schedulerWork](1)),
			rp.NewPooler(rp.WithSize[request](1)),
			rp.NewPooler(rp.WithSize[response](1)), nil)

		// Framework lifecycle context (alive)
		fCtx, fCancel := context.WithCancel(context.Background())
		defer fCancel()
		
		// Request context with short timeout
		rCtx, rCancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
		defer rCancel()
		
		req := &request{method: "GET", header: make(http.Header), ctx: rCtx}
		
		work := &schedulerWork{
			request: req,
			next: func(ctx context.Context, resp core.IResponseReader) {},
		}

		errCh := make(chan error, 1)
		go func() {
			errCh <- worker.execute(fCtx, work)
		}()

		<-executor.started

		select {
		case err := <-errCh:
			assert.ErrorIs(t, err, context.DeadlineExceeded, "request should respect its own timeout")
		case <-time.After(2 * time.Second):
			t.Fatal("request did not respect its own timeout")
		}

		// Verify worker lifecycle is still active by checking fCtx
		assert.NoError(t, fCtx.Err(), "framework lifecycle should remain active")
	})

	t.Run("ValueAndWorkerIDPropagation", func(t *testing.T) {
		// Mock executor that finishes immediately
		executor := &mockExecutorFunc{
			execute: func(req core.IRequestReader, res core.IResponseWriter) error {
				return nil
			},
		}
		
		worker := NewWorker(42, executor, make(WorkerQueue, 1),
			rp.NewPooler(rp.WithSize[schedulerWork](1)),
			rp.NewPooler(rp.WithSize[request](1)),
			rp.NewPooler(rp.WithSize[response](1)), nil)

		fCtx := context.WithValue(context.Background(), "trace-id", "12345")
		
		// Spider provided some context value
		rCtx := context.WithValue(context.Background(), "spider-val", "foobar")
		req := &request{method: "GET", header: make(http.Header), ctx: rCtx}
		
		callbackDone := make(chan struct{})
		work := &schedulerWork{
			request: req,
			next: func(ctx context.Context, resp core.IResponseReader) {
				assert.Equal(t, uint16(42), ctx.Value("WORKER_ID"))
				assert.Equal(t, "12345", ctx.Value("trace-id"), "framework values should propagate")
				assert.Equal(t, "foobar", ctx.Value("spider-val"), "spider values should propagate")
				close(callbackDone)
			},
		}

		err := worker.execute(fCtx, work)
		assert.NoError(t, err)
		
		select {
		case <-callbackDone:
		case <-time.After(1 * time.Second):
			t.Fatal("callback was not called or values were missing")
		}
	})

	t.Run("RecursiveContextPersistence", func(t *testing.T) {
		executor := &mockExecutorFunc{
			execute: func(req core.IRequestReader, res core.IResponseWriter) error {
				res.WriteRequest(new(http.Request))
				return nil
			},
		}

		worker := NewWorker(42, executor, make(WorkerQueue, 1),
			rp.NewPooler(rp.WithSize[schedulerWork](1)),
			rp.NewPooler(rp.WithSize[request](1)),
			rp.NewPooler(rp.WithSize[response](1)), nil)

		fCtx, fCancel := context.WithCancel(context.Background())
		rCtx := context.Background()

		req := &request{method: "GET", header: make(http.Header), ctx: rCtx}
		
		callbackDone := make(chan struct{})
		var callbackCtx context.Context

		work := &schedulerWork{
			request: req,
			next: func(ctx context.Context, resp core.IResponseReader) {
				callbackCtx = ctx
				
				// Assertions must happen INSIDE the callback before worker.execute returns and resets the response
				assert.NoError(t, ctx.Err(), "callback context should be alive")
				assert.NoError(t, resp.Request().Context().Err(), "native http request context should be alive")
				
				close(callbackDone)
			},
		}

		err := worker.execute(fCtx, work)
		assert.NoError(t, err)

		<-callbackDone
		
		// 1. SURVIVAL: At this point worker.execute has returned, meaning its internal defer cancel() has run.
		assert.NoError(t, callbackCtx.Err(), "callback context should NOT be cancelled after execution completes")
		
		// 2. VALUE INTEGRITY: Framework values should still be accessible
		assert.Equal(t, uint16(42), callbackCtx.Value("WORKER_ID"), "survived context should still resolve framework values")
		
		// 3. PERSISTENCE: Engine shutdown should not kill the "idle" callback context (it binds to new requests later)
		fCancel()
		assert.NoError(t, callbackCtx.Err(), "idle callback context should survive framework shutdown until next execution")

		// 4. SESSION COHERENCE: If the session itself is cancelled, then the callback context MUST die
		sCtx, sCancel := context.WithCancel(context.Background())
		
		// Re-initialize request since worker.execute resets the work object
		work.request = &request{method: "GET", header: make(http.Header), ctx: sCtx}
		
		sessionCallbackDone := make(chan struct{})
		work.next = func(ctx context.Context, resp core.IResponseReader) {
			callbackCtx = ctx
			close(sessionCallbackDone)
		}
		
		err = worker.execute(context.Background(), work)
		assert.NoError(t, err)
		<-sessionCallbackDone
		
		sCancel()
		assert.ErrorIs(t, callbackCtx.Err(), context.Canceled, "callback context MUST cancel if the underlying session cancels")
	})
}

type mockExecutorFunc struct {
	execute func(req core.IRequestReader, res core.IResponseWriter) error
}

func (m *mockExecutorFunc) Execute(req core.IRequestReader, res core.IResponseWriter) error {
	return m.execute(req, res)
}

func (m *mockExecutorFunc) WithLogger(logger core.ILogger) {}
