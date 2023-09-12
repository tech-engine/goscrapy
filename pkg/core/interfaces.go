package core

import (
	"context"
	"io"
	"net/http"

	metadata "github.com/tech-engine/goscrapy/pkg/meta_data"
)

type Delegator[IN Job, OUT any] interface {
	ExRequest(ctx context.Context, req *Request, cb ResponseCallback)
	NewRequest() *Request
	Yield(Output[IN, OUT])
}

type SpiderCore[IN Job, OUT any] interface {
	StartRequest(context.Context, IN)
	Close(context.Context)
	NewJob(string) IN
}

type SpiderCoreUtility[IN Job, OUT any] interface {
	StartRequest(context.Context, IN)
}

type Spider[IN Job, OUT any] interface {
	SpiderCore[IN, OUT]
	SpiderCoreUtility[IN, OUT]
	SetDelegator(Delegator[IN, OUT])
}

type Output[IN Job, OUT any] interface {
	Records() OUT
	RecordKeys() []string
	RecordsFlat() [][]any
	Error() error
	Job() IN
	IsEmpty() bool
}

type Job interface {
	Id() string
}

type Pipeline[J Job, IN any, OUT any, OR Output[J, OUT]] interface {
	Open(context.Context) error
	Close()
	ProcessItem(IN, OR, metadata.MetaData) (IN, error)
}

type RequestWriter interface {
	SetUrl(string) RequestWriter
	SetHeaders(map[string]string) RequestWriter
	SetMethod(string) RequestWriter
	SetBody(any) RequestWriter
	SetMetaData(string, any) RequestWriter
	SetCookieJar(string) RequestWriter
}

type ResponseReader interface {
	Headers() http.Header
	Body() io.ReadCloser
	Bytes() []byte
	StatusCode() int
	Cookies() []*http.Cookie
}

type ResponseCallback func(context.Context, ResponseReader)

type CoreRequestProcessor func(context.Context, *Request, ResponseCallback)
