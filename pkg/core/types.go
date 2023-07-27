package core

import (
	"context"
	"net/http"
	"sync"
	"time"

	executer "github.com/tech-engine/goscrapy/internal/executer/http"
	rp "github.com/tech-engine/goscrapy/internal/resource_pool"
)

type manager[IN Job, OUT any] struct {
	mu           *sync.RWMutex
	scraper      Scraper[IN, OUT]
	Jobs         map[string]ManagerJob[IN]
	Pipelines    *PipelineManager[IN, any, OUT, Output[IN, OUT]]
	ctx          context.Context
	requestPool  *rp.Pooler[Request]
	responsePool *rp.Pooler[Response]
	executer     *executer.Executer
}

type ManagerJob[IN Job] struct {
	id                string
	name              string
	startTime         time.Time
	lastCursor        string
	maxResultsAllowed uint64
	status            string
	ctx               context.Context
	cancel            context.CancelFunc
	ScraperJob        IN
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
