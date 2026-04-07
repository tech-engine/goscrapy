// Note: Tests generated using AI
package scheduler

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestRequest(ctx context.Context) *request {
	return &request{
		method: "GET",
		header: make(http.Header),
		ctx: ctx,
	}
}

func TestRequest_Method_TableDriven(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"lowercase method", "post", "POST"},
		{"already uppercase", http.MethodPut, http.MethodPut},
		{"mixed case", "pAtCh", "PATCH"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := newTestRequest(context.Background())
			req.Method(tt.input)
			assert.Equal(t, tt.expected, req.ReadMethod())
		})
	}
}

func TestRequest_Url(t *testing.T) {
	req := newTestRequest(context.Background())

	testUrl := "https://example.com/path?foo=bar"
	req.Url(testUrl)

	parsedUrl, err := url.Parse(testUrl)
	require.NoError(t, err)
	assert.Equal(t, parsedUrl, req.ReadUrl())
}

func TestRequest_Url_TableDriven(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		shouldPanic bool
	}{
		{"valid url", "https://example.com", false},
		{"invalid url", "://invalid\x7f", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := newTestRequest(context.Background())

			if tt.shouldPanic {
				assert.Panics(t, func() {
					req.Url(tt.input)
				})
			} else {
				req.Url(tt.input)
				parsed, err := url.Parse(tt.input)
				require.NoError(t, err)
				assert.Equal(t, parsed, req.ReadUrl())
			}
		})
	}
}

func TestRequest_Header(t *testing.T) {
	req := newTestRequest(context.Background())

	testHeader := http.Header{
		"Content-Type":  []string{"application/json"},
		"Authorization": []string{"Bearer xyz"},
	}
	req.Header(testHeader)
	assert.Equal(t, testHeader, req.ReadHeader())
}

func TestRequest_Body_TableDriven(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected string
	}{
		{"string", "hello", "hello"},
		{"bytes", []byte("byte data"), "byte data"},
		{"reader", strings.NewReader("reader data"), "reader data"},
		{"readcloser", io.NopCloser(strings.NewReader("rc data")), "rc data"},
		{"struct", struct {
			Name string `json:"name"`
		}{"Gopher"}, "Gopher"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := newTestRequest(context.Background())
			req.Body(tt.input)

			body, err := io.ReadAll(req.ReadBody())
			require.NoError(t, err)

			assert.Contains(t, string(body), tt.expected)
		})
	}
}

func TestRequest_Context(t *testing.T) {
	type ctxKeyType string
	ctx := context.WithValue(context.Background(), ctxKeyType("key"), "value")
	req := newTestRequest(ctx)
	assert.Equal(t, ctx, req.ReadContext())
}

func TestRequest_CookieJar_TableDriven(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty default", "", ""},
		{"set value", "session1", "session1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := newTestRequest(context.Background())

			if tt.input != "" {
				req.CookieJar(tt.input)
			}

			assert.Equal(t, tt.expected, req.ReadCookieJar())
		})
	}
}

func TestRequest_Meta_TableDriven(t *testing.T) {
	req := newTestRequest(context.Background())

	tests := []struct {
		key   string
		value any
	}{
		{"foo", "bar"},
		{"count", 42},
	}

	for _, tt := range tests {
		req.Meta(tt.key, tt.value)
	}

	require.NotNil(t, req.ReadMeta())

	for _, tt := range tests {
		val, exists := req.meta.Get(tt.key)
		assert.True(t, exists)
		assert.Equal(t, tt.value, val)
	}

	_, exists := req.meta.Get("missing")
	assert.False(t, exists)
}

func TestRequest_Reset(t *testing.T) {
	req := newTestRequest(context.Background())
	req.Method(http.MethodDelete)
	req.Url("https://example.com")
	req.Header(http.Header{"Cache-Control": []string{"no-cache"}})
	req.Body([]byte("data"))
	req.CookieJar("jar1")
	req.Meta("test", "data")

	req.Reset()

	assert.Equal(t, "", req.ReadMethod())
	assert.Nil(t, req.ReadUrl())
	assert.Empty(t, req.ReadHeader())
	assert.Nil(t, req.ReadBody())
	assert.Equal(t, "", req.ReadCookieJar())

	// Meta map is cleared, not nil'd
	if req.ReadMeta() != nil {
		_, exists := req.meta.Get("test")
		assert.False(t, exists, "meta should be cleared after Reset")
	}
}

func TestRequest_Chaining(t *testing.T) {
	req := newTestRequest(context.Background())

	// All writer methods should return the request itself for chaining
	result := req.
		Method("POST").
		Url("https://example.com").
		Header(http.Header{}).
		Body("data").
		CookieJar("jar").
		Meta("k", "v")

	assert.NotNil(t, result)
	assert.Equal(t, "POST", req.ReadMethod())
	assert.Equal(t, "jar", req.ReadCookieJar())
}
