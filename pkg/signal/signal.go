package signal

import (
	"context"
	"sync"

	"github.com/tech-engine/goscrapy/pkg/core"
)

// Type defines the goscraoy signal type
type Type string

const (
	// spider signals
	SpiderOpened Type = "spider_opened"
	SpiderClosed Type = "spider_closed"
	SpiderError  Type = "spider_error"
	SpiderIdle   Type = "spider_idle"

	// item signals
	ItemScraped Type = "item_scraped"
	ItemDropped Type = "item_dropped"
	ItemError   Type = "item_error"

	// request signals
	RequestScheduled Type = "request_scheduled"
	RequestDropped   Type = "request_dropped"
	RequestError     Type = "request_error"

	// response signals
	ResponseReceived Type = "response_received"

	// rngine signals
	EngineStarted Type = "engine_started"
	EngineStopped Type = "engine_stopped"
)

type Bus struct {
	mu       sync.RWMutex
	handlers map[Type][]any
}

// Returns a new signal bus
func New() *Bus {
	return &Bus{
		handlers: make(map[Type][]any),
	}
}

// Connect attaches a handler to a signal
func (b *Bus) Connect(sig Type, h any) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.handlers[sig] = append(b.handlers[sig], h)
}

// spider signals
func (b *Bus) EmitSpiderOpened(ctx context.Context) {
	b.mu.RLock()
	handlers := b.handlers[SpiderOpened]
	b.mu.RUnlock()
	for _, h := range handlers {
		if fn, ok := h.(func(context.Context)); ok {
			fn(ctx)
		}
	}
}

func (b *Bus) EmitSpiderClosed(ctx context.Context) {
	b.mu.RLock()
	handlers := b.handlers[SpiderClosed]
	b.mu.RUnlock()
	for _, h := range handlers {
		if fn, ok := h.(func(context.Context)); ok {
			fn(ctx)
		}
	}
}

func (b *Bus) EmitSpiderError(ctx context.Context, err error) {
	b.mu.RLock()
	handlers := b.handlers[SpiderError]
	b.mu.RUnlock()
	for _, h := range handlers {
		if fn, ok := h.(func(context.Context, error)); ok {
			fn(ctx, err)
		}
	}
}

func (b *Bus) EmitSpiderIdle(ctx context.Context) {
	b.mu.RLock()
	handlers := b.handlers[SpiderIdle]
	b.mu.RUnlock()
	for _, h := range handlers {
		if fn, ok := h.(func(context.Context)); ok {
			fn(ctx)
		}
	}
}

// item signals
func (b *Bus) EmitItemScraped(ctx context.Context, item any) {
	b.mu.RLock()
	handlers := b.handlers[ItemScraped]
	b.mu.RUnlock()
	for _, h := range handlers {
		if fn, ok := h.(func(context.Context, any)); ok {
			fn(ctx, item)
		}
	}
}

func (b *Bus) EmitItemDropped(ctx context.Context, item any, err error) {
	b.mu.RLock()
	handlers := b.handlers[ItemDropped]
	b.mu.RUnlock()
	for _, h := range handlers {
		if fn, ok := h.(func(context.Context, any, error)); ok {
			fn(ctx, item, err)
		}
	}
}

func (b *Bus) EmitItemError(ctx context.Context, item any, err error) {
	b.mu.RLock()
	handlers := b.handlers[ItemError]
	b.mu.RUnlock()
	for _, h := range handlers {
		if fn, ok := h.(func(context.Context, any, error)); ok {
			fn(ctx, item, err)
		}
	}
}

// request signals

func (b *Bus) EmitRequestScheduled(ctx context.Context, req *core.Request) {
	b.mu.RLock()
	handlers := b.handlers[RequestScheduled]
	b.mu.RUnlock()
	for _, h := range handlers {
		if fn, ok := h.(func(context.Context, *core.Request)); ok {
			fn(ctx, req)
		}
	}
}

func (b *Bus) EmitRequestDropped(ctx context.Context, req *core.Request, err error) {
	b.mu.RLock()
	handlers := b.handlers[RequestDropped]
	b.mu.RUnlock()
	for _, h := range handlers {
		if fn, ok := h.(func(context.Context, *core.Request, error)); ok {
			fn(ctx, req, err)
		}
	}
}

func (b *Bus) EmitRequestError(ctx context.Context, req *core.Request, err error) {
	b.mu.RLock()
	handlers := b.handlers[RequestError]
	b.mu.RUnlock()
	for _, h := range handlers {
		if fn, ok := h.(func(context.Context, *core.Request, error)); ok {
			fn(ctx, req, err)
		}
	}
}

// response signals
func (b *Bus) EmitResponseReceived(ctx context.Context, resp core.IResponseReader) {
	b.mu.RLock()
	handlers := b.handlers[ResponseReceived]
	b.mu.RUnlock()
	for _, h := range handlers {
		if fn, ok := h.(func(context.Context, core.IResponseReader)); ok {
			fn(ctx, resp)
		}
	}
}

// engine signals
func (b *Bus) EmitEngineStarted(ctx context.Context) {
	b.mu.RLock()
	handlers := b.handlers[EngineStarted]
	b.mu.RUnlock()
	for _, h := range handlers {
		if fn, ok := h.(func(context.Context)); ok {
			fn(ctx)
		}
	}
}

func (b *Bus) EmitEngineStopped(ctx context.Context) {
	b.mu.RLock()
	handlers := b.handlers[EngineStopped]
	b.mu.RUnlock()
	for _, h := range handlers {
		if fn, ok := h.(func(context.Context)); ok {
			fn(ctx)
		}
	}
}
