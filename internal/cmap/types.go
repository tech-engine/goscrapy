package cmap

type opts struct {
	size   int
	hashFn hashFn
}

type hashFn func(string) uint64

type void struct {
	data any
}
