package httpnative

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tech-engine/goscrapy/internal/fsm"
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

func (r *testResponseWriter) WriteMeta(m *fsm.FixedSizeMap[string, any]) {
}

func handler() *http.ServeMux {
	mux := http.NewServeMux()
	// /get-cookie receives headers from client and set those headers as response cookies
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

func run(t *testing.T, adapter *HTTPAdapter, method string, body io.ReadCloser, expected []byte) {

	var err error
	resp := &testResponseWriter{}

	urlParsed, err := url.Parse(testServer.URL)

	assert.NoError(t, err)

	switch method {
	case "GET":
		err = adapter.Get(resp, urlParsed)
	case "DELETE":
		err = adapter.Delete(resp, urlParsed)
	case "POST":
		adapter.Body(body)
		err = adapter.Post(resp, urlParsed)
	case "PUT":
		adapter.Body(body)
		err = adapter.Put(resp, urlParsed)
	case "PATCH":
		adapter.Body(body)
		err = adapter.Patch(resp, urlParsed)
	default:
		err = adapter.Get(resp, urlParsed)
	}

	assert.NoError(t, err)

	defer resp.body.Close()

	assert.Equal(t, 200, resp.statuscode)

	respB, err := io.ReadAll(resp.body)

	assert.NoError(t, err)

	assert.Equalf(t, expected, respB, "expected %s, got %s", string(expected), string(respB))

}

func TestAdapterRequest(t *testing.T) {

	adapter := NewHTTPClientAdapter(&http.Client{})
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
		func(tc testCase) {
			t.Run(tc.method, func(t *testing.T) {
				t.Parallel()
				run(t, adapter, tc.method, tc.body, tc.expected)
			})
		}(tc)
	}
}

func TestAdapterRequestCtx(t *testing.T) {
	adapter := NewHTTPClientAdapter(&http.Client{})

	resp := &testResponseWriter{}

	urlParsed, err := url.Parse(testServer.URL)

	assert.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// added so that we can distinguise this request and sleep selectively for 3 seconds in our test server
	// which will cause the context to expire before we get a response from server
	headers := http.Header{}
	headers.Add("delay", "3")

	adapter.WithContext(ctx)
	adapter.Header(headers)

	err = adapter.Get(resp, urlParsed)

	assert.ErrorIs(t, err, context.DeadlineExceeded)
}
