package core

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/tech-engine/goscrapy/internal/fsmap"
)

type Request struct {
	Ctx          context.Context
	URL          *url.URL
	Method_      string
	Body_        io.ReadCloser
	Header_      http.Header
	Meta_        *fsmap.FixedSizeMap[string, any]
	CookieJarKey string
}

func (r *Request) Context(ctx context.Context) *Request {
	r.Ctx = ctx
	return r
}

// Url sets the request URL. It accepts either a string or a *url.URL object.
func (r *Request) Url(u any) *Request {
	switch v := u.(type) {
	case string:
		parsed, err := url.Parse(v)
		if err != nil {
			panic(fmt.Sprintf("Request.Url: error parsing string: %v", err))
		}
		r.URL = parsed
	case *url.URL:
		r.URL = v
	default:
		panic(fmt.Sprintf("Request.Url: unsupported type %T", u))
	}
	return r
}

func (r *Request) Method(method string) *Request {
	r.Method_ = strings.ToUpper(method)
	return r
}

func (r *Request) Body(body any) *Request {
	switch v := body.(type) {
	case io.ReadCloser:
		r.Body_ = v
	case []byte:
		r.Body_ = io.NopCloser(bytes.NewReader(v))
	case string:
		r.Body_ = io.NopCloser(strings.NewReader(v))
	case io.Reader:
		r.Body_ = io.NopCloser(v)
	default:
		panic(fmt.Sprintf("Request.Body: unsupported type %T", body))
	}
	return r
}

func (r *Request) GetHeader(key string) string {
	if r.Header_ == nil {
		return ""
	}
	return r.Header_.Get(key)
}

func (r *Request) AddHeader(key, value string) *Request {
	if r.Header_ == nil {
		r.Header_ = make(http.Header)
	}
	r.Header_.Add(key, value)
	return r
}

func (r *Request) SetHeader(key, value string) *Request {
	if r.Header_ == nil {
		r.Header_ = make(http.Header)
	}
	r.Header_.Set(key, value)
	return r
}

func (r *Request) Header(header http.Header) *Request {
	r.Header_ = header
	return r
}

func (r *Request) CookieJar(key string) *Request {
	r.CookieJarKey = key
	return r
}

// Meta sets a value in the request metadata map. It initializes the map if it's nil.
func (r *Request) Meta(key string, val any) *Request {
	if r.Meta_ == nil {
		r.Meta_ = fsmap.New[string, any](24)
	}
	_ = r.Meta_.Set(key, val)
	return r
}
