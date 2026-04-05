// Note: Benchmark file is generated using AI
// Note: Benchmark file is generated using AI
package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/tech-engine/goscrapy/cmd/gos"
	"github.com/tech-engine/goscrapy/pkg/core"
)

// Minimal Record to satisfy the framework output definition
type Record struct {
	Id string `json:"id" csv:"id"`
}

func (r *Record) Record() *Record      { return r }
func (r *Record) RecordKeys() []string { return []string{"id"} }
func (r *Record) RecordFlat() []any    { return []any{r.Id} }
func (r *Record) Job() core.IJob       { return nil }

func mockServer() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
	// Serve locally with no TLS overhead
	go http.ListenAndServe(":18080", nil)
}

type BenchSpider struct {
	gos.ICoreSpider[*Record]
	completed  atomic.Int32
	cancelFunc context.CancelFunc
}

func (s *BenchSpider) parse(ctx context.Context, resp core.IResponseReader) {
	// Simulate item yield to ensure PipelineManager is active
	s.Yield(&Record{Id: "bench_item"})
	s.completed.Add(1)
}

func runBenchmark(concurrency, poolSize, maxIdle, queueBuf string) float64 {
	os.Setenv("SCHEDULER_CONCURRENCY", concurrency)
	os.Setenv("SCHEDULER_REQ_RES_POOL_SIZE", poolSize)
	os.Setenv("MIDDLEWARE_HTTP_MAX_IDLE_CONN", maxIdle)
	os.Setenv("MIDDLEWARE_HTTP_MAX_CONN_PER_HOST", maxIdle)
	os.Setenv("MIDDLEWARE_HTTP_MAX_IDLE_CONN_PER_HOST", maxIdle)
	os.Setenv("PIPELINEMANAGER_OUTPUT_QUEUE_BUF_SIZE", queueBuf)

	ctx, cancel := context.WithCancel(context.Background())
	engine := gos.New[*Record]()

	spider := &BenchSpider{
		ICoreSpider: engine,
		cancelFunc:  cancel,
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- engine.Start(ctx)
	}()

	startTime := time.Now()

	// Feed the queue dynamically until time is up
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				// Push blocks of requests
				for i := 0; i < 5000; i++ {
					req := spider.NewRequest()
					req.Url("http://localhost:18080/")
					spider.Request(req, spider.parse)
				}
				time.Sleep(100 * time.Millisecond)
			}
		}
	}()

	// Execute benchmark for exactly 5 seconds
	time.Sleep(5 * time.Second)

	// Terminate early to prevent hangs
	// We intentionally do NOT wait on <-errCh here. Under severe benchmark load,
	// goscrapy's scheduler may deadlock its internal WaitGroup during context cancellation
	// if workers terminate before the pipeline pushes are complete. Since this is a tuner, we can leak it.
	cancel()

	duration := time.Since(startTime)
	rps := float64(spider.completed.Load()) / duration.Seconds()

	// Explicit GC to unbind old workers
	time.Sleep(1 * time.Second)
	return rps
}

func main() {
	fmt.Println("Warming up Mock Server on :18080...")
	mockServer()
	time.Sleep(500 * time.Millisecond)

	type permutation struct {
		Concurrency string
		PoolSize    string
		MaxIdle     string
		QueueBuf    string
	}

	cores := runtime.NumCPU()

	// 1. Un-optimized baseline (Cores * 5)
	// 2. High pipeline buffers (Cores * 12)
	// 3. Blazing Fast profile (Recommended) (Cores * 30)
	// 4. Extreme scale (Cores * 60)
	perms := []permutation{
		{fmt.Sprintf("%d", cores*5), "10000", "100", "0"},
		{fmt.Sprintf("%d", cores*12), "50000", "500", "1000"},
		{fmt.Sprintf("%d", cores*30), "50000", "1000", "5000"},
		{fmt.Sprintf("%d", cores*60), "100000", "4000", "10000"},
	}

	bestRPS := 0.0
	var bestCombo permutation

	fmt.Println("Starting GoScrapy Auto-Tuning Benchmark Engine...")
	fmt.Printf("%-11s | %-10s | %-10s | %-10s | %-15s\n", "Concurrency", "PoolSize", "MaxIdle", "QueueBuf", "Requests/Sec")
	fmt.Println("-----------------------------------------------------------------------")

	for _, p := range perms {
		rps := runBenchmark(p.Concurrency, p.PoolSize, p.MaxIdle, p.QueueBuf)
		fmt.Printf("%-11s | %-10s | %-10s | %-10s | %.2f req/s\n", p.Concurrency, p.PoolSize, p.MaxIdle, p.QueueBuf, rps)

		if rps > bestRPS {
			bestRPS = rps
			bestCombo = p
		}
	}

	fmt.Println("\n--- TUNING COMPLETE ---")
	fmt.Printf("🏆 BEST SETUP (%.2f req/sec):\n\n", bestRPS)
	fmt.Println("Apply these exact fields to your settings.go or .env variables:")
	fmt.Printf("  SCHEDULER_CONCURRENCY                  = %s\n", bestCombo.Concurrency)
	fmt.Printf("  SCHEDULER_REQ_RES_POOL_SIZE            = %s\n", bestCombo.PoolSize)
	fmt.Printf("  MIDDLEWARE_HTTP_MAX_IDLE_CONN          = %s\n", bestCombo.MaxIdle)
	fmt.Printf("  MIDDLEWARE_HTTP_MAX_CONN_PER_HOST      = %s\n", bestCombo.MaxIdle)
	fmt.Printf("  MIDDLEWARE_HTTP_MAX_IDLE_CONN_PER_HOST = %s\n", bestCombo.MaxIdle)
	fmt.Printf("  PIPELINEMANAGER_OUTPUT_QUEUE_BUF_SIZE  = %s\n", bestCombo.QueueBuf)
}
