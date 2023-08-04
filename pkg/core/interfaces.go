package core

import (
	"context"
	"net/http"

	metadata "github.com/tech-engine/goscrapy/pkg/meta_data"
)

type Delegator interface {
	ExRequest(ctx context.Context, req *Request, cb ResponseCallback)
	NewRequest() *Request
}

type ScraperCore[IN Job, OUT any] interface {
	StartRequest(context.Context, IN)
	Close()
	PullResult() Output[IN, OUT]
	NewJob(string) IN
}

type ScraperCoreUtility[IN Job, OUT any] interface {
	StartRequest(context.Context, IN)
}

type Scraper[IN Job, OUT any] interface {
	ScraperCore[IN, OUT]
	ScraperCoreUtility[IN, OUT]
	SetDelegator(Delegator)
}

type Output[IN Job, OUT any] interface {
	Records() OUT
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
}

type ResponseReader interface {
	Headers() http.Header
	Body() []byte
	StatusCode() int
}

type ResponseCallback func(context.Context, ResponseReader)

type CoreRequestProcessor func(context.Context, *Request, ResponseCallback)
