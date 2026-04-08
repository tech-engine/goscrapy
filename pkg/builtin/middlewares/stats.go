package middlewares

import (
	"crypto/tls"
	"fmt"
	"io"
	"math/rand"
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
	maxSamples    = 10000
	MetricTLS     = "tls"
	MetricLatency = "latency"
)

// For tracking scraping metrics
type StatsCollector struct {
	totalCount    atomic.Uint64
	totalDuration atomic.Int64 // nanoseconds
	totalBytes    atomic.Uint64
	statusCodes   sync.Map // map[int]*atomic.Uint64
	startTime     time.Time

	// timing relted metrics
	metricsMu sync.Mutex
	tlsTimes  []time.Duration
	ttfbTimes []time.Duration

	// counters for reservoir sampling
	tlsCount  atomic.Uint64
	ttfbCount atomic.Uint64

	// to track each worker StatsCollector
	workerMu sync.Mutex
	workers  []*StatsCollector
}

func NewStats() *StatsCollector {
	return &StatsCollector{
		startTime: time.Now(),
	}
}

// Stats middleware tracks request counts, status codes, data usage and timing
func Stats(global *StatsCollector) middlewaremanager.Middleware {
	return func(next http.RoundTripper) http.RoundTripper {
		return middlewaremanager.MiddlewareFunc(func(req *http.Request) (*http.Response, error) {
			// get recorder from context
			recorder := ts.FromContext(req.Context())
			if recorder == nil {
				recorder = global
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

type bodyCounter struct {
	io.ReadCloser
	recorder ts.StatRecorder
}

func (bc *bodyCounter) Read(p []byte) (n int, err error) {
	n, err = bc.ReadCloser.Read(p)
	bc.recorder.AddBytes(uint64(n))
	return n, err
}

func (s *StatsCollector) AddBytes(n uint64) {
	s.totalBytes.Add(n)
}

func (s *StatsCollector) AddSample(metric string, d time.Duration) {
	switch metric {
	case MetricTLS:
		s.addSample(&s.tlsTimes, &s.tlsCount, d)
	case MetricLatency:
		s.addSample(&s.ttfbTimes, &s.ttfbCount, d)
	}
}

func (s *StatsCollector) addSample(target *[]time.Duration, count *atomic.Uint64, d time.Duration) {
	c := count.Add(1)
	s.metricsMu.Lock()
	defer s.metricsMu.Unlock()

	if len(*target) < maxSamples {
		*target = append(*target, d)
	} else {
		// sampling
		j := rand.Int63n(int64(c))
		if j < maxSamples {
			(*target)[j] = d
		}
	}
}

func (s *StatsCollector) NewWorkerCollector() *StatsCollector {
	w := &StatsCollector{
		startTime: s.startTime,
	}
	s.workerMu.Lock()
	s.workers = append(s.workers, w)
	s.workerMu.Unlock()
	return w
}

func (s *StatsCollector) Merge(other *StatsCollector) {
	s.totalCount.Add(other.totalCount.Load())
	s.totalDuration.Add(other.totalDuration.Load())
	s.totalBytes.Add(other.totalBytes.Load())

	other.statusCodes.Range(func(key, value any) bool {
		code := key.(int)
		count := value.(*atomic.Uint64).Load()
		v, _ := s.statusCodes.LoadOrStore(code, new(atomic.Uint64))
		v.(*atomic.Uint64).Add(count)
		return true
	})

	s.metricsMu.Lock()
	defer s.metricsMu.Unlock()
	s.mergeSamples(&s.tlsTimes, other.tlsTimes)
	s.mergeSamples(&s.ttfbTimes, other.ttfbTimes)
}

func (s *StatsCollector) mergeSamples(target *[]time.Duration, source []time.Duration) {
	*target = append(*target, source...)
	if len(*target) > maxSamples {
		// Reservoir merge would be better, but simple trim is fine for now
		*target = (*target)[:maxSamples]
	}
}

func (s *StatsCollector) Print() {
	// merge worker stats
	s.workerMu.Lock()
	for _, w := range s.workers {
		s.Merge(w)
	}
	s.workers = nil // reset workers after merge
	s.workerMu.Unlock()

	total := s.totalCount.Load()
	if total == 0 {
		return
	}

	totalDuration := time.Duration(s.totalDuration.Load())
	avgDuration := totalDuration / time.Duration(total)
	elapsed := time.Since(s.startTime)
	totalMB := float64(s.totalBytes.Load()) / (1024 * 1024)

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', tabwriter.TabIndent)
	fmt.Fprintln(w, "\n---------------------- Stats ----------------------")
	fmt.Fprintf(w, "Start Time:\t%v\n", s.startTime.Format("2006-01-02 15:04:05"))
	fmt.Fprintf(w, "Elapsed Time:\t%v\n", elapsed.Round(time.Millisecond))
	fmt.Fprintf(w, "Total Requests:\t%d\n", total)
	fmt.Fprintf(w, "Total Data:\t%.2f MB\n", totalMB)
	fmt.Fprintf(w, "Avg Time/URL:\t%v\n", avgDuration.Round(time.Millisecond))

	fmt.Fprintln(w, "\nHTTP Timing Stats (ms):")
	fmt.Fprintln(w, "  Metric\tMin\tMax\tMean\tMedian")
	s.printTiming(w, "  TLS", s.tlsTimes)
	s.printTiming(w, "  Latency", s.ttfbTimes)

	fmt.Fprintln(w, "\nStatus Code Breakdown:")

	// Sort status codes for consistent output
	var codes []int
	s.statusCodes.Range(func(key, value any) bool {
		codes = append(codes, key.(int))
		return true
	})
	sort.Ints(codes)

	for _, code := range codes {
		v, _ := s.statusCodes.Load(code)
		count := v.(*atomic.Uint64).Load()
		fmt.Fprintf(w, "  %d:\t%d\n", code, count)
	}
	fmt.Fprintln(w, "---------------------------------------------------")
	w.Flush()
}

func (s *StatsCollector) printTiming(w io.Writer, name string, times []time.Duration) {
	if len(times) == 0 {
		fmt.Fprintf(w, "%s:\t-\t-\t-\t-\n", name)
		return
	}

	// copy samples to avoid lock contention
	s.metricsMu.Lock()
	sorted := make([]time.Duration, len(times))
	copy(sorted, times)
	s.metricsMu.Unlock()

	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })

	min := sorted[0]
	max := sorted[len(sorted)-1]
	var sum time.Duration
	for _, t := range sorted {
		sum += t
	}
	mean := sum / time.Duration(len(sorted))
	median := sorted[len(sorted)/2]

	fmt.Fprintf(w, "%s:\t%.2f\t%.2f\t%.2f\t%.2f\n",
		name,
		float64(min.Microseconds())/1000,
		float64(max.Microseconds())/1000,
		float64(mean.Microseconds())/1000,
		float64(median.Microseconds())/1000,
	)
}
