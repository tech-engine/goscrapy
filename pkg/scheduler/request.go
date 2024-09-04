package scheduler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/tech-engine/goscrapy/internal/fsm"
	"github.com/tech-engine/goscrapy/pkg/core"
)

type request struct {
	ctx          context.Context
	url          *url.URL
	method       string
	body         io.ReadCloser
	header       http.Header
	meta         *fsm.FixedSizeMap[string, any]
	cookieJarKey string
}

// Request inplements core.IRequestReader
func (r *request) ReadMethod() string {
	return r.method
}

func (r *request) ReadUrl() *url.URL {
	return r.url
}

func (r *request) ReadHeader() http.Header {
	return r.header
}

func (r *request) ReadBody() io.ReadCloser {
	return r.body
}

func (r *request) ReadContext() context.Context {
	return r.ctx
}

// ReadMeta give us a shallow copy of meta.
func (r *request) ReadMeta() *fsm.FixedSizeMap[string, any] {
	return r.meta
}

// Read the cookie jar key associated with a request
func (r *request) ReadCookieJar() string {
	return r.cookieJarKey
}

// Request inplements core.IRequestWriter
func (r *request) Url(_url string) core.IRequestWriter {
	__url, err := url.Parse(_url)

	if err != nil {
		panic(fmt.Sprintf("SetUrl: error parsing url"))
	}

	r.url = __url
	return r
}

func (r *request) Method(method string) core.IRequestWriter {
	r.method = strings.ToUpper(method)
	return r
}

func (r *request) Body(body any) core.IRequestWriter {
	switch v := body.(type) {
	case io.Reader:
		r.body = io.NopCloser(v)
	case io.ReadCloser:
		r.body = v
	case string:
		r.body = io.NopCloser(strings.NewReader(v))
	case []byte:
		r.body = io.NopCloser(bytes.NewReader(v))
	default:
		var buf *bytes.Buffer
		_ = json.NewEncoder(buf).Encode(v)
		r.body = io.NopCloser(buf)
	}

	return r
}

func (r *request) Header(header http.Header) core.IRequestWriter {
	r.header = header
	return r
}

func (r *request) CookieJar(key string) core.IRequestWriter {
	r.cookieJarKey = key
	return r
}

// Pass meta data as key/value pair to be available in callback response.
func (r *request) Meta(key string, val any) core.IRequestWriter {
	if r.meta == nil {
		r.meta = fsm.New[string, any](24)
	}
	r.meta.Set(key, val)
	return r
}

func (r *request) WithContext(ctx context.Context) core.IRequestWriter {
	r.ctx = ctx
	return r
}

// func (r *request) MetaDataKey(key string) (any, bool) {
// 	if r.meta == nil {
// 		return nil, false
// 	}

// 	val, ok := r.meta[key]
// 	return val, ok
// }

func (r *request) Reset() {
	r.method = ""
	r.url = nil
	if r.header != nil {
		for key := range r.header {
			r.header.Del(key)
		}
	}
	r.body = nil
	r.cookieJarKey = ""
	if r.meta != nil {
		r.meta.Clear()
	}
}
