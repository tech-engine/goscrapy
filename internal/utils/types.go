package utils

import (
	"time"
)

type RandomTicker struct {
	C     chan time.Time
	stopc chan struct{}
	min   int64
	max   int64
}
