package middlewares

import (
	"net/http"
	"sync/atomic"
	"testing"
	"time"

)

func BenchmarkStatsCollector_Snapshot(b *testing.B) {
	s := NewStatsCollector()

	// Setup some mock data
	for i := 0; i < 1000; i++ {
		s.totalCount.Add(1)
		s.totalBytes.Add(1024)
		s.AddSample(MetricTLS, 10*time.Millisecond)
		s.AddSample(MetricLatency, 50*time.Millisecond)
		v, _ := s.statusCodes.LoadOrStore(200, new(atomic.Uint64))
		v.(*atomic.Uint64).Add(1)
	}

	// Add some workers
	for i := 0; i < 10; i++ {
		w := s.NewStatsRecorder().(*StatsCollector)
		w.totalCount.Add(100)
		w.AddSample(MetricLatency, 40*time.Millisecond)
	}

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_ = s.Snapshot()
	}
}

// Measures the overhead of the middleware itself
// during a simulated request cycle.
func BenchmarkStatsMiddleware_Impact(b *testing.B) {
	global := NewStatsCollector()
	mw := Stats(global)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	// Wrap handler inRoundTripper mock
	rt := &mockRT{handler: handler}
	instrumentedRT := mw(rt)

	req, _ := http.NewRequest("GET", "http://localhost/", nil)

	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		_, _ = instrumentedRT.RoundTrip(req)
	}
}

type mockRT struct {
	handler http.Handler
}

func (m *mockRT) RoundTrip(req *http.Request) (*http.Response, error) {
	// Minimal mock response
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Body:       http.NoBody,
	}, nil
}
