// Note: AI generated bench
package scheduler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"runtime"
	"sync/atomic"
	"testing"
	"time"

	rp "github.com/tech-engine/goscrapy/internal/resource_pool"
	"github.com/tech-engine/goscrapy/pkg/core"
	"github.com/tech-engine/goscrapy/pkg/engine"
	"github.com/tech-engine/goscrapy/pkg/logger"
)

// passthrough executor that actually hits a real local HTTP server
type benchExecutor struct {
	client *http.Client
}

func (e *benchExecutor) Execute(req core.IRequestReader, res engine.IResponseWriter) error {
	ctx := req.ReadContext()
	if ctx == nil {
		ctx = context.Background()
	}
	method := "GET"
	if req.ReadMethod() != "" {
		method = req.ReadMethod()
	}
	httpReq, err := http.NewRequestWithContext(ctx, method, req.ReadUrl().String(), nil)
	if err != nil {
		return err
	}
	httpReq.Header = req.ReadHeader()
	resp, err := e.client.Do(httpReq)
	if err != nil {
		return err
	}
	res.WriteStatusCode(resp.StatusCode)
	res.WriteHeader(resp.Header)
	res.WriteBody(resp.Body)
	res.WriteRequest(httpReq)
	return nil
}

func (e *benchExecutor) WithLogger(l core.ILogger) {}

// BenchmarkSchedulerPooling measures allocation overhead of the full scheduler pipeline.
func BenchmarkSchedulerPooling(b *testing.B) {
	// Suppress framework logs
	logger.SetLevel(core.LevelNone)

	// Start a local mock server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer ts.Close()

	executor := &benchExecutor{client: ts.Client()}

	sched := New(executor,
		WithWorkers(4),
		WithReqResPoolSize(1024),
		WithWorkQueueSize(1024),
	)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go sched.Start(ctx)
	// let workers spawn
	time.Sleep(100 * time.Millisecond)

	var completed atomic.Int64

	b.ResetTimer()
	b.ReportAllocs()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			req := sched.NewRequest(ctx)
			req.Url(ts.URL + "/")
			done := make(chan struct{})
			sched.Schedule(req, func(ctx context.Context, resp core.IResponseReader) {
				// drain body to ensure it gets pooled properly
				if resp.Body() != nil {
					io.Copy(io.Discard, resp.Body())
				}
				completed.Add(1)
				close(done)
			})
			<-done
		}
	})

	b.StopTimer()
	b.ReportMetric(float64(completed.Load()), "completed")
}

// BenchmarkWorkerExecute measures allocation overhead inside a single worker.execute call.
func BenchmarkWorkerExecute(b *testing.B) {
	logger.SetLevel(core.LevelNone)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer ts.Close()

	executor := &benchExecutor{client: ts.Client()}

	schedulerWorkPool := rp.NewPooler(rp.WithSize[schedulerWork](64))
	requestPool := rp.NewPooler(rp.WithSize[request](64))

	worker := NewWorker(1, executor, make(WorkerQueue, 1), schedulerWorkPool, requestPool, rp.NewPooler(rp.WithSize[response](64)), nil)

	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		req := requestPool.Acquire()
		if req == nil {
			req = &request{method: "GET", header: make(http.Header)}
		}
		req.ctx = ctx
		req.method = "GET"
		req.Url(ts.URL + "/")

		work := schedulerWorkPool.Acquire()
		if work == nil {
			work = &schedulerWork{}
		}
		work.request = req
		work.next = func(ctx context.Context, resp core.IResponseReader) {
			if resp.Body() != nil {
				io.Copy(io.Discard, resp.Body())
			}
		}

		worker.execute(ctx, work)
	}
}

// TestGoroutineStability verifies that workers don't leak goroutines over
// a sustained burst-idle-burst cycle.
func TestGoroutineStability(t *testing.T) {
	logger.SetLevel(core.LevelNone)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer ts.Close()

	executor := &benchExecutor{client: ts.Client()}

	sched := New(executor,
		WithWorkers(8),
		WithReqResPoolSize(256),
		WithWorkQueueSize(256),
	)

	ctx, cancel := context.WithCancel(context.Background())

	go sched.Start(ctx)
	time.Sleep(100 * time.Millisecond)

	baseGoroutines := runtime.NumGoroutine()
	t.Logf("Baseline goroutines: %d", baseGoroutines)

	// BURST 1: send 200 requests
	var completed atomic.Int64
	for i := 0; i < 200; i++ {
		req := sched.NewRequest(ctx)
		req.Url(ts.URL + "/")
		done := make(chan struct{})
		sched.Schedule(req, func(ctx context.Context, resp core.IResponseReader) {
			if resp.Body() != nil {
				io.Copy(io.Discard, resp.Body())
			}
			completed.Add(1)
			close(done)
		})
		<-done
	}

	burstGoroutines := runtime.NumGoroutine()
	t.Logf("After burst 1: %d goroutines, %d completed", burstGoroutines, completed.Load())

	// IDLE: Let the system settle
	time.Sleep(500 * time.Millisecond)

	idleGoroutines := runtime.NumGoroutine()
	t.Logf("After idle: %d goroutines", idleGoroutines)

	// BURST 2: send another 200 requests
	for i := 0; i < 200; i++ {
		req := sched.NewRequest(ctx)
		req.Url(ts.URL + "/")
		done := make(chan struct{})
		sched.Schedule(req, func(ctx context.Context, resp core.IResponseReader) {
			if resp.Body() != nil {
				io.Copy(io.Discard, resp.Body())
			}
			completed.Add(1)
			close(done)
		})
		<-done
	}

	burst2Goroutines := runtime.NumGoroutine()
	t.Logf("After burst 2: %d goroutines, %d total completed", burst2Goroutines, completed.Load())
	
	// Shutdown
	cancel()
	time.Sleep(200 * time.Millisecond)

	finalGoroutines := runtime.NumGoroutine()
	t.Logf("After shutdown: %d goroutines", finalGoroutines)

	// Verify no goroutine leak: final count should be near baseline
	// Allow some tolerance for GC, timers, etc.
	if finalGoroutines > baseGoroutines+5 {
		t.Errorf("potential goroutine leak: baseline=%d, final=%d", baseGoroutines, finalGoroutines)
	}

	if completed.Load() != 400 {
		t.Errorf("expected 400 completed, got %d", completed.Load())
	}

	fmt.Printf("\n=== GOROUTINE STABILITY REPORT ===\n")
	fmt.Printf("Baseline:     %d\n", baseGoroutines)
	fmt.Printf("After Burst1: %d\n", burstGoroutines)
	fmt.Printf("After Idle:   %d\n", idleGoroutines)
	fmt.Printf("After Burst2: %d\n", burst2Goroutines)
	fmt.Printf("After Stop:   %d\n", finalGoroutines)
	fmt.Printf("Completed:    %d\n", completed.Load())
	fmt.Printf("==================================\n")
}
