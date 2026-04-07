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

func (c *Core[OUT]) Request(req IRequestReader, cb ResponseCallback) {
	c.engine.Schedule(req, cb)
}

func (c *Core[OUT]) NewRequest(ctx context.Context) IRequestRW {
	return c.engine.NewRequest(ctx)
}

func (c *Core[OUT]) Yield(out IOutput[OUT]) {
	c.engine.Yield(out)
}
