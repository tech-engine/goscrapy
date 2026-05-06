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

type permutation struct {
	MaxWorkers         string
	HTTPPoolSize       string
	PipelineBufferSize string
	ResultHandlerCount string
}

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

func runBenchmark(workers, httpPool, pipeBuf, resHandlers string) float64 {
	os.Setenv("AUTOSCALER_MAX_WORKERS", workers)
	os.Setenv("MIDDLEWARE_HTTP_MAX_IDLE_CONN", httpPool)
	os.Setenv("MIDDLEWARE_HTTP_MAX_CONN_PER_HOST", httpPool)
	os.Setenv("MIDDLEWARE_HTTP_MAX_IDLE_CONN_PER_HOST", httpPool)
	os.Setenv("PIPELINEMANAGER_OUTPUT_QUEUE_BUF_SIZE", pipeBuf)
	os.Setenv("ENGINE_RESULT_HANDLERS", resHandlers)

	// Use a timeout context to run each benchmark for a fixed duration
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Root logger for tuning, nullified to avoid interference
	l := logger.NewNoopLogger()

	// Initialize the application using the modern factory with NoopLogger
	cfg := gos.DefaultConfig()
	cfg.Logger = l
	app, err := gos.New[*Record](cfg)
	if err != nil {
		fmt.Printf("failed to create app: %v\n", err)
		return 0
	}
	app.WithPipelines(&NoopPipeline{})

	spider := &BenchSpider{
		ICoreSpider: app,
	}
	app.RegisterSpider(spider)

	// prime engine to avoid premature idle signal
	for range 100 {
		req := spider.Request(ctx)
		req.Url("http://localhost:18080/")
		spider.Parse(req, spider.parse)
	}

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
				for range 1000 {
					req := spider.Request(ctx)
					req.Url("http://localhost:18080/")
					spider.Parse(req, spider.parse)
				}
				runtime.Gosched()
			}
		}
	}()

	// block until benchmark window closes
	if err := app.Wait(); err != nil && !errors.Is(err, context.Canceled) {
		// We capture this for diagnostic purposes in the benchmark
	}

	duration := time.Since(startTime)
	completed := spider.completed.Load()
	rps := float64(completed) / duration.Seconds()

	return rps
}

func main() {
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

	cores := runtime.NumCPU()

	// Define ranges for automatic grid generation
	workerCounts := []int{cores * 20, cores * 80}
	httpPoolSizes := []int{2000, 5000}
	pipelineBufSizes := []int{10000, 30000, 60000, 100000}
	resHandlerCounts := []int{cores, cores * 4, cores * 8}

	// Ensure at least 1 handler
	for i, v := range resHandlerCounts {
		if v <= 0 {
			resHandlerCounts[i] = 1
		}
	}

	var perms []permutation
	for _, w := range workerCounts {
		for _, hp := range httpPoolSizes {
			for _, pb := range pipelineBufSizes {
				for _, rh := range resHandlerCounts {
					perms = append(perms, permutation{
						MaxWorkers:         fmt.Sprintf("%d", w),
						HTTPPoolSize:       fmt.Sprintf("%d", hp),
						PipelineBufferSize: fmt.Sprintf("%d", pb),
						ResultHandlerCount: fmt.Sprintf("%d", rh),
					})
				}
			}
		}
	}

	bestRPS := 0.0
	var bestCombo permutation

	fmt.Printf("Starting GoScrapy Auto-Tuning (%d permutations)...\n", len(perms))
	fmt.Printf("%-11s | %-12s | %-12s | %-12s | %-15s\n", "MaxWorkers", "HTTPPool", "PipeBuf", "ResHandlers", "Requests/Sec")
	fmt.Println("-------------------------------------------------------------------------------")

	for i, p := range perms {
		fmt.Printf("[%2d/%2d] Testing profile: W=%-4s, HTTP=%-4s, BUF=%-5s, RH=%-2s... ",
			i+1, len(perms), p.MaxWorkers, p.HTTPPoolSize, p.PipelineBufferSize, p.ResultHandlerCount)

		rps := runBenchmark(p.MaxWorkers, p.HTTPPoolSize, p.PipelineBufferSize, p.ResultHandlerCount)
		fmt.Printf("%.2f req/s\n", rps)

		if rps > bestRPS {
			bestRPS = rps
			bestCombo = p
		}
	}

	if bestCombo.MaxWorkers == "" {
		fmt.Println("\nNo results collected.")
		return
	}

	fmt.Println("\n--- TUNING COMPLETE ---")
	fmt.Printf("🏆 BEST SETUP (%.2f req/sec):\n\n", bestRPS)
	fmt.Println("Apply these exact fields to your settings.go or .env variables:")
	fmt.Printf("  AUTOSCALER_MAX_WORKERS                 = %s\n", bestCombo.MaxWorkers)
	fmt.Printf("  ENGINE_RESULT_HANDLERS                 = %s\n", bestCombo.ResultHandlerCount)
	fmt.Printf("  MIDDLEWARE_HTTP_MAX_IDLE_CONN          = %s\n", bestCombo.HTTPPoolSize)
	fmt.Printf("  MIDDLEWARE_HTTP_MAX_CONN_PER_HOST      = %s\n", bestCombo.HTTPPoolSize)
	fmt.Printf("  MIDDLEWARE_HTTP_MAX_IDLE_CONN_PER_HOST = %s\n", bestCombo.HTTPPoolSize)
	fmt.Printf("  PIPELINEMANAGER_OUTPUT_QUEUE_BUF_SIZE  = %s\n", bestCombo.PipelineBufferSize)
}
