package signal

import (
	"context"
	"sync"

	"github.com/tech-engine/goscrapy/pkg/core"
)

// request signals emitted by worker pool and scheduler
type RequestBus interface {
	EmitRequestScheduled(ctx context.Context, req *core.Request)
	EmitRequestDropped(ctx context.Context, req *core.Request, err error)
	EmitRequestError(ctx context.Context, req *core.Request, err error)
	EmitResponseReceived(ctx context.Context, resp core.IResponseReader)
}

// item signals emitted by pipeline manager
type ItemBus[OUT any] interface {
	EmitItemScraped(ctx context.Context, item OUT)
	EmitItemDropped(ctx context.Context, item OUT, err error)
	EmitItemError(ctx context.Context, item OUT, err error)
}

// typed signal bus
type Bus[OUT any] struct {
	mu sync.RWMutex

	// spider signals
	spiderOpened []func(context.Context)
	spiderClosed []func(context.Context)
	spiderError  []func(context.Context, error)
	spiderIdle   []func(context.Context)

	// item signals
	itemScraped []func(context.Context, OUT)
	itemDropped []func(context.Context, OUT, error)
	itemError   []func(context.Context, OUT, error)

	// request signals
	requestScheduled []func(context.Context, *core.Request)
	requestDropped   []func(context.Context, *core.Request, error)
	requestError     []func(context.Context, *core.Request, error)

	// response signals
	responseReceived []func(context.Context, core.IResponseReader)

	// engine signals
	engineStarted []func(context.Context)
	engineStopped []func(context.Context)
}

// create new signal bus
func New[OUT any]() *Bus[OUT] {
	return &Bus[OUT]{}
}

func (b *Bus[OUT]) OnSpiderOpened(h func(context.Context)) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.spiderOpened = append(b.spiderOpened, h)
}

func (b *Bus[OUT]) OnSpiderClosed(h func(context.Context)) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.spiderClosed = append(b.spiderClosed, h)
}

func (b *Bus[OUT]) OnSpiderError(h func(context.Context, error)) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.spiderError = append(b.spiderError, h)
}

func (b *Bus[OUT]) OnSpiderIdle(h func(context.Context)) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.spiderIdle = append(b.spiderIdle, h)
}

func (b *Bus[OUT]) OnItemScraped(h func(context.Context, OUT)) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.itemScraped = append(b.itemScraped, h)
}

func (b *Bus[OUT]) OnItemDropped(h func(context.Context, OUT, error)) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.itemDropped = append(b.itemDropped, h)
}

func (b *Bus[OUT]) OnItemError(h func(context.Context, OUT, error)) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.itemError = append(b.itemError, h)
}

func (b *Bus[OUT]) OnRequestScheduled(h func(context.Context, *core.Request)) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.requestScheduled = append(b.requestScheduled, h)
}

func (b *Bus[OUT]) OnRequestDropped(h func(context.Context, *core.Request, error)) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.requestDropped = append(b.requestDropped, h)
}

func (b *Bus[OUT]) OnRequestError(h func(context.Context, *core.Request, error)) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.requestError = append(b.requestError, h)
}

func (b *Bus[OUT]) OnResponseReceived(h func(context.Context, core.IResponseReader)) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.responseReceived = append(b.responseReceived, h)
}

func (b *Bus[OUT]) OnEngineStarted(h func(context.Context)) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.engineStarted = append(b.engineStarted, h)
}

func (b *Bus[OUT]) OnEngineStopped(h func(context.Context)) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.engineStopped = append(b.engineStopped, h)
}

func (b *Bus[OUT]) EmitSpiderOpened(ctx context.Context) {
	b.mu.RLock()
	handlers := b.spiderOpened
	b.mu.RUnlock()
	for _, h := range handlers {
		h(ctx)
	}
}

func (b *Bus[OUT]) EmitSpiderClosed(ctx context.Context) {
	b.mu.RLock()
	handlers := b.spiderClosed
	b.mu.RUnlock()
	for _, h := range handlers {
		h(ctx)
	}
}

func (b *Bus[OUT]) EmitSpiderError(ctx context.Context, err error) {
	b.mu.RLock()
	handlers := b.spiderError
	b.mu.RUnlock()
	for _, h := range handlers {
		h(ctx, err)
	}
}

func (b *Bus[OUT]) EmitSpiderIdle(ctx context.Context) {
	b.mu.RLock()
	handlers := b.spiderIdle
	b.mu.RUnlock()
	for _, h := range handlers {
		h(ctx)
	}
}

func (b *Bus[OUT]) EmitItemScraped(ctx context.Context, item OUT) {
	b.mu.RLock()
	handlers := b.itemScraped
	b.mu.RUnlock()
	for _, h := range handlers {
		h(ctx, item)
	}
}

func (b *Bus[OUT]) EmitItemDropped(ctx context.Context, item OUT, err error) {
	b.mu.RLock()
	handlers := b.itemDropped
	b.mu.RUnlock()
	for _, h := range handlers {
		h(ctx, item, err)
	}
}

func (b *Bus[OUT]) EmitItemError(ctx context.Context, item OUT, err error) {
	b.mu.RLock()
	handlers := b.itemError
	b.mu.RUnlock()
	for _, h := range handlers {
		h(ctx, item, err)
	}
}

func (b *Bus[OUT]) EmitRequestScheduled(ctx context.Context, req *core.Request) {
	b.mu.RLock()
	handlers := b.requestScheduled
	b.mu.RUnlock()
	for _, h := range handlers {
		h(ctx, req)
	}
}

func (b *Bus[OUT]) EmitRequestDropped(ctx context.Context, req *core.Request, err error) {
	b.mu.RLock()
	handlers := b.requestDropped
	b.mu.RUnlock()
	for _, h := range handlers {
		h(ctx, req, err)
	}
}

func (b *Bus[OUT]) EmitRequestError(ctx context.Context, req *core.Request, err error) {
	b.mu.RLock()
	handlers := b.requestError
	b.mu.RUnlock()
	for _, h := range handlers {
		h(ctx, req, err)
	}
}

func (b *Bus[OUT]) EmitResponseReceived(ctx context.Context, resp core.IResponseReader) {
	b.mu.RLock()
	handlers := b.responseReceived
	b.mu.RUnlock()
	for _, h := range handlers {
		h(ctx, resp)
	}
}

func (b *Bus[OUT]) EmitEngineStarted(ctx context.Context) {
	b.mu.RLock()
	handlers := b.engineStarted
	b.mu.RUnlock()
	for _, h := range handlers {
		h(ctx)
	}
}

func (b *Bus[OUT]) EmitEngineStopped(ctx context.Context) {
	b.mu.RLock()
	handlers := b.engineStopped
	b.mu.RUnlock()
	for _, h := range handlers {
		h(ctx)
	}
}
