package cmap

import (
	"errors"
	"fmt"
	"sync"

	"github.com/tech-engine/goscrapy/internal/types"
)

var ERR_MAX_ITEM_EXCEEDED = errors.New("CMAP max item exceeded")

type CMap[K comparable, V any] struct {
	opts
	lock sync.RWMutex
	data map[K]Void[V]
}

func NewCMap[K comparable, V any](optFuncs ...types.OptFunc[opts]) *CMap[K, V] {

	opts := defaultOpts()

	for _, fn := range optFuncs {
		fn(&opts)
	}

	return &CMap[K, V]{
		opts: opts,
		data: make(map[K]Void[V], opts.size),
	}
}

func (cm *CMap[K, V]) Get(key K) (V, bool) {

	cm.lock.RLock()
	defer cm.lock.RUnlock()

	val, ok := cm.data[key]

	return val.Data, ok
}

func (cm *CMap[K, V]) Set(key K, val V) error {

	cm.lock.Lock()
	defer cm.lock.Unlock()

	_, ok := cm.data[key]

	if (len(cm.data) >= cm.size) && !ok {
		return fmt.Errorf("Set: %w: max allowed=[%d]", ERR_MAX_ITEM_EXCEEDED, cm.size)
	}

	cm.data[key] = Void[V]{val}

	return nil
}

func (cm *CMap[K, V]) Del(key K) {
	delete(cm.data, key)
}

func (cm *CMap[K, V]) Clear() {
	clear(cm.data)
}

func (cm *CMap[K, V]) Keys() []any {
	keys := make([]any, cm.size)

	var i = 0
	for key := range cm.data {
		keys[i] = key
		i++
	}

	return keys
}

func (cm *CMap[K, V]) Len() int {

	cm.lock.RLock()
	defer cm.lock.RUnlock()

	return len(cm.data)
}
