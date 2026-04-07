package middlewares

import (
	"fmt"
	"net/http"
	"os"
	"sort"
	"sync"
	"sync/atomic"
	"text/tabwriter"
	"time"

	"github.com/tech-engine/goscrapy/pkg/middlewaremanager"
)

// StatsCollector tracks scraping metrics
type StatsCollector struct {
	totalCount    atomic.Uint64
	totalDuration atomic.Int64 // nanoseconds
	statusCodes   sync.Map     // map[int]*atomic.Uint64
	startTime     time.Time
}

// NewStats creates a new stats collector
func NewStats() *StatsCollector {
	return &StatsCollector{
		startTime: time.Now(),
	}
}

// Stats middleware tracks request counts and status codes
func Stats(s *StatsCollector) middlewaremanager.Middleware {
	return func(next http.RoundTripper) http.RoundTripper {
		return middlewaremanager.MiddlewareFunc(func(req *http.Request) (*http.Response, error) {
			s.totalCount.Add(1)
			start := time.Now()

			resp, err := next.RoundTrip(req)
			duration := time.Since(start)
			s.totalDuration.Add(int64(duration))

			if err == nil && resp != nil {
				v, _ := s.statusCodes.LoadOrStore(resp.StatusCode, new(atomic.Uint64))
				v.(*atomic.Uint64).Add(1)
			}

			return resp, err
		})
	}
}

// Print displays a summary of the collected statistics
func (s *StatsCollector) Print() {
	total := s.totalCount.Load()
	if total == 0 {
		return
	}

	totalDuration := time.Duration(s.totalDuration.Load())
	avgDuration := totalDuration / time.Duration(total)
	elapsed := time.Since(s.startTime)

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', tabwriter.TabIndent)
	fmt.Fprintln(w, "\n---------------------- Stats ----------------------")
	fmt.Fprintf(w, "Start Time:\t%v\n", s.startTime.Format("2006-01-02 15:04:05"))
	fmt.Fprintf(w, "Elapsed Time:\t%v\n", elapsed.Round(time.Millisecond))
	fmt.Fprintf(w, "Total Requests:\t%d\n", total)
	fmt.Fprintf(w, "Avg Time/URL:\t%v\n", avgDuration.Round(time.Millisecond))
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
