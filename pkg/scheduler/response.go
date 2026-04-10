package scheduler

import (
	"bytes"
	"io"
	"net/http"

	"github.com/tech-engine/goscrapy/internal/fsmap"
	"github.com/tech-engine/goscrapy/pkg/core"
	"golang.org/x/net/html"
)

func NewResponse() *response {
	return &response{}
}

type response struct {
	statusCode int
	body       io.ReadCloser
	bodyBytes  []byte // cache body
	header     http.Header
	cookies    []*http.Cookie
	request    *http.Request
	meta       *fsmap.FixedSizeMap[string, any]
	nodes      Selectors
}

// response implementing core.ResponseReader
func (r *response) Request() *http.Request {
	return r.request
}

func (r *response) StatusCode() int {
	return r.statusCode
}

func (r *response) Body() io.ReadCloser {
	return r.body
}

func (r *response) Header() http.Header {
	return r.header
}

func (r *response) Cookies() []*http.Cookie {
	return r.cookies
}

func (r *response) Meta(key string) (any, bool) {
	return r.meta.Get(key)
}

func (r *response) Detach() core.IResponseReader {
	// ensure body is cached
	r.Bytes()

	dr := &response{
		statusCode: r.statusCode,
		bodyBytes:  r.bodyBytes,
		request:    r.request,
	}

	// clone body
	if dr.bodyBytes != nil {
		dr.body = io.NopCloser(bytes.NewReader(dr.bodyBytes))
	}

	// deep copy headers
	if r.header != nil {
		dr.header = r.header.Clone()
	}

	// deep copy cookies
	if r.cookies != nil {
		dr.cookies = make([]*http.Cookie, len(r.cookies))
		copy(dr.cookies, r.cookies)
	}

	// deep copy meta
	if r.meta != nil {
		dr.meta = r.meta.Clone()
	}

	// copy nodes
	dr.nodes = r.nodes

	return dr
}

func (r *response) Bytes() []byte {
	if r.bodyBytes != nil {
		return r.bodyBytes
	}

	if r.body == nil {
		return nil
	}

	data, err := io.ReadAll(r.body)

	if err != nil {
		return nil
	}

	r.bodyBytes = data

	r.body = io.NopCloser(bytes.NewReader(data))
	return data
}

func (r *response) Reset() {
	r.statusCode = 0
	r.body = nil
	// reuse the header map
	if r.header != nil {
		for key := range r.header {
			r.header.Del(key)
		}
	}
	// reuse the cookies slice by resetting length
	r.cookies = r.cookies[:0]
	r.request = nil
	r.bodyBytes = nil
	// because we there isn't guarantee that we will have the same pair for req-res from the pools,
	// we must set it meta=nil upon releasing req-res to their respective pools, otherwise we will have corrupt data.
	r.meta = nil
	r.nodes = nil
}

// response implementing engine.ResponseWriter
func (r *response) WriteRequest(request *http.Request) {
	r.request = request
}

func (r *response) WriteHeader(header http.Header) {
	r.header = header
}

func (r *response) WriteBody(body io.ReadCloser) {
	r.body = body
}

func (r *response) WriteStatusCode(statuscode int) {
	r.statusCode = statuscode
}

func (r *response) WriteCookies(cookies []*http.Cookie) {
	r.cookies = cookies
}

func (r *response) WriteMeta(meta *fsmap.FixedSizeMap[string, any]) {
	r.meta = meta
}

func (r *response) Css(selector string) core.ISelector {

	if r.nodes == nil {
		body := r.Bytes()
		if nodes, err := NewSelector(io.NopCloser(bytes.NewReader(body))); err == nil {
			r.nodes = nodes
		}
	}

	return r.nodes.Css(selector)
}

func (r *response) Xpath(xpath string) core.ISelector {

	if r.nodes == nil {
		body := r.Bytes()
		if nodes, err := NewSelector(io.NopCloser(bytes.NewReader(body))); err == nil {
			r.nodes = nodes
		}
	}
	return r.nodes.Xpath(xpath)
}

func (r *response) Text(def ...string) []string {
	return r.nodes.Text(def...)
}

func (r *response) Attr(attrName string) []string {
	return r.nodes.Attr(attrName)
}

func (r *response) Get() *html.Node {
	return r.nodes.Get()
}

func (r *response) GetAll() []*html.Node {
	return r.nodes.GetAll()
}
