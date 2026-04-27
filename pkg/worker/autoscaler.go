package worker

import (
	"context"
	"math"
	"sync/atomic"
	"time"

	"github.com/tech-engine/goscrapy/pkg/core"
)

// autoscaler handles the adaptive worker scaling logic using Little's Law.
// L = lambda * W
type autoscaler struct {
	// config related
	minWorkers    uint16
	maxWorkers    uint16
	scalingFactor float32
	scalingWindow time.Duration

	// metrics
	taskArrivalCnt      atomic.Uint64
	cummulativeExecTime atomic.Uint64
	doneTaskCnt         atomic.Uint64

	// scaling state
	desiredWorkerCnt atomic.Uint32
	lastScaleTime    atomic.Int64

	// smoothed metrics
	lambdaEMA      *ema
	serviceTimeEMA *ema

	// scheduler callbacks
	currentWorkerCntFn func() int32
	spawnWorkerFn      func(ctx context.Context)
	despawnWorkerFn    func()
	logger             core.ILogger
}

type AutoscalerConfig struct {
	MaxWorkers    uint32
	MinWorkers    uint32
	ScalingFactor float32
	ScalingWindow time.Duration
	EMAAlpha      float32
	// Scheduler wiring
	currentWorkerCntFn func() int32
	spawnWorkerFn      func(ctx context.Context)
	despawnWorkerFn    func()
	logger             core.ILogger
}

func newAutoscaler(cfg *AutoscalerConfig) *autoscaler {
	if cfg.MaxWorkers == 0 {
		cfg.MaxWorkers = 16
	}
	if cfg.MinWorkers == 0 {
		cfg.MinWorkers = 1
	}
	if cfg.ScalingFactor == 0 {
		cfg.ScalingFactor = 1.0
	}
	if cfg.ScalingWindow == 0 {
		cfg.ScalingWindow = 1 * time.Second
	}
	if cfg.EMAAlpha == 0 {
		cfg.EMAAlpha = 0.1
	}

	a := &autoscaler{
		minWorkers:         uint16(cfg.MinWorkers),
		maxWorkers:         uint16(cfg.MaxWorkers),
		scalingFactor:      cfg.ScalingFactor,
		scalingWindow:      cfg.ScalingWindow,
		lambdaEMA:          newEMA(cfg.EMAAlpha),
		serviceTimeEMA:     newEMA(cfg.EMAAlpha),
		currentWorkerCntFn: cfg.currentWorkerCntFn,
		spawnWorkerFn:      cfg.spawnWorkerFn,
		despawnWorkerFn:    cfg.despawnWorkerFn,
		logger:             cfg.logger,
	}
	return a
}

// incr taskArrivalCnt, use to keep track of number of requests arriving in the scheduler
func (a *autoscaler) OnTaskArrival() {
	a.taskArrivalCnt.Add(1)
}

// called by workers after completing a task, use to keep track of number of completed tasks and their execution time
func (a *autoscaler) OnTaskDone(d time.Duration) {
	a.cummulativeExecTime.Add(uint64(d))
	a.doneTaskCnt.Add(1)
}

// initializes the desired worker count on startup
func (a *autoscaler) SetDesired(n uint32) {
	a.desiredWorkerCnt.Store(n)
}

// runs the adaptive scaling loop, blocks until ctx is cancelled
func (a *autoscaler) Start(ctx context.Context) {
	window := a.scalingWindow
	if window <= 0 {
		window = time.Second
	}

	ticker := time.NewTicker(window)
	defer ticker.Stop()

	var prevTaskArrivalCnt uint64

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			a.tick(window, &prevTaskArrivalCnt, ctx)
		}
	}
}

// tick performs a single scaling evaluation
func (a *autoscaler) tick(window time.Duration, prevTaskArrivalCnt *uint64, ctx context.Context) {
	// we calculate the task arrival rate(req/s) i.e lambda
	currTaskArrivalCnt := a.taskArrivalCnt.Load()
	taskDeltaCnt := currTaskArrivalCnt - *prevTaskArrivalCnt
	*prevTaskArrivalCnt = currTaskArrivalCnt
	lambda := float64(taskDeltaCnt) / window.Seconds()
	a.lambdaEMA.Add(lambda)

	// we calculate the average service time of tasks i.e W
	cummulativeExecTime := a.cummulativeExecTime.Swap(0)
	doneTaskCnt := a.doneTaskCnt.Swap(0)

	if doneTaskCnt > 0 {
		avgW := float64(cummulativeExecTime) / float64(doneTaskCnt) / float64(time.Second)
		a.serviceTimeEMA.Add(avgW)
	}

	// Little's Law: L = lambda * W
	smoothedLambda := a.lambdaEMA.Value()
	smoothedW := a.serviceTimeEMA.Value()

	// no service time data, we return
	if smoothedW == 0 {
		return
	}

	targetL := uint16(math.Ceil(smoothedLambda * smoothedW * float64(a.scalingFactor)))
	a.adjust(ctx, targetL)
}

func (a *autoscaler) adjust(ctx context.Context, target uint16) {
	if target < a.minWorkers {
		target = a.minWorkers
	}
	if target > a.maxWorkers {
		target = a.maxWorkers
	}

	current := uint16(a.currentWorkerCntFn())
	if target == current {
		return
	}

	// cooldown to prevent oscillation
	cooldown := a.scalingWindow / 2
	if cooldown < 200*time.Millisecond {
		cooldown = 200 * time.Millisecond
	}
	lastScaleTime := a.lastScaleTime.Load()
	if lastScaleTime > 0 && time.Since(time.Unix(0, lastScaleTime)) < cooldown {
		return
	}

	a.desiredWorkerCnt.Store(uint32(target))

	if target > current {
		diff := target - current
		a.logger.Debugf("Scaling UP: %d -> %d (+%d workers)", current, target, diff)
		for i := uint16(0); i < diff; i++ {
			a.spawnWorkerFn(ctx)
		}
	} else if target < current {
		diff := current - target
		for i := uint16(0); i < diff; i++ {
			a.despawnWorkerFn()
		}
	}
	a.lastScaleTime.Store(time.Now().UnixNano())
}

type ema struct {
	alpha float64
	val   float64
	init  bool
}

func newEMA(alpha float32) *ema {
	return &ema{alpha: float64(alpha)}
}

func (e *ema) Add(val float64) {
	if !e.init {
		e.val = val
		e.init = true
		return
	}
	e.val = e.alpha*val + (1-e.alpha)*e.val
}

func (e *ema) Value() float64 {
	return e.val
}
