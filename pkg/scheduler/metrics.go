package scheduler

import (
	"math"
	"sync/atomic"
)

type ema struct {
	bits   atomic.Uint64
	alpha  float32 // smoothing factor
	inited bool
}

func newEMA(alpha float32) *ema {
	return &ema{
		alpha: alpha,
	}
}

// Feeds a new sample into the EMA
func (e *ema) Add(sample float64) {
	if !e.inited {
		e.bits.Store(math.Float64bits(sample))
		e.inited = true
		return
	}

	prev := math.Float64frombits(e.bits.Load())
	alpha := float64(e.alpha)
	next := alpha*sample + (1-alpha)*prev
	e.bits.Store(math.Float64bits(next))
}

// Value returns the last stored EMA value
func (e *ema) Value() float64 {
	return math.Float64frombits(e.bits.Load())
}
