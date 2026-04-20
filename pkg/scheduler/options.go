package scheduler

import (
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/tech-engine/goscrapy/internal/types"
	ts "github.com/tech-engine/goscrapy/pkg/telemetry/stats"
)

type opts struct {
	numWorkers      uint16
	reqResPoolSize  uint64
	workQueueSize   uint64
	statsFactory    ts.IStatsRecorderFactory
	adaptiveScaling bool
	minWorkers      uint16
	maxWorkers      uint16
	scalingFactor   float32
	emaAlpha        float32
	scalingWindow   time.Duration
}

func defaultOpts() opts {
	opts := opts{}
	opts.reqResPoolSize = SCHEDULER_DEFAULT_REQ_RES_POOL_SIZE
	value, ok := os.LookupEnv("SCHEDULER_REQ_RES_POOL_SIZE")

	if ok {
		parsedPoolSize, err := strconv.ParseUint(value, 10, 64)
		if err == nil {
			opts.reqResPoolSize = parsedPoolSize
		}
	}

	opts.numWorkers = uint16(runtime.GOMAXPROCS(0)) * SCHEDULER_DEFAULT_WORKER_MULTIPLIER
	value, ok = os.LookupEnv("SCHEDULER_CONCURRENCY")

	if ok {
		multiplier, err := strconv.ParseUint(value, 10, 16)
		if err == nil {
			opts.numWorkers = uint16(multiplier)
		}
	}

	opts.workQueueSize = SCHEDULER_DEFAULT_WORK_QUEUE_SIZE
	value, ok = os.LookupEnv("SCHEDULER_WORK_QUEUE_SIZE")

	if ok {
		workQueueSize, err := strconv.ParseUint(value, 10, 64)
		if err == nil {
			opts.workQueueSize = workQueueSize
		}
	}

	// adaptive scaling defaults
	opts.adaptiveScaling = false
	opts.minWorkers = opts.numWorkers
	opts.maxWorkers = opts.numWorkers * 5
	opts.scalingFactor = 1.2 // 20% headroom
	opts.emaAlpha = 0.3
	opts.scalingWindow = time.Second

	return opts
}

func WithReqResPoolSize(n uint64) types.OptFunc[opts] {
	return func(opts *opts) {
		opts.reqResPoolSize = n
	}
}

func WithWorkers(n uint16) types.OptFunc[opts] {
	return func(opts *opts) {
		opts.numWorkers = n
	}
}

func WithWorkQueueSize(n uint64) types.OptFunc[opts] {
	return func(opts *opts) {
		opts.workQueueSize = n
	}
}

func WithStatsRecorderFactory(p ts.IStatsRecorderFactory) types.OptFunc[opts] {
	return func(opts *opts) {
		opts.statsFactory = p
	}
}

func WithAdaptiveScaling(enabled bool) types.OptFunc[opts] {
	return func(opts *opts) {
		opts.adaptiveScaling = enabled
	}
}

func WithMinWorkers(n uint16) types.OptFunc[opts] {
	return func(opts *opts) {
		opts.minWorkers = n
	}
}

func WithMaxWorkers(n uint16) types.OptFunc[opts] {
	return func(opts *opts) {
		opts.maxWorkers = n
	}
}

func WithScalingFactor(f float32) types.OptFunc[opts] {
	return func(opts *opts) {
		opts.scalingFactor = f
	}
}

func WithEMAAlpha(a float32) types.OptFunc[opts] {
	return func(opts *opts) {
		opts.emaAlpha = a
	}
}

func WithScalingWindow(d time.Duration) types.OptFunc[opts] {
	return func(opts *opts) {
		opts.scalingWindow = d
	}
}
