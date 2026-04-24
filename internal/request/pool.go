package request

import (
	"context"
	"net/http"
	"sync"

	"github.com/tech-engine/goscrapy/internal/fsmap"
	"github.com/tech-engine/goscrapy/pkg/core"
)

type Pool struct {
	pool sync.Pool
}

func NewPool() *Pool {
	return &Pool{
		pool: sync.Pool{
			New: func() any {
				return &core.Request{}
			},
		},
	}
}

// fetches a clean request from the pool.
// Defaults: Method(GET)
func (p *Pool) Acquire(ctx context.Context) *core.Request {
	req := p.pool.Get().(*core.Request)
	req.Ctx = ctx
	req.Method = http.MethodGet
	if req.Header == nil {
		req.Header = make(http.Header)
	}
	if req.Meta == nil {
		req.Meta = fsmap.New[string, any](24)
	}
	return req
}

func (p *Pool) Release(req *core.Request) {
	if req == nil {
		return
	}
	req.Ctx = nil
	req.URL = nil
	req.Body = nil
	req.CookieJarKey = ""
	clear(req.Header)
	if req.Meta != nil {
		req.Meta.Clear()
	}
	p.pool.Put(req)
}
