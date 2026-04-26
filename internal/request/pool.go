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
	req.Method_ = http.MethodGet
	if req.Header_ == nil {
		req.Header_ = make(http.Header)
	}
	if req.Meta_ == nil {
		req.Meta_ = fsmap.New[string, any](24)
	}
	return req
}

func (p *Pool) Release(req *core.Request) {
	if req == nil {
		return
	}
	req.Ctx = nil
	req.URL = nil
	req.Body_ = nil
	req.CookieJarKey = ""
	clear(req.Header_)
	if req.Meta_ != nil {
		req.Meta_.Clear()
	}
	p.pool.Put(req)
}
