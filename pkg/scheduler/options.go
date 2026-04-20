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
	numWorkers     uint16
	reqResPoolSize uint64
	workQueueSize  uint64
	statsFactory   ts.IStatsRecorderFactory
	adaptive       *AdaptiveScalingConfig
}

// Holds all configs for adaptive worker scaling
type AdaptiveScalingConfig struct {
	MinWorkers    uint16
	MaxWorkers    uint16
	ScalingFactor float32
	EMAAlpha      float32
	ScalingWindow time.Duration
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

// All adaptive scaling config field under AdaptiveScalingConfig
func WithAdaptiveScaling(cfg AdaptiveScalingConfig) types.OptFunc[opts] {
	return func(opts *opts) {

		if cfg.MinWorkers == 0 {
			cfg.MinWorkers = opts.numWorkers
		}
		if cfg.MaxWorkers == 0 {
			cfg.MaxWorkers = opts.numWorkers * 5
		}
		if cfg.ScalingFactor == 0 {
			cfg.ScalingFactor = 1.2
		}
		if cfg.EMAAlpha == 0 {
			cfg.EMAAlpha = 0.3
		}
		if cfg.ScalingWindow == 0 {
			cfg.ScalingWindow = time.Second
		}
		opts.adaptive = &cfg
	}
}
