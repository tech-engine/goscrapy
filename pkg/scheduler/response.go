package scheduler

import (
	"bytes"
	"io"
	"net/http"

	"github.com/tech-engine/goscrapy/internal/fsm"
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
	meta       *fsm.FixedSizeMap[string, any]
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
	// because we there isn't guarantee that we will have the same pair for req-res from the pools,
	// we must set it meta=nil upon releasing req-res to their respective pools, otherwise we will have corrupt data.
	r.meta = nil
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

func (r *response) WriteMeta(meta *fsm.FixedSizeMap[string, any]) {
	r.meta = meta
}
