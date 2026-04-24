package core

import (
	"bytes"
	"context"
	"encoding/json"
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
	Method       string
	Body         io.ReadCloser
	Header       http.Header
	Meta         *fsmap.FixedSizeMap[string, any]
	CookieJarKey string
}

func (r *Request) WithContext(ctx context.Context) *Request {
	r.Ctx = ctx
	return r
}

func (r *Request) WithUrl(_url string) *Request {
	__url, err := url.Parse(_url)

	if err != nil {
		panic(fmt.Sprintf("SetUrl: error parsing url: %v", err))
	}

	r.URL = __url
	return r
}

func (r *Request) WithMethod(method string) *Request {
	r.Method = strings.ToUpper(method)
	return r
}

func (r *Request) WithBody(body any) *Request {
	switch v := body.(type) {
	case io.ReadCloser:
		r.Body = v
	case io.Reader:
		r.Body = io.NopCloser(v)
	case string:
		r.Body = io.NopCloser(strings.NewReader(v))
	case []byte:
		r.Body = io.NopCloser(bytes.NewReader(v))
	default:
		buf := new(bytes.Buffer)
		_ = json.NewEncoder(buf).Encode(v)
		r.Body = io.NopCloser(buf)
	}

	return r
}

func (r *Request) WithHeader(header http.Header) *Request {
	r.Header = header
	return r
}

func (r *Request) WithCookieJar(key string) *Request {
	r.CookieJarKey = key
	return r
}

func (r *Request) WithMeta(key string, val any) *Request {
	if r.Meta == nil {
		r.Meta = fsmap.New[string, any](24)
	}
	r.Meta.Set(key, val)
	return r
}
