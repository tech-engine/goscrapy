package fsm

import (
	"errors"
	"fmt"

	"github.com/tech-engine/goscrapy/internal/cmap"
)

var ERR_MAX_ITEM_EXCEEDED = errors.New("FSM max item exceeded")

type FixedSizeMap[K comparable, V any] struct {
	size uint64
	data map[K]cmap.Void[V]
}

func New[K comparable, V any](size uint64) *FixedSizeMap[K, V] {
	return &FixedSizeMap[K, V]{
		size: size,
		data: make(map[K]cmap.Void[V], size),
	}
}

func (fsm *FixedSizeMap[K, V]) Get(key K) (V, bool) {

	val, ok := fsm.data[key]

	return val.Data, ok
}

func (fsm *FixedSizeMap[K, V]) Set(key K, val V) error {

	if _, ok := fsm.data[key]; !ok && (len(fsm.data) >= int(fsm.size)) {
		return fmt.Errorf("Set:fixedsizemap.go: %w: max allowed=[%d]", ERR_MAX_ITEM_EXCEEDED, fsm.size)
	}

	fsm.data[key] = cmap.Void[V]{Data: val}

	return nil
}

func (fsm *FixedSizeMap[K, V]) Clear() {
	clear(fsm.data)
}
