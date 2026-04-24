// Note: generate benchmark tests for adaptive scaling
package scheduler

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"runtime"
	"sync/atomic"
	"testing"
	"time"

	"github.com/tech-engine/goscrapy/internal/request"
	"github.com/tech-engine/goscrapy/pkg/core"
	"github.com/tech-engine/goscrapy/pkg/logger"
)

// Compares fixed vs adaptive worker pools on the same workload.
// Measures: throughput (req/s), memory, goroutine count.
func TestFixedVsAdaptive(t *testing.T) {
	l := logger.NewNoopLogger()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(20 * time.Millisecond) // simulate 20ms latency
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer srv.Close()

	executor := &benchExecutor{client: srv.Client()}
	totalRequests := 500

	type result struct {
		name       string
		dur        time.Duration
		rps        float64
		goroutines int
		allocKB    int64
	}

	var results []result

	// Fixed Workers (default: GOMAXPROCS * 30)
	for _, numW := range []uint16{4, 120} {
		runtime.GC()
		time.Sleep(50 * time.Millisecond)

		var memBefore runtime.MemStats
		runtime.ReadMemStats(&memBefore)

		ctx, cancel := context.WithCancel(context.Background())
		sched := New(executor, request.NewPool(),
			WithWorkers(numW),
			WithReqResPoolSize(uint64(totalRequests)),
			WithWorkQueueSize(uint64(totalRequests)),
		)
		sched.WithLogger(l)

		go sched.Start(ctx)
		time.Sleep(100 * time.Millisecond)

		var completed atomic.Int64
		start := time.Now()

		for i := 0; i < totalRequests; i++ {
			req := request.NewPool().Acquire(ctx)
			req.URL, _ = url.Parse(srv.URL + "/")
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

		dur := time.Since(start)
		goroutines := runtime.NumGoroutine()

		var memAfter runtime.MemStats
		runtime.ReadMemStats(&memAfter)

		cancel()
		time.Sleep(200 * time.Millisecond)

		results = append(results, result{
			name:       fmt.Sprintf("Fixed-%d", numW),
			dur:        dur,
			rps:        float64(completed.Load()) / dur.Seconds(),
			goroutines: goroutines,
			allocKB:    (int64(memAfter.TotalAlloc) - int64(memBefore.TotalAlloc)) / 1024,
		})
	}

	// Adaptive (start=4, min=4, max=120)
	runtime.GC()
	time.Sleep(50 * time.Millisecond)

	var memBefore runtime.MemStats
	runtime.ReadMemStats(&memBefore)

	ctx, cancel := context.WithCancel(context.Background())
	sched := New(executor, request.NewPool(),
		WithWorkers(4),
		WithAdaptiveScaling(AdaptiveScalingConfig{
			MinWorkers:    4,
			MaxWorkers:    120,
			ScalingFactor: 1.2,
			EMAAlpha:      0.5,
			ScalingWindow: 200 * time.Millisecond,
		}),
		WithReqResPoolSize(uint64(totalRequests)),
		WithWorkQueueSize(uint64(totalRequests)),
	)
	sched.WithLogger(l)

	go sched.Start(ctx)
	time.Sleep(100 * time.Millisecond)

	var completed atomic.Int64
	start := time.Now()

	for i := 0; i < totalRequests; i++ {
		req := request.NewPool().Acquire(ctx)
		req.URL, _ = url.Parse(srv.URL + "/")
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

	dur := time.Since(start)
	goroutines := runtime.NumGoroutine()
	peakWorkers := sched.currentWorkerCnt.Load()

	var memAfter runtime.MemStats
	runtime.ReadMemStats(&memAfter)

	cancel()
	time.Sleep(200 * time.Millisecond)

	results = append(results, result{
		name:       fmt.Sprintf("Adaptive-4..120(peak=%d)", peakWorkers),
		dur:        dur,
		rps:        float64(completed.Load()) / dur.Seconds(),
		goroutines: goroutines,
		allocKB:    (int64(memAfter.TotalAlloc) - int64(memBefore.TotalAlloc)) / 1024,
	})

	// Print Results
	fmt.Printf("\n=== FIXED vs ADAPTIVE (%d requests, 20ms latency) ===\n", totalRequests)
	fmt.Printf("%-30s | %-12s | %-10s | %-12s | %-15s\n",
		"Config", "Duration", "Req/Sec", "Goroutines", "TotalAlloc(KB)")
	fmt.Println("-------------------------------------------------------------------------------------")
	for _, r := range results {
		fmt.Printf("%-30s | %-12s | %-10.0f | %-12d | %-15d\n",
			r.name, r.dur.Round(time.Millisecond), r.rps, r.goroutines, r.allocKB)
	}
	fmt.Println("=====================================================================================")
}

// Tests memory profile during idle period: fixed has many idle goroutines,
// adaptive should scale down to minWorkers.
func TestIdleMemory_FixedVsAdaptive(t *testing.T) {
	l := logger.NewNoopLogger()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer srv.Close()

	executor := &benchExecutor{client: srv.Client()}

	fmt.Printf("\n=== IDLE MEMORY: FIXED vs ADAPTIVE ===\n")
	fmt.Printf("%-25s | %-12s | %-12s | %-15s\n",
		"Config", "Workers", "Goroutines", "Alloc(KB)")
	fmt.Println("--------------------------------------------------------------")

	// Fixed 120 workers (idle)
	runtime.GC()
	time.Sleep(50 * time.Millisecond)
	var memBefore runtime.MemStats
	runtime.ReadMemStats(&memBefore)

	ctx, cancel := context.WithCancel(context.Background())
	sched := New(executor, request.NewPool(), WithWorkers(120))
	sched.WithLogger(l)
	go sched.Start(ctx)
	time.Sleep(300 * time.Millisecond)

	goroutines := runtime.NumGoroutine()
	var memAfter runtime.MemStats
	runtime.ReadMemStats(&memAfter)

	fmt.Printf("%-25s | %-12d | %-12d | %-15d\n",
		"Fixed-120",
		sched.currentWorkerCnt.Load(),
		goroutines,
		(int64(memAfter.Alloc)-int64(memBefore.Alloc))/1024)

	cancel()
	time.Sleep(200 * time.Millisecond)
	// Adaptive (min=4, max=120) - after brief burst then idle
	runtime.GC()
	time.Sleep(50 * time.Millisecond)
	runtime.ReadMemStats(&memBefore)

	ctx, cancel = context.WithCancel(context.Background())
	sched = New(executor, request.NewPool(),
		WithWorkers(4),
		WithAdaptiveScaling(AdaptiveScalingConfig{
			MinWorkers:    4,
			MaxWorkers:    120,
			ScalingWindow: 200 * time.Millisecond,
			EMAAlpha:      0.9,
		}),
	)
	sched.WithLogger(l)
	go sched.Start(ctx)
	time.Sleep(100 * time.Millisecond)

	// Brief burst
	for i := 0; i < 50; i++ {
		req := request.NewPool().Acquire(ctx)
		req.URL, _ = url.Parse(srv.URL + "/")
		done := make(chan struct{})
		sched.Schedule(req, func(ctx context.Context, resp core.IResponseReader) {
			if resp.Body() != nil {
				io.Copy(io.Discard, resp.Body())
			}
			close(done)
		})
		<-done
	}

	// Let it scale down
	time.Sleep(3 * time.Second)

	goroutines = runtime.NumGoroutine()
	runtime.ReadMemStats(&memAfter)

	fmt.Printf("%-25s | %-12d | %-12d | %-15d\n",
		"Adaptive (idle)",
		sched.currentWorkerCnt.Load(),
		goroutines,
		(int64(memAfter.Alloc)-int64(memBefore.Alloc))/1024)

	cancel()
	time.Sleep(200 * time.Millisecond)
	fmt.Println("==============================================================")
}
