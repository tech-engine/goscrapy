package gos

import (
	"context"
)

type cancelableSignal struct {
	cancel context.CancelFunc
	ctx    context.Context
}

func newCancelableSignal(ctx context.Context) *cancelableSignal {
	ctx, cancel := context.WithCancel(ctx)
	return &cancelableSignal{
		cancel: cancel,
		ctx:    ctx,
	}
}
