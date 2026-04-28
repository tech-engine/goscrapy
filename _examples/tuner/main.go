// Note: Benchmark file is generated using AI
package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/tech-engine/goscrapy/pkg/core"
	"github.com/tech-engine/goscrapy/pkg/engine"
	"github.com/tech-engine/goscrapy/pkg/gos"
	"github.com/tech-engine/goscrapy/pkg/logger"
)

// Record represents a minimal data structure for benchmarking.
type Record struct {
	Id string `json:"id" csv:"id"`
}

func (r *Record) Record() *Record      { return r }
func (r *Record) RecordKeys() []string { return []string{"id"} }
func (r *Record) RecordFlat() []any    { return []any{r.Id} }
func (r *Record) Job() core.IJob       { return nil }

// NoopPipeline silences warnings and ensures the pipeline hot path is executed.
type NoopPipeline struct{}

func (p *NoopPipeline) ProcessItem(pi engine.IPipelineItem, out core.IOutput[*Record]) error {
	return nil
}

func (p *NoopPipeline) Open(ctx context.Context) error { return nil }
func (p *NoopPipeline) Close()                         {}

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
	completed atomic.Int32
}

func (s *BenchSpider) parse(ctx context.Context, resp core.IResponseReader) {
	// Yield a record to ensure PipelineManager is executed
	s.Yield(&Record{Id: "bench_item"})
	s.completed.Add(1)
}

func runBenchmark(concurrency, maxIdle, queueBuf, resultHandlers string) float64 {
	os.Setenv("SCHEDULER_CONCURRENCY", concurrency)
	os.Setenv("MIDDLEWARE_HTTP_MAX_IDLE_CONN", maxIdle)
	os.Setenv("MIDDLEWARE_HTTP_MAX_CONN_PER_HOST", maxIdle)
	os.Setenv("MIDDLEWARE_HTTP_MAX_IDLE_CONN_PER_HOST", maxIdle)
	os.Setenv("PIPELINEMANAGER_OUTPUT_QUEUE_BUF_SIZE", queueBuf)
	os.Setenv("ENGINE_RESULT_HANDLERS", resultHandlers)

	// Use a timeout context to run each benchmark for a fixed duration
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Root logger for tuning, nullified to avoid interference
	l := logger.NewNoopLogger()

	// Initialize the application using the modern factory
	app, err := gos.New[*Record]()
	if err != nil {
		fmt.Printf("failed to create app: %v\n", err)
		return 0
	}
	app.WithLogger(l).WithPipelines(&NoopPipeline{})

	spider := &BenchSpider{
		ICoreSpider: app,
	}
	app.RegisterSpider(spider)

	go func() {
		_ = app.Start(ctx)
	}()

	startTime := time.Now()

	// Feed the queue dynamically until the timeout is reached
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				// Push blocks of requests to saturate the engine
				for i := 0; i < 1000; i++ {
					req := spider.Request(ctx)
					req.Url("http://localhost:18080/")
					spider.Parse(req, spider.parse)
				}
				runtime.Gosched()
			}
		}
	}()

	if err := app.Wait(true); err != nil && !errors.Is(err, context.Canceled) {
		// We capture this for diagnostic purposes in the benchmark
	}

	duration := time.Since(startTime)
	rps := float64(spider.completed.Load()) / duration.Seconds()

	return rps
}

func main() {
	f, _ := os.Create("cpu.prof")
	defer f.Close()
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()

	fmt.Println("Warming up Mock Server on :18080...")
	mockServer()
	time.Sleep(500 * time.Millisecond)

	// handle signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Println("\n🛑 Tuning interrupted.")
		os.Exit(0)
	}()

	type permutation struct {
		Concurrency    string
		MaxIdle        string
		QueueBuf       string
		ResultHandlers string
	}

	cores := runtime.NumCPU()

	// perms represent different scaling profiles
	perms := []permutation{
		{fmt.Sprintf("%d", cores*40), "2000", "10000", fmt.Sprintf("%d", cores)},
		{fmt.Sprintf("%d", cores*40), "2000", "10000", fmt.Sprintf("%d", cores*2)},
		{fmt.Sprintf("%d", cores*60), "4000", "15000", fmt.Sprintf("%d", cores)},
		{fmt.Sprintf("%d", cores*60), "4000", "15000", fmt.Sprintf("%d", cores*2)},
	}

	bestRPS := 0.0
	var bestCombo permutation

	fmt.Println("Starting GoScrapy Auto-Tuning Benchmark Engine...")
	fmt.Printf("%-11s | %-10s | %-10s | %-12s | %-15s\n", "Concurrency", "MaxIdle", "QueueBuf", "ResHandlers", "Requests/Sec")
	fmt.Println("-------------------------------------------------------------------------------")

	for _, p := range perms {
		rps := runBenchmark(p.Concurrency, p.MaxIdle, p.QueueBuf, p.ResultHandlers)
		fmt.Printf("%-11s | %-10s | %-10s | %-12s | %.2f req/s\n", p.Concurrency, p.MaxIdle, p.QueueBuf, p.ResultHandlers, rps)

		if rps > bestRPS {
			bestRPS = rps
			bestCombo = p
		}
	}

	if bestCombo.Concurrency == "" {
		fmt.Println("\nNo results collected.")
		return
	}

	fmt.Println("\n--- TUNING COMPLETE ---")
	fmt.Printf("🏆 BEST SETUP (%.2f req/sec):\n\n", bestRPS)
	fmt.Println("Apply these exact fields to your settings.go or .env variables:")
	fmt.Printf("  SCHEDULER_CONCURRENCY                  = %s\n", bestCombo.Concurrency)
	fmt.Printf("  ENGINE_RESULT_HANDLERS                 = %s\n", bestCombo.ResultHandlers)
	fmt.Printf("  MIDDLEWARE_HTTP_MAX_IDLE_CONN          = %s\n", bestCombo.MaxIdle)
	fmt.Printf("  MIDDLEWARE_HTTP_MAX_CONN_PER_HOST      = %s\n", bestCombo.MaxIdle)
	fmt.Printf("  MIDDLEWARE_HTTP_MAX_IDLE_CONN_PER_HOST = %s\n", bestCombo.MaxIdle)
	fmt.Printf("  PIPELINEMANAGER_OUTPUT_QUEUE_BUF_SIZE  = %s\n", bestCombo.QueueBuf)
}
