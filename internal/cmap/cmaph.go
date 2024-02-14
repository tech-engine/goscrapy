package cmap

import (
	"fmt"
	"sync"

	"github.com/segmentio/fasthash/fnv1a"
	"github.com/tech-engine/goscrapy/internal/types"
)

type CMapH struct {
	opts
	lock         sync.RWMutex
	data         map[uint64]Void[any]
	keys         []any
	lastKeyIndex int
}

func defaultOpts() opts {
	return opts{
		size:   24,
		hashFn: fnv1a.HashString64,
	}
}

func WithSize(size int) types.OptFunc[opts] {
	return func(opts *opts) {
		opts.size = size
	}
}

func WithHashFn(fn hashFn) types.OptFunc[opts] {
	return func(opts *opts) {
		opts.hashFn = fn
	}
}

func NewHCMap(optFuncs ...types.OptFunc[opts]) *CMapH {

	opts := defaultOpts()

	for _, fn := range optFuncs {
		fn(&opts)
	}

	return &CMapH{
		opts: opts,
		data: make(map[uint64]Void[any], opts.size),
		keys: make([]any, opts.size),
	}
}

func (cm *CMapH) Get(key string) (any, bool) {

	hkey := cm.hashFn(key)

	cm.lock.RLock()
	defer cm.lock.RUnlock()

	val, ok := cm.data[hkey]

	return val.Data, ok
}

func (cm *CMapH) Set(key string, val any) error {

	hkey := cm.hashFn(key)

	cm.lock.Lock()
	defer cm.lock.Unlock()

	_, ok := cm.data[hkey]

	if (len(cm.data) > cm.size) && !ok {
		return fmt.Errorf("Set: max items of %d exceeded", cm.size)
	}

	cm.data[hkey] = Void[any]{val}

	if cm.lastKeyIndex < cm.opts.size {
		cm.keys[cm.lastKeyIndex] = key
		cm.lastKeyIndex++
	}

	return nil
}

func (cm *CMapH) Len() int {

	cm.lock.RLock()
	defer cm.lock.RUnlock()

	return len(cm.data)
}

func (cm *CMapH) Del(key string) {
	hkey := cm.hashFn(key)
	cm.lock.Lock()
	delete(cm.data, hkey)
	cm.lock.Unlock()
}

func (cm *CMapH) Clear() {
	clear(cm.data)
}

func (cm *CMapH) Keys() []any {
	return cm.keys
}
