package core

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"sync"

	executer "github.com/tech-engine/goscrapy/internal/executer/http"
	rp "github.com/tech-engine/goscrapy/internal/resource_pool"
)

type Manager[IN Job, OUT any] interface {
	AddMiddlewares(...Middleware)
	AddPipeline(Pipeline[IN, any, OUT, Output[IN, OUT]], error) *pipeline[IN, any, OUT, Output[IN, OUT]]
	NewJob(string) IN
	Run(IN)
	Start(context.Context) error
	Wait()
}

type manager[IN Job, OUT any] struct {
	wg           sync.WaitGroup
	spider       Spider[IN, OUT]
	pipelines    *PipelineManager[IN, any, OUT, Output[IN, OUT]]
	ctx          context.Context
	requestPool  *rp.Pooler[Request]
	responsePool *rp.Pooler[Response]
	middlewares  []Middleware
	executer     *executer.Executer
	outputCh     chan Output[IN, OUT]
}

type pipeline[J Job, IN any, OUT any, OR Output[J, OUT]] struct {
	p               Pipeline[J, IN, OUT, OR]
	required, async bool
}

type PipelineManager[J Job, IN any, OUT any, OR Output[J, OUT]] struct {
	pipelines []*pipeline[J, IN, OUT, OR]
}

type PipelineOption[J Job, IN any, OUT any, OR Output[J, OUT]] func(pipeline[J, IN, OUT, OR])

type Request struct {
	url          *url.URL
	method       string
	body         io.ReadCloser
	headers      map[string]string
	meta         map[string]any
	cookieJarKey string
}

type Response struct {
	statuscode int
	body       io.ReadCloser
	headers    http.Header
	cookies    []*http.Cookie
}

type DelegatedOperator[IN Job, OUT any] struct {
	m *manager[IN, OUT]
}
