package core

import (
	"bytes"
	"io"
	"net/http"

	executer "github.com/tech-engine/goscrapy/internal/executer/http"
)

func NewResponse() *Response {
	return &Response{}
}

func (r *Response) Body() io.ReadCloser {
	return r.body
}

func (r *Response) Bytes() []byte {
	buff := new(bytes.Buffer)
	buff.ReadFrom(r.body)
	return buff.Bytes()
}

func (r *Response) StatusCode() int {
	return r.statuscode
}

func (r *Response) Headers() http.Header {
	return r.headers
}

func (r *Response) Cookies() []*http.Cookie {
	return r.cookies
}

// setters
func (r *Response) SetStatusCode(statuscode int) executer.ResponseSetter {
	r.statuscode = statuscode
	return r
}

func (r *Response) SetBody(body io.ReadCloser) executer.ResponseSetter {
	r.body = body
	return r
}

func (r *Response) SetHeaders(headers http.Header) executer.ResponseSetter {
	r.headers = headers
	return r
}

func (r *Response) SetCookies(cookies []*http.Cookie) executer.ResponseSetter {
	r.cookies = cookies
	return r
}

func (r *Response) Reset() {
	r.statuscode = 0
	r.body = nil
	r.headers = nil
}
