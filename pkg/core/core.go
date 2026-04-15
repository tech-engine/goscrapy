package core

import (
	"context"
)

type Core[OUT any] struct {
	engine IEngine[OUT]
}

func New[OUT any](engine IEngine[OUT]) *Core[OUT] {
	return &Core[OUT]{
		engine: engine,
	}
}

// Parse schedules a request and passed the processed response to the callback
func (c *Core[OUT]) Parse(req IRequestReader, cb ResponseCallback) {
	c.engine.Schedule(req, cb)
}

// Request creates a new request
func (c *Core[OUT]) Request(ctx context.Context) IRequestRW {
	return c.engine.NewRequest(ctx)
}

func (c *Core[OUT]) Yield(out IOutput[OUT]) {
	c.engine.Yield(out)
}
