// Note: AI generated tests.
package scheduler

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tech-engine/goscrapy/internal/fsmap"
)

func newTestResponse() *response {
	return NewResponse()
}

func TestResponse_Request(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "https://example.com", nil)
	require.NoError(t, err)

	resp := newTestResponse()
	resp.WriteRequest(req)

	assert.Equal(t, req, resp.Request())
}

func TestResponse_StatusCode_TableDriven(t *testing.T) {
	tests := []int{200, 404, 500, 0}

	for _, code := range tests {
		t.Run(http.StatusText(code), func(t *testing.T) {
			resp := newTestResponse()
			resp.WriteStatusCode(code)
			assert.Equal(t, code, resp.StatusCode())
		})
	}
}

func TestResponse_Header(t *testing.T) {
	resp := newTestResponse()

	header := http.Header{
		"Server":       {"nginx"},
		"Content-Type": {"text/html"},
	}

	resp.WriteHeader(header)
	assert.Equal(t, header, resp.Header())
}

func TestResponse_Body_TableDriven(t *testing.T) {
	tests := []struct {
		name     string
		body     io.ReadCloser
		expected string
	}{
		{"simple body", io.NopCloser(bytes.NewReader([]byte("hello"))), "hello"},
		{"empty body", io.NopCloser(bytes.NewReader([]byte{})), ""},
		{"html body", io.NopCloser(bytes.NewReader([]byte("<html></html>"))), "<html></html>"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := newTestResponse()
			resp.WriteBody(tt.body)

			data := resp.Bytes()
			assert.Equal(t, tt.expected, string(data))
		})
	}
}

func TestResponse_Cookies_TableDriven(t *testing.T) {
	tests := []struct {
		name     string
		cookies  []*http.Cookie
		expected int
	}{
		{"nil cookies", nil, 0},
		{"empty cookies", []*http.Cookie{}, 0},
		{
			"multiple cookies",
			[]*http.Cookie{
				{Name: "a", Value: "1"},
				{Name: "b", Value: "2"},
			},
			2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := newTestResponse()
			resp.WriteCookies(tt.cookies)

			got := resp.Cookies()
			assert.Len(t, got, tt.expected)
		})
	}
}

func TestResponse_Meta_TableDriven(t *testing.T) {
	resp := newTestResponse()

	tests := []struct {
		key   string
		value any
	}{
		{"foo", "bar"},
		{"count", 42},
	}

	meta := fsmap.New[string, any](10)
	resp.WriteMeta(meta)

	for _, tt := range tests {
		meta.Set(tt.key, tt.value)
	}

	for _, tt := range tests {
		val, exists := resp.Meta(tt.key)
		assert.True(t, exists)
		assert.Equal(t, tt.value, val)
	}

	_, exists := resp.Meta("missing")
	assert.False(t, exists)
}

func TestResponse_Bytes_Idempotent(t *testing.T) {
	resp := newTestResponse()
	content := "hello world"

	resp.WriteBody(io.NopCloser(bytes.NewReader([]byte(content))))

	first := resp.Bytes()
	second := resp.Bytes()

	assert.Equal(t, first, second, "Bytes() should be idempotent")
}

func TestResponse_Selectors_TableDriven(t *testing.T) {
	content := `<html><body>
		<div class="test">Hello World</div>
		<a href="link.html">Go</a>
	</body></html>`

	tests := []struct {
		name     string
		extract  func(r *response) []string
		expected []string
	}{
		{
			"css text",
			func(r *response) []string {
				return r.Css(".test").Text()
			},
			[]string{"Hello World"},
		},
		{
			"xpath attr",
			func(r *response) []string {
				return r.Xpath("//a").Attr("href")
			},
			[]string{"link.html"},
		},
		{
			"css anchor text",
			func(r *response) []string {
				return r.Css("a").Text()
			},
			[]string{"Go"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp := newTestResponse()
			resp.WriteBody(io.NopCloser(bytes.NewReader([]byte(content))))

			result := tt.extract(resp)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestResponse_Selectors_CacheReuse(t *testing.T) {
	resp := newTestResponse()
	content := `<html><body><div class="test">Hello</div><a>Go</a></body></html>`

	resp.WriteBody(io.NopCloser(bytes.NewReader([]byte(content))))

	// First call parses
	_ = resp.Css(".test").Text()

	// Second call should not fail even though body is consumed
	text := resp.Css("a").Text()

	require.Len(t, text, 1)
	assert.Equal(t, "Go", text[0])
}

func TestResponse_Reset(t *testing.T) {
	resp := newTestResponse()

	resp.WriteStatusCode(404)
	resp.WriteHeader(http.Header{"Cache": {"none"}})
	resp.WriteBody(io.NopCloser(bytes.NewReader([]byte("<html></html>"))))
	resp.WriteCookies([]*http.Cookie{{Name: "a", Value: "b"}})
	resp.WriteRequest(&http.Request{})
	resp.WriteMeta(fsmap.New[string, any](10))

	// Force parsing
	_ = resp.Css("html").Get()

	resp.Reset()

	assert.Equal(t, 0, resp.StatusCode())
	assert.Nil(t, resp.Body())
	assert.Nil(t, resp.Header())
	assert.Nil(t, resp.Cookies())
	assert.Nil(t, resp.Request())

	assert.Nil(t, resp.meta, "meta must be nil after Reset")
	assert.Nil(t, resp.nodes, "nodes must be nil after Reset")
}
