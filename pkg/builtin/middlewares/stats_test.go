// Note: AI generated test file
package middlewares

import (
	"net/http"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockRoundTripper struct {
	response *http.Response
	err      error
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.response, m.err
}

func TestStats_Middleware(t *testing.T) {
	stats := NewStats()
	mock := &mockRoundTripper{
		response: &http.Response{StatusCode: 200},
	}

	middleware := Stats(stats)(mock)

	// Single request
	req, _ := http.NewRequest("GET", "https://example.com", nil)
	resp, err := middleware.RoundTrip(req)
	require.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	assert.Equal(t, uint64(1), stats.totalCount.Load())
	v, ok := stats.statusCodes.Load(200)
	assert.True(t, ok)
	assert.Equal(t, uint64(1), v.(*atomic.Uint64).Load())
}

func TestStats_Concurrent(t *testing.T) {
	stats := NewStats()
	mock200 := &mockRoundTripper{response: &http.Response{StatusCode: 200}}
	mock404 := &mockRoundTripper{response: &http.Response{StatusCode: 404}}

	mw200 := Stats(stats)(mock200)
	mw404 := Stats(stats)(mock404)

	var wg sync.WaitGroup
	const iterations = 50

	wg.Add(iterations * 2)
	for i := 0; i < iterations; i++ {
		go func() {
			defer wg.Done()
			req, _ := http.NewRequest("GET", "http://test.com", nil)
			mw200.RoundTrip(req)
		}()
		go func() {
			defer wg.Done()
			req, _ := http.NewRequest("GET", "http://test.com", nil)
			mw404.RoundTrip(req)
		}()
	}
	wg.Wait()

	assert.Equal(t, uint64(iterations*2), stats.totalCount.Load())

	v200, _ := stats.statusCodes.Load(200)
	assert.Equal(t, uint64(iterations), v200.(*atomic.Uint64).Load())

	v404, _ := stats.statusCodes.Load(404)
	assert.Equal(t, uint64(iterations), v404.(*atomic.Uint64).Load())
}

func TestStats_Print(t *testing.T) {
	stats := NewStats()
	// Should not panic on empty stats
	assert.NotPanics(t, func() {
		stats.Print()
	})

	stats.totalCount.Add(1)
	stats.statusCodes.Store(200, &atomic.Uint64{})
	stats.statusCodes.Store(404, &atomic.Uint64{})

	// ensure no panic with data
	assert.NotPanics(t, func() {
		stats.Print()
	})
}
