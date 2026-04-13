// Note: generated tests
package middlewaremanager

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockTransport struct {
	called bool
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	m.called = true
	return &http.Response{StatusCode: http.StatusOK}, nil
}

func TestMiddlewareManager_Add(t *testing.T) {
	var capturedHeader string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedHeader = r.Header.Get("X-Test")
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	httpClient := &http.Client{}
	mm := New(httpClient)

	// Sample middleware that adds a header
	middleware := func(next http.RoundTripper) http.RoundTripper {
		return MiddlewareFunc(func(req *http.Request) (*http.Response, error) {
			req.Header.Set("X-Test", "middleware")
			if next == nil {
				return &http.Response{StatusCode: http.StatusTeapot, Header: make(http.Header)}, nil
			}
			return next.RoundTrip(req)
		})
	}

	mm.Add(middleware)

	// If httpClient.Transport was nil, our middleware handled it by returning Teapot
	req, _ := http.NewRequest("GET", ts.URL, nil)
	res, err := httpClient.Do(req)
	
	assert.NoError(t, err)
	assert.Equal(t, http.StatusTeapot, res.StatusCode)
	assert.Equal(t, "middleware", req.Header.Get("X-Test")) // This works locally because Teapot is returned before Do clones/dispatches

	// Now try with a real base transport
	httpClient.Transport = ts.Client().Transport
	mm.Add(middleware)
	
	req2, _ := http.NewRequest("GET", ts.URL, nil)
	res2, err := httpClient.Do(req2)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, res2.StatusCode)
	assert.Equal(t, "middleware", capturedHeader)
}
