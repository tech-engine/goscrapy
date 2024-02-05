package cmap

import (
	"fmt"
	"sync"

	"github.com/tech-engine/goscrapy/internal/types"
)

type CMap struct {
	opts
	lock sync.RWMutex
	data map[string]void
}

func NewCMap(optFuncs ...types.OptFunc[opts]) *CMap {

	opts := defaultOpts()

	for _, fn := range optFuncs {
		fn(&opts)
	}

	return &CMap{
		opts: opts,
		data: make(map[string]void, opts.size),
	}
}

func (cm *CMap) Get(key string) (any, bool) {

	cm.lock.RLock()
	defer cm.lock.RUnlock()

	val, ok := cm.data[key]

	return val.data, ok
}

func (cm *CMap) Set(key string, val any) error {

	cm.lock.Lock()
	defer cm.lock.Unlock()

	_, ok := cm.data[key]

	if (len(cm.data) > cm.size) && !ok {
		return fmt.Errorf("Set: max items of %d exceeded", cm.size)
	}

	cm.data[key] = void{val}

	return nil
}

func (cm *CMap) Del(key string) {
	delete(cm.data, key)
}

func (cm *CMap) Clear() {
	clear(cm.data)
}

func (cm *CMap) Keys() []any {
	keys := make([]any, cm.size)

	var i = 0
	for key := range cm.data {
		keys[i] = key
		i++
	}

	return keys
}

func (cm *CMap) Len() int {

	cm.lock.RLock()
	defer cm.lock.RUnlock()

	return len(cm.data)
}
