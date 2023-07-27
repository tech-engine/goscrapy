package utils

import (
	"math/rand"
	"time"
)

func NewRandomTicker(min, max time.Duration) *RandomTicker {
	rt := &RandomTicker{
		C:     make(chan time.Time),
		stopc: make(chan struct{}),
		min:   min.Nanoseconds(),
		max:   max.Nanoseconds(),
	}
	go rt.loop()
	return rt
}

func (rt *RandomTicker) loop() {
	t := time.NewTimer(rt.nextInterval())
	for {
		select {
		case <-rt.stopc:
			t.Stop()
			return
		case <-t.C:
			select {
			case rt.C <- time.Now():
				t.Stop()
				t = time.NewTimer(rt.nextInterval())
			}
		}
	}
}

func (rt *RandomTicker) nextInterval() time.Duration {
	interval := rand.Int63n(rt.max-rt.min) + rt.min
	return time.Duration(interval) * time.Nanosecond
}

func (rt *RandomTicker) Stop() {
	close(rt.stopc)
	close(rt.C)
}
