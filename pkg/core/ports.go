package core

import (
	"context"
	"io"
	"net/http"
	"net/url"
)

type IEngine[OUT any] interface {
	Start(context.Context) error
	NewRequest() IRequestRW
	Schedule(IRequestReader, ResponseCallback)
	Yield(IOutput[OUT])
}

type IRequestReader interface {
	ReadContext() context.Context
	ReadUrl() *url.URL
	ReadHeader() http.Header
	ReadMethod() string
	ReadBody() io.ReadCloser
}

type IRequestWriter interface {
	WithContext(context.Context) IRequestWriter
	Url(string) IRequestWriter
	Header(http.Header) IRequestWriter
	Method(string) IRequestWriter
	Body(any) IRequestWriter
	MetaData(string, any) IRequestWriter
	CookieJar(string) IRequestWriter
}

type IRequestRW interface {
	IRequestReader
	IRequestWriter
	Reset()
}

type IResponseReader interface {
	Header() http.Header
	Body() io.ReadCloser
	Bytes() []byte
	StatusCode() int
	Cookies() []*http.Cookie
	Request() *http.Request
	Meta() map[string]any
}

type IJob interface {
	Id() string
}

type IOutput[OUT any] interface {
	Records() OUT
	RecordKeys() []string
	RecordsFlat() [][]any
	Error() error
	Job() IJob
	IsEmpty() bool
}

type ResponseCallback func(context.Context, IResponseReader)
