package httpnative

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tech-engine/goscrapy/internal/fsmap"
)

type testCase struct {
	name,
	method string
	body     io.ReadCloser
	expected []byte
}

var testServer = httptest.NewServer(handler())

type testResponseWriter struct {
	statuscode int
	body       io.ReadCloser
}

func (r *testResponseWriter) WriteHeader(h http.Header) {
}

func (r *testResponseWriter) WriteBody(b io.ReadCloser) {
	r.body = b
}

func (r *testResponseWriter) WriteStatusCode(s int) {
	r.statuscode = s
}

func (r *testResponseWriter) WriteCookies(c []*http.Cookie) {
}

func (r *testResponseWriter) WriteRequest(req *http.Request) {
}

func (r *testResponseWriter) WriteMeta(m *fsmap.FixedSizeMap[string, any]) {
}

func handler() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET", "DELETE":
			// selectively sleep based of delay header to test context in request
			if delay := r.Header.Get("delay"); delay != "" {
				d, _ := strconv.Atoi(delay)
				time.Sleep(time.Duration(d) * time.Second)
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("ok"))
		case "POST", "PATCH", "PUT":
			w.WriteHeader(http.StatusOK)
			b, _ := io.ReadAll(r.Body)
			w.Write(b)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	return mux
}

func run(t *testing.T, adapter *httpAdapter, method string, body io.ReadCloser, expected []byte) {

	var err error
	resp := &testResponseWriter{}

	req, err := http.NewRequestWithContext(context.Background(), method, testServer.URL, body)
	assert.NoError(t, err)

	err = adapter.Do(resp, req)

	assert.NoError(t, err)

	if resp.body != nil {
		defer resp.body.Close()
	}

	assert.Equal(t, 200, resp.statuscode)

	var respB []byte
	if resp.body != nil {
		respB, _ = io.ReadAll(resp.body)
	}

	assert.Equalf(t, expected, respB, "expected %s, got %s", string(expected), string(respB))

}

func TestAdapterRequest(t *testing.T) {

	adapter := NewAdapter()
	testCases := []testCase{
		{
			name:     "GET",
			method:   "GET",
			expected: []byte("ok"),
		},
		{
			name:     "DELETE",
			method:   "DELETE",
			expected: []byte("ok"),
		},
		{
			name:     "POST",
			method:   "POST",
			body:     io.NopCloser(strings.NewReader("post")),
			expected: []byte("post"),
		},
		{
			name:     "PATCH",
			method:   "PATCH",
			body:     io.NopCloser(strings.NewReader("patch")),
			expected: []byte("patch"),
		},
		{
			name:     "PUT",
			method:   "PUT",
			body:     io.NopCloser(strings.NewReader("put")),
			expected: []byte("put"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.method, func(t *testing.T) {
			t.Parallel()
			run(t, adapter, tc.method, tc.body, tc.expected)
		})
	}
}

func TestAdapterRequestCtx(t *testing.T) {
	adapter := NewAdapter()

	resp := &testResponseWriter{}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", testServer.URL, nil)
	assert.NoError(t, err)

	req.Header.Add("delay", "3")

	err = adapter.Do(resp, req)

	assert.ErrorIs(t, err, context.DeadlineExceeded)
}
