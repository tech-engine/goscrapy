// Note: generated benchmark
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

	"github.com/tech-engine/goscrapy/pkg/core"
	"github.com/tech-engine/goscrapy/pkg/logger"
)

// Measures the memory cost of idle workers.
// If idle workers are cheap, dynamic scaling adds complexity for no benefit.
func TestIdleWorkerCost(t *testing.T) {
	l := logger.NewNoopLogger()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer ts.Close()

	executor := &benchExecutor{client: ts.Client()}

	workerCounts := []uint16{4, 16, 64, 120, 240}

	fmt.Printf("\n=== IDLE WORKER MEMORY COST ===\n")
	fmt.Printf("%-10s | %-12s | %-12s | %-15s | %-15s\n",
		"Workers", "Goroutines", "Alloc (KB)", "TotalAlloc (KB)", "Sys (KB)")
	fmt.Println("--------------------------------------------------------------------------")

	for _, numWorkers := range workerCounts {
		// Force GC before measurement
		runtime.GC()
		time.Sleep(50 * time.Millisecond)

		var memBefore runtime.MemStats
		runtime.ReadMemStats(&memBefore)
		goroutinesBefore := runtime.NumGoroutine()

		ctx, cancel := context.WithCancel(context.Background())

		sched := New(executor,
			WithWorkers(numWorkers),
			WithReqResPoolSize(256),
			WithWorkQueueSize(256),
		).WithLogger(l)

		go sched.Start(ctx)
		time.Sleep(200 * time.Millisecond) // let workers spin up

		goroutinesAfter := runtime.NumGoroutine()

		var memAfter runtime.MemStats
		runtime.ReadMemStats(&memAfter)

		allocDelta := int64(memAfter.Alloc) - int64(memBefore.Alloc)
		totalAllocDelta := int64(memAfter.TotalAlloc) - int64(memBefore.TotalAlloc)
		sysDelta := int64(memAfter.Sys) - int64(memBefore.Sys)

		fmt.Printf("%-10d | %-12d | %-12d | %-15d | %-15d\n",
			numWorkers,
			goroutinesAfter-goroutinesBefore,
			allocDelta/1024,
			totalAllocDelta/1024,
			sysDelta/1024,
		)

		cancel()
		time.Sleep(200 * time.Millisecond) // let workers exit
	}
	fmt.Println("=================================================================")
}

// Measures throughput under different worker counts with the same workload.
// This shows if fewer workers can handle the same load efficiently.
func TestWorkerCountVsThroughput(t *testing.T) {
	l := logger.NewNoopLogger()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer ts.Close()

	executor := &benchExecutor{client: ts.Client()}

	workerCounts := []uint16{4, 8, 16, 32, 64, 120}
	totalRequests := 2000

	fmt.Printf("\n=== WORKER COUNT vs THROUGHPUT (%d requests) ===\n", totalRequests)
	fmt.Printf("%-10s | %-12s | %-12s | %-15s\n",
		"Workers", "Duration", "Req/Sec", "Goroutines")
	fmt.Println("----------------------------------------------------------")

	for _, numWorkers := range workerCounts {
		runtime.GC()

		ctx, cancel := context.WithCancel(context.Background())

		sched := New(executor,
			WithWorkers(numWorkers),
			WithReqResPoolSize(uint64(totalRequests)),
			WithWorkQueueSize(uint64(totalRequests)),
		).WithLogger(l)

		go sched.Start(ctx)
		time.Sleep(100 * time.Millisecond)

		var completed atomic.Int64
		start := time.Now()

		for i := 0; i < totalRequests; i++ {
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

		duration := time.Since(start)
		goroutines := runtime.NumGoroutine()
		rps := float64(completed.Load()) / duration.Seconds()

		fmt.Printf("%-10d | %-12s | %-12.0f | %-15d\n",
			numWorkers, duration.Round(time.Millisecond), rps, goroutines)

		cancel()
		time.Sleep(200 * time.Millisecond)
	}
	fmt.Println("==========================================================")
}

// Measures BURST-IDLE-BURST behavior: how quickly the system adapts
// when load drops to zero then spikes again.
func TestBurstIdleBurstLatency(t *testing.T) {
	l := logger.NewNoopLogger()

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))
	defer ts.Close()

	executor := &benchExecutor{client: ts.Client()}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sched := New(executor,
		WithWorkers(32),
		WithReqResPoolSize(1024),
		WithWorkQueueSize(1024),
	).WithLogger(l)

	go sched.Start(ctx)
	time.Sleep(100 * time.Millisecond)

	burstSize := 500

	fmt.Printf("\n=== BURST-IDLE-BURST LATENCY (burst=%d) ===\n", burstSize)

	for cycle := 1; cycle <= 3; cycle++ {
		var completed atomic.Int64
		start := time.Now()

		for i := 0; i < burstSize; i++ {
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

		duration := time.Since(start)
		rps := float64(completed.Load()) / duration.Seconds()
		goroutines := runtime.NumGoroutine()

		fmt.Printf("  Cycle %d: %s (%0.f req/s), goroutines=%d\n",
			cycle, duration.Round(time.Millisecond), rps, goroutines)

		// Idle period between bursts
		if cycle < 3 {
			fmt.Printf("  ... idle for 2s ...\n")
			time.Sleep(2 * time.Second)
			fmt.Printf("  After idle: goroutines=%d\n", runtime.NumGoroutine())
		}
	}
	fmt.Println("============================================")
}

// Measures overhead of creating a scheduler with many workers quickly
// to see if startup cost is a concern.
func BenchmarkSchedulerStartup(b *testing.B) {
	l := logger.NewNoopLogger()

	executor := &benchExecutor{client: http.DefaultClient}

	workerCounts := []uint16{4, 32, 120, 240}

	for _, n := range workerCounts {
		b.Run(fmt.Sprintf("workers_%d", n), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				ctx, cancel := context.WithCancel(context.Background())
				sched := New(executor,
					WithWorkers(n),
					WithReqResPoolSize(256),
					WithWorkQueueSize(256),
				).WithLogger(l)
				go sched.Start(ctx)
				// let workers actually spawn
				runtime.Gosched()
				time.Sleep(10 * time.Millisecond)
				cancel()
				time.Sleep(10 * time.Millisecond)
			}
		})
	}
}
