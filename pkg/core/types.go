package core

import (
	"context"
	"net/http"
	"sync"

	executer "github.com/tech-engine/goscrapy/internal/executer/http"
	rp "github.com/tech-engine/goscrapy/internal/resource_pool"
)

type manager[IN Job, OUT any] struct {
	wg           sync.WaitGroup
	scraper      Scraper[IN, OUT]
	Pipelines    *PipelineManager[IN, any, OUT, Output[IN, OUT]]
	ctx          context.Context
	requestPool  *rp.Pooler[Request]
	responsePool *rp.Pooler[Response]
	executer     *executer.Executer
}

type PipelineManager[J Job, IN any, OUT any, OR Output[J, OUT]] struct {
	pipelines []Pipeline[J, IN, OUT, OR]
}

type Request struct {
	url     string
	method  string
	body    any
	headers map[string]string
	meta    map[string]any
}

type Response struct {
	statuscode int
	body       []byte
	headers    http.Header
}

type DelegatedOperator[IN Job, OUT any] struct {
	m *manager[IN, OUT]
}
