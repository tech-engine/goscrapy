package core

import (
	"context"
)

type Core[OUT any] struct {
	engine IEngine[OUT]
	pool   IRequestPool
}

func New[OUT any](engine IEngine[OUT], pool IRequestPool) *Core[OUT] {
	return &Core[OUT]{
		engine: engine,
		pool:   pool,
	}
}

// Parse schedules a request and passed the processed response to the callback
func (c *Core[OUT]) Parse(req *Request, cb ResponseCallback) {
	c.engine.Schedule(req, cb)
}

// Request creates a new request
func (c *Core[OUT]) Request(ctx context.Context) *Request {
	return c.pool.Acquire(ctx)
}

func (c *Core[OUT]) Yield(out IOutput[OUT]) {
	c.engine.Yield(out)
}
