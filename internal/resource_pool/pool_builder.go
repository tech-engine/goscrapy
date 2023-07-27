package rp

func WithSize[T any](size uint64) PoolOption[T] {
	return func(p *Pooler[T]) {
		p.pool = NewPool[T](size)
	}
}

type Pooler[T any] struct {
	pool Pool[T]
}

type PoolOption[T any] func(*Pooler[T])

func NewPooler[T any](options ...PoolOption[T]) *Pooler[T] {
	pool := &Pooler[T]{}

	for _, option := range options {
		option(pool)
	}

	return pool
}

func (p *Pooler[T]) Acquire() *T {
	if p.pool != nil {
		return p.pool.Acquire()
	}
	return new(T)
}

func (p *Pooler[T]) Release(item *T) {
	if p.pool != nil {
		p.pool.Release(item)
	}
}
