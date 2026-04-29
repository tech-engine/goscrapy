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

// Represent an Request in goscrapy.
// Implements json.Marshaler, json.Unmarshaler, fmt.Stringer.
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

// Clone creates a copy of the request. The Body, Meta is shallow copied.
func (r *Request) Clone() *Request {
	nr := &Request{
		Ctx:          r.Ctx,
		Method_:      r.Method_,
		CookieJarKey: r.CookieJarKey,
		Body_:        r.Body_,
	}

	if r.URL != nil {
		u := *r.URL
		nr.URL = &u
	}

	if r.Header_ != nil {
		nr.Header_ = r.Header_.Clone()
	}

	if r.Meta_ != nil {
		nr.Meta_ = r.Meta_.Clone()
	}

	return nr
}

type requestData struct {
	URL          string              `json:"url"`
	Method       string              `json:"method"`
	Body         []byte              `json:"body,omitempty"`
	Header       map[string][]string `json:"header,omitempty"`
	Meta         map[string]any      `json:"meta,omitempty"`
	CookieJarKey string              `json:"cookie_jar_key,omitempty"`
}

// It safely reads and reconstructs the Request Body to prevent draining it during serialization.
func (r *Request) MarshalJSON() ([]byte, error) {
	data := requestData{
		Method:       r.Method_,
		Header:       r.Header_,
		CookieJarKey: r.CookieJarKey,
	}

	if r.URL != nil {
		data.URL = r.URL.String()
	}

	if r.Meta_ != nil {
		data.Meta = r.Meta_.ToMap()
	}

	if r.Body_ != nil {
		bodyBytes, err := io.ReadAll(r.Body_)
		r.Body_.Close() // close the original stream to prevent file descriptor leaks
		if err != nil {
			return nil, fmt.Errorf("failed to read request body for serialization: %w", err)
		}
		data.Body = bodyBytes
		// Restore the body stream so it is not destroyed for the framework
		r.Body_ = io.NopCloser(bytes.NewReader(bodyBytes))
	}

	return json.Marshal(data)
}

func (r *Request) UnmarshalJSON(data []byte) error {
	var rd requestData
	if err := json.Unmarshal(data, &rd); err != nil {
		return err
	}

	r.Method_ = rd.Method
	r.Header_ = rd.Header
	r.CookieJarKey = rd.CookieJarKey

	if rd.URL != "" {
		u, err := url.Parse(rd.URL)
		if err != nil {
			return fmt.Errorf("failed to parse url: %w", err)
		}
		r.URL = u
	}

	if len(rd.Body) > 0 {
		r.Body_ = io.NopCloser(bytes.NewReader(rd.Body))
	}

	if rd.Meta != nil {
		r.Meta_ = fsmap.New[string, any](24)
		for k, v := range rd.Meta {
			_ = r.Meta_.Set(k, v)
		}
	}

	return nil
}

func (r *Request) String() string {
	b, err := json.Marshal(r)
	if err != nil {
		return fmt.Sprintf("Request{Method: %s, URL: %v} (serialization error: %v)", r.Method_, r.URL, err)
	}
	return string(b)
}
