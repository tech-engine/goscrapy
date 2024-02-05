package scheduler

import (
	"bytes"
	"io"
	"net/http"
)

func NewResponse() *response {
	return &response{}
}

type response struct {
	statusCode int
	body       io.ReadCloser
	header     http.Header
	cookies    []*http.Cookie
	request    *http.Request
	meta       map[string]any
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

func (r *response) Meta() map[string]any {
	return r.meta
}

func (r *response) Bytes() []byte {
	buff := new(bytes.Buffer)
	buff.ReadFrom(r.body)
	return buff.Bytes()
}

func (r *response) Reset() {
	r.statusCode = 0
	r.body = nil
	r.header = nil
	r.cookies = nil
	r.request = nil
	r.meta = nil
}

// response implementing core.ResponseWriter
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

func (r *response) WriteMeta(meta map[string]any) {
	r.meta = meta
}
