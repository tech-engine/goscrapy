package rp

type Pool[T any] chan *T

func (p Pool[T]) Acquire() (_p *T) {
	select {
	case item := <-p:
		return item
	default:
		return _p
	}
}

func (p Pool[T]) Release(item *T) {
	select {
	case p <- item:
	default:
	}
}

func NewPool[K any](max uint64) Pool[K] {
	itemPool := make(Pool[K], max)
	return itemPool
}
