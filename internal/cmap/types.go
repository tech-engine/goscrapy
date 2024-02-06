package cmap

type opts struct {
	size   int
	hashFn hashFn
}

type hashFn func(string) uint64

type Void[V any] struct {
	Data V
}
