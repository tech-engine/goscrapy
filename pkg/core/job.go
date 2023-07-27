package core

import (
	"context"
	"time"
)

func newManagerJob[IN Job](ctx context.Context, name string, job IN) ManagerJob[IN] {
	_ctx, cancel := context.WithCancel(ctx)
	return ManagerJob[IN]{
		name:       name,
		ScraperJob: job,
		startTime:  time.Now(),
		ctx:        _ctx,
		cancel:     cancel,
	}
}

func (m *ManagerJob[IN]) SetMaxResultsAllowed(maxResultsAllowed uint64) {
	m.maxResultsAllowed = maxResultsAllowed
}
