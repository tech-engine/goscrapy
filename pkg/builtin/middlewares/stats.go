package middlewares

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/http/httptrace"
	"os"
	"sort"
	"sync"
	"sync/atomic"
	"text/tabwriter"
	"time"

	"github.com/tech-engine/goscrapy/pkg/middlewaremanager"
	ts "github.com/tech-engine/goscrapy/pkg/telemetry/stats"
)

const (
	MetricTLS     = "tls"
	MetricLatency = "latency"
)

type HttpMetrics struct {
	TotalRequests uint64
	TotalDuration time.Duration
	TotalBytes    uint64
	StatusCodes   map[int]uint64
	StartTime     time.Time
	Uptime        time.Duration
	AvgLatency    time.Duration
	AvgTLS        time.Duration
}

// For tracking scraping metrics
// implements ts.IStatsRecorder
type StatsCollector struct {
	totalCount    atomic.Uint64
	totalDuration atomic.Int64 // nanoseconds
	totalBytes    atomic.Uint64
	statusCodes   sync.Map // map[int]*atomic.Uint64
	startTime     time.Time

	// running sums for O(1) averages
	tlsSum  atomic.Int64 // nanoseconds
	ttfbSum atomic.Int64 // nanoseconds

	// counters for averages
	tlsCount  atomic.Uint64
	ttfbCount atomic.Uint64

	// to track each worker StatsCollector
	workerMu sync.Mutex
	workers  []*StatsCollector
}

type bodyCounter struct {
	io.ReadCloser
	recorder ts.IStatsRecorder
}

func (bc *bodyCounter) Read(p []byte) (n int, err error) {
	n, err = bc.ReadCloser.Read(p)
	bc.recorder.AddBytes(uint64(n))
	return n, err
}

func NewStatsCollector() *StatsCollector {
	return &StatsCollector{
		startTime: time.Now(),
	}
}

// Stats middleware tracks request counts, status codes, data usage and timing
func Stats(sc *StatsCollector) middlewaremanager.Middleware {
	return func(next http.RoundTripper) http.RoundTripper {
		return middlewaremanager.MiddlewareFunc(func(req *http.Request) (*http.Response, error) {
			// get recorder from context
			recorder := ts.FromContext(req.Context())
			if recorder == nil {
				recorder = sc
			}

			// track request count
			if s, ok := recorder.(*StatsCollector); ok {
				s.totalCount.Add(1)
			}

			var tlsStart, reqStart time.Time

			if recorder != nil {
				trace := &httptrace.ClientTrace{
					TLSHandshakeStart: func() { tlsStart = time.Now() },
					TLSHandshakeDone: func(_ tls.ConnectionState, err error) {
						if err == nil && !tlsStart.IsZero() {
							recorder.AddSample(MetricTLS, time.Since(tlsStart))
						}
					},
					WroteRequest: func(_ httptrace.WroteRequestInfo) {
						reqStart = time.Now()
					},
					GotFirstResponseByte: func() {
						if !reqStart.IsZero() {
							recorder.AddSample(MetricLatency, time.Since(reqStart))
						}
					},
				}
				req = req.WithContext(httptrace.WithClientTrace(req.Context(), trace))
			}
			start := time.Now()

			resp, err := next.RoundTrip(req)
			duration := time.Since(start)

			if s, ok := recorder.(*StatsCollector); ok {
				s.totalDuration.Add(int64(duration))
			}

			if err == nil && resp != nil {
				if s, ok := recorder.(*StatsCollector); ok {
					v, _ := s.statusCodes.LoadOrStore(resp.StatusCode, new(atomic.Uint64))
					v.(*atomic.Uint64).Add(1)
				}

				// wrap body to track data usage
				if resp.Body != nil {
					resp.Body = &bodyCounter{
						ReadCloser: resp.Body,
						recorder:   recorder,
					}
				}
			}

			return resp, err
		})
	}
}

func (s *StatsCollector) AddBytes(n uint64) {
	s.totalBytes.Add(n)
}

func (s *StatsCollector) AddSample(metric string, d time.Duration) {
	switch metric {
	case MetricTLS:
		s.tlsSum.Add(int64(d))
		s.tlsCount.Add(1)
	case MetricLatency:
		s.ttfbSum.Add(int64(d))
		s.ttfbCount.Add(1)
	}
}

func (s *StatsCollector) NewStatsRecorder() ts.IStatsRecorder {
	w := &StatsCollector{
		startTime: s.startTime,
	}
	s.workerMu.Lock()
	s.workers = append(s.workers, w)
	s.workerMu.Unlock()
	return w
}

func (s *StatsCollector) Name() string {
	return "http"
}

// Returns a point in time view of the current snapshot.
// It merges all active worker stats.
func (s *StatsCollector) Snapshot() ts.ComponentSnapshot {
	// to avoid any race conditions during iteration
	// we create a deep copy of the stats.
	snap := HttpMetrics{
		StartTime:   s.startTime,
		Uptime:      time.Since(s.startTime),
		StatusCodes: make(map[int]uint64),
	}

	// merge global counters
	snap.TotalRequests = s.totalCount.Load()
	snap.TotalDuration = time.Duration(s.totalDuration.Load())
	snap.TotalBytes = s.totalBytes.Load()

	s.statusCodes.Range(func(key, value any) bool {
		snap.StatusCodes[key.(int)] = value.(*atomic.Uint64).Load()
		return true
	})

	// merge worker counters without destroying them
	s.workerMu.Lock()
	var (
		aggTLSCount  uint64
		aggTLSSum    int64
		aggTTFBCount uint64
		aggTTFBSum   int64
	)
	for _, w := range s.workers {
		snap.TotalRequests += w.totalCount.Load()
		snap.TotalDuration += time.Duration(w.totalDuration.Load())
		snap.TotalBytes += w.totalBytes.Load()

		aggTLSCount += w.tlsCount.Load()
		aggTLSSum += w.tlsSum.Load()
		aggTTFBCount += w.ttfbCount.Load()
		aggTTFBSum += w.ttfbSum.Load()

		w.statusCodes.Range(func(key, value any) bool {
			snap.StatusCodes[key.(int)] += value.(*atomic.Uint64).Load()
			return true
		})
	}
	s.workerMu.Unlock()

	// calculate averages from combined running sums
	totalTLSCount := s.tlsCount.Load() + aggTLSCount
	totalTLSSum := s.tlsSum.Load() + aggTLSSum
	if totalTLSCount > 0 {
		snap.AvgTLS = time.Duration(totalTLSSum) / time.Duration(totalTLSCount)
	}

	totalTTFBCount := s.ttfbCount.Load() + aggTTFBCount
	totalTTFBSum := s.ttfbSum.Load() + aggTTFBSum
	if totalTTFBCount > 0 {
		snap.AvgLatency = time.Duration(totalTTFBSum) / time.Duration(totalTTFBCount)
	}

	return snap
}

func (s *StatsCollector) Print() {
	s.PrintTo(os.Stdout)
}

func (s *StatsCollector) PrintTo(out io.Writer) {
	if out == nil {
		out = os.Stdout
	}

	snap := s.Snapshot().(HttpMetrics)
	if snap.TotalRequests == 0 {
		return
	}

	avgDuration := snap.TotalDuration / time.Duration(snap.TotalRequests)

	w := tabwriter.NewWriter(out, 0, 0, 3, ' ', tabwriter.TabIndent)
	fmt.Fprintln(w, "\n---------------------- Stats ----------------------")
	fmt.Fprintf(w, "Start Time:\t%v\n", snap.StartTime.Format("2006-01-02 15:04:05"))
	fmt.Fprintf(w, "Elapsed Time:\t%v\n", snap.Uptime.Round(time.Millisecond))
	fmt.Fprintf(w, "Total Requests:\t%d\n", snap.TotalRequests)
	fmt.Fprintf(w, "Total Data:\t%.2f MB\n", float64(snap.TotalBytes)/(1024*1024))
	fmt.Fprintf(w, "Avg Time/URL:\t%v\n", avgDuration.Round(time.Millisecond))

	fmt.Fprintln(w, "\nHTTP Timing Stats (ms):")
	fmt.Fprintf(w, "  Avg Latency (TTFB):\t%.2f\n", float64(snap.AvgLatency.Microseconds())/1000)
	fmt.Fprintf(w, "  Avg TLS Handshake:\t%.2f\n", float64(snap.AvgTLS.Microseconds())/1000)

	fmt.Fprintln(w, "\nStatus Code Breakdown:")

	// sort status codes for consistent output
	var codes []int
	for code := range snap.StatusCodes {
		codes = append(codes, code)
	}
	sort.Ints(codes)

	for _, code := range codes {
		fmt.Fprintf(w, "  %d:\t%d\n", code, snap.StatusCodes[code])
	}
	fmt.Fprintln(w, "---------------------------------------------------")
	w.Flush()
}
