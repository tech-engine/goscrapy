package pipelinemanager

import (
	"os"
	"strconv"

	"github.com/tech-engine/goscrapy/internal/types"
)

type opts struct {
	itemPoolSize, outputQueueBuffSize, maxProcessItemConcurrency uint64
}

// Setup all the default pipelinemanger options.
func defaultOpts() opts {
	opts := opts{}
	opts.itemPoolSize = PIPELINEMANAGER_ITEMPOOL_SIZE
	envVal, ok := os.LookupEnv("PIPELINEMANAGER_ITEMPOOL_SIZE")

	if ok {
		parsedPoolSize, err := strconv.ParseUint(envVal, 10, 64)
		if err == nil {
			opts.itemPoolSize = parsedPoolSize
		}
	}

	opts.outputQueueBuffSize = PIPELINEMANAGER_OUTPUT_QUEUE_BUF_SIZE
	envVal, ok = os.LookupEnv("PIPELINEMANAGER_OUTPUT_QUEUE_BUF_SIZE")

	if ok {
		parsedOutputBufSize, err := strconv.ParseUint(envVal, 10, 64)
		if err == nil {
			opts.outputQueueBuffSize = parsedOutputBufSize
		}
	}

	opts.maxProcessItemConcurrency = PIPELINEMANAGER_MAX_PROCESS_ITEM_CONCURRENCY
	envVal, ok = os.LookupEnv("PIPELINEMANAGER_MAX_PROCESS_ITEM_CONCURRENCY")

	if ok {
		parsedMaxItem, err := strconv.ParseUint(envVal, 10, 64)
		if err == nil {
			opts.maxProcessItemConcurrency = parsedMaxItem
		}
	}

	return opts
}

func WithItemPoolSize(val uint64) types.OptFunc[opts] {
	return func(opts *opts) {
		opts.itemPoolSize = val
	}
}

func WithOutputQueueSize(val uint64) types.OptFunc[opts] {
	return func(opts *opts) {
		opts.outputQueueBuffSize = val
	}
}

func WithProcessItemConcurrency(val uint64) types.OptFunc[opts] {
	return func(opts *opts) {
		opts.maxProcessItemConcurrency = val
	}
}
