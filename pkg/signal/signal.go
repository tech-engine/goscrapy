package signal

import (
	"context"
	"sync"
	"sync/atomic"

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
	mu sync.Mutex

	// spider signals
	spiderOpened atomic.Pointer[[]func(context.Context)]
	spiderClosed atomic.Pointer[[]func(context.Context)]
	spiderError  atomic.Pointer[[]func(context.Context, error)]
	spiderIdle   atomic.Pointer[[]func(context.Context)]

	// item signals
	itemScraped atomic.Pointer[[]func(context.Context, OUT)]
	itemDropped atomic.Pointer[[]func(context.Context, OUT, error)]
	itemError   atomic.Pointer[[]func(context.Context, OUT, error)]

	// request signals
	requestScheduled atomic.Pointer[[]func(context.Context, *core.Request)]
	requestDropped   atomic.Pointer[[]func(context.Context, *core.Request, error)]
	requestError     atomic.Pointer[[]func(context.Context, *core.Request, error)]

	// response signals
	responseReceived atomic.Pointer[[]func(context.Context, core.IResponseReader)]

	// engine signals
	engineStarted atomic.Pointer[[]func(context.Context)]
	engineStopped atomic.Pointer[[]func(context.Context)]
}

// create new signal bus
func New[OUT any]() *Bus[OUT] {
	return &Bus[OUT]{}
}

func (b *Bus[OUT]) OnSpiderOpened(h func(context.Context)) {
	b.mu.Lock()
	defer b.mu.Unlock()
	old := b.spiderOpened.Load()
	var newHandlers []func(context.Context)
	if old != nil {
		newHandlers = append([]func(context.Context){}, *old...)
	}
	newHandlers = append(newHandlers, h)
	b.spiderOpened.Store(&newHandlers)
}

func (b *Bus[OUT]) OnSpiderClosed(h func(context.Context)) {
	b.mu.Lock()
	defer b.mu.Unlock()
	old := b.spiderClosed.Load()
	var newHandlers []func(context.Context)
	if old != nil {
		newHandlers = append([]func(context.Context){}, *old...)
	}
	newHandlers = append(newHandlers, h)
	b.spiderClosed.Store(&newHandlers)
}

func (b *Bus[OUT]) OnSpiderError(h func(context.Context, error)) {
	b.mu.Lock()
	defer b.mu.Unlock()
	old := b.spiderError.Load()
	var newHandlers []func(context.Context, error)
	if old != nil {
		newHandlers = append([]func(context.Context, error){}, *old...)
	}
	newHandlers = append(newHandlers, h)
	b.spiderError.Store(&newHandlers)
}

func (b *Bus[OUT]) OnSpiderIdle(h func(context.Context)) {
	b.mu.Lock()
	defer b.mu.Unlock()
	old := b.spiderIdle.Load()
	var newHandlers []func(context.Context)
	if old != nil {
		newHandlers = append([]func(context.Context){}, *old...)
	}
	newHandlers = append(newHandlers, h)
	b.spiderIdle.Store(&newHandlers)
}

func (b *Bus[OUT]) OnItemScraped(h func(context.Context, OUT)) {
	b.mu.Lock()
	defer b.mu.Unlock()
	old := b.itemScraped.Load()
	var newHandlers []func(context.Context, OUT)
	if old != nil {
		newHandlers = append([]func(context.Context, OUT){}, *old...)
	}
	newHandlers = append(newHandlers, h)
	b.itemScraped.Store(&newHandlers)
}

func (b *Bus[OUT]) OnItemDropped(h func(context.Context, OUT, error)) {
	b.mu.Lock()
	defer b.mu.Unlock()
	old := b.itemDropped.Load()
	var newHandlers []func(context.Context, OUT, error)
	if old != nil {
		newHandlers = append([]func(context.Context, OUT, error){}, *old...)
	}
	newHandlers = append(newHandlers, h)
	b.itemDropped.Store(&newHandlers)
}

func (b *Bus[OUT]) OnItemError(h func(context.Context, OUT, error)) {
	b.mu.Lock()
	defer b.mu.Unlock()
	old := b.itemError.Load()
	var newHandlers []func(context.Context, OUT, error)
	if old != nil {
		newHandlers = append([]func(context.Context, OUT, error){}, *old...)
	}
	newHandlers = append(newHandlers, h)
	b.itemError.Store(&newHandlers)
}

func (b *Bus[OUT]) OnRequestScheduled(h func(context.Context, *core.Request)) {
	b.mu.Lock()
	defer b.mu.Unlock()
	old := b.requestScheduled.Load()
	var newHandlers []func(context.Context, *core.Request)
	if old != nil {
		newHandlers = append([]func(context.Context, *core.Request){}, *old...)
	}
	newHandlers = append(newHandlers, h)
	b.requestScheduled.Store(&newHandlers)
}

func (b *Bus[OUT]) OnRequestDropped(h func(context.Context, *core.Request, error)) {
	b.mu.Lock()
	defer b.mu.Unlock()
	old := b.requestDropped.Load()
	var newHandlers []func(context.Context, *core.Request, error)
	if old != nil {
		newHandlers = append([]func(context.Context, *core.Request, error){}, *old...)
	}
	newHandlers = append(newHandlers, h)
	b.requestDropped.Store(&newHandlers)
}

func (b *Bus[OUT]) OnRequestError(h func(context.Context, *core.Request, error)) {
	b.mu.Lock()
	defer b.mu.Unlock()
	old := b.requestError.Load()
	var newHandlers []func(context.Context, *core.Request, error)
	if old != nil {
		newHandlers = append([]func(context.Context, *core.Request, error){}, *old...)
	}
	newHandlers = append(newHandlers, h)
	b.requestError.Store(&newHandlers)
}

func (b *Bus[OUT]) OnResponseReceived(h func(context.Context, core.IResponseReader)) {
	b.mu.Lock()
	defer b.mu.Unlock()
	old := b.responseReceived.Load()
	var newHandlers []func(context.Context, core.IResponseReader)
	if old != nil {
		newHandlers = append([]func(context.Context, core.IResponseReader){}, *old...)
	}
	newHandlers = append(newHandlers, h)
	b.responseReceived.Store(&newHandlers)
}

func (b *Bus[OUT]) OnEngineStarted(h func(context.Context)) {
	b.mu.Lock()
	defer b.mu.Unlock()
	old := b.engineStarted.Load()
	var newHandlers []func(context.Context)
	if old != nil {
		newHandlers = append([]func(context.Context){}, *old...)
	}
	newHandlers = append(newHandlers, h)
	b.engineStarted.Store(&newHandlers)
}

func (b *Bus[OUT]) OnEngineStopped(h func(context.Context)) {
	b.mu.Lock()
	defer b.mu.Unlock()
	old := b.engineStopped.Load()
	var newHandlers []func(context.Context)
	if old != nil {
		newHandlers = append([]func(context.Context){}, *old...)
	}
	newHandlers = append(newHandlers, h)
	b.engineStopped.Store(&newHandlers)
}

func (b *Bus[OUT]) EmitSpiderOpened(ctx context.Context) {
	if handlers := b.spiderOpened.Load(); handlers != nil {
		for _, h := range *handlers {
			h(ctx)
		}
	}
}

func (b *Bus[OUT]) EmitSpiderClosed(ctx context.Context) {
	if handlers := b.spiderClosed.Load(); handlers != nil {
		for _, h := range *handlers {
			h(ctx)
		}
	}
}

func (b *Bus[OUT]) EmitSpiderError(ctx context.Context, err error) {
	if handlers := b.spiderError.Load(); handlers != nil {
		for _, h := range *handlers {
			h(ctx, err)
		}
	}
}

func (b *Bus[OUT]) EmitSpiderIdle(ctx context.Context) {
	if handlers := b.spiderIdle.Load(); handlers != nil {
		for _, h := range *handlers {
			h(ctx)
		}
	}
}

func (b *Bus[OUT]) EmitItemScraped(ctx context.Context, item OUT) {
	if handlers := b.itemScraped.Load(); handlers != nil {
		for _, h := range *handlers {
			h(ctx, item)
		}
	}
}

func (b *Bus[OUT]) EmitItemDropped(ctx context.Context, item OUT, err error) {
	if handlers := b.itemDropped.Load(); handlers != nil {
		for _, h := range *handlers {
			h(ctx, item, err)
		}
	}
}

func (b *Bus[OUT]) EmitItemError(ctx context.Context, item OUT, err error) {
	if handlers := b.itemError.Load(); handlers != nil {
		for _, h := range *handlers {
			h(ctx, item, err)
		}
	}
}

func (b *Bus[OUT]) EmitRequestScheduled(ctx context.Context, req *core.Request) {
	if handlers := b.requestScheduled.Load(); handlers != nil {
		for _, h := range *handlers {
			h(ctx, req)
		}
	}
}

func (b *Bus[OUT]) EmitRequestDropped(ctx context.Context, req *core.Request, err error) {
	if handlers := b.requestDropped.Load(); handlers != nil {
		for _, h := range *handlers {
			h(ctx, req, err)
		}
	}
}

func (b *Bus[OUT]) EmitRequestError(ctx context.Context, req *core.Request, err error) {
	if handlers := b.requestError.Load(); handlers != nil {
		for _, h := range *handlers {
			h(ctx, req, err)
		}
	}
}

func (b *Bus[OUT]) EmitResponseReceived(ctx context.Context, resp core.IResponseReader) {
	if handlers := b.responseReceived.Load(); handlers != nil {
		for _, h := range *handlers {
			h(ctx, resp)
		}
	}
}

func (b *Bus[OUT]) EmitEngineStarted(ctx context.Context) {
	if handlers := b.engineStarted.Load(); handlers != nil {
		for _, h := range *handlers {
			h(ctx)
		}
	}
}

func (b *Bus[OUT]) EmitEngineStopped(ctx context.Context) {
	if handlers := b.engineStopped.Load(); handlers != nil {
		for _, h := range *handlers {
			h(ctx)
		}
	}
}
