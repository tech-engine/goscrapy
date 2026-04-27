package engine

import (
	"maps"
	"sync"
	"sync/atomic"

	"github.com/tech-engine/goscrapy/pkg/core"
)

type CallbackRegistry struct {
	mu sync.Mutex
	m  atomic.Pointer[map[string]core.ResponseCallback]
}

func NewCallbackRegistry() *CallbackRegistry {
	r := &CallbackRegistry{}
	m := make(map[string]core.ResponseCallback)
	r.m.Store(&m)
	return r
}

func (r *CallbackRegistry) Register(name string, cb core.ResponseCallback) {
	r.mu.Lock()
	defer r.mu.Unlock()

	oldMapPtr := r.m.Load()
	newMap := maps.Clone(*oldMapPtr)
	newMap[name] = cb
	r.m.Store(&newMap)
}

func (r *CallbackRegistry) Resolve(name string) (core.ResponseCallback, bool) {
	cb, ok := (*r.m.Load())[name]
	return cb, ok
}

func (r *CallbackRegistry) Deregister(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	oldMapPtr := r.m.Load()
	newMap := maps.Clone(*oldMapPtr)
	delete(newMap, name)
	r.m.Store(&newMap)
}
