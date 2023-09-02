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

type Manager[IN Job] interface {
	AddMiddlewares(...Middleware)
	NewJob(string) IN
	Run(IN)
	Start(context.Context) error
	Wait()
}

type manager[IN Job, OUT any] struct {
	wg           sync.WaitGroup
	spider       Spider[IN, OUT]
	Pipelines    *PipelineManager[IN, any, OUT, Output[IN, OUT]]
	ctx          context.Context
	requestPool  *rp.Pooler[Request]
	responsePool *rp.Pooler[Response]
	middlewares  []Middleware
	executer     *executer.Executer
	outputCh     chan Output[IN, OUT]
}

type PipelineManager[J Job, IN any, OUT any, OR Output[J, OUT]] struct {
	pipelines []Pipeline[J, IN, OUT, OR]
}

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
