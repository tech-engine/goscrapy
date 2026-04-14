package gos

import (
	"context"
	"time"

	"github.com/tech-engine/goscrapy/pkg/tui"
)

// Handles the graceful initialization and shutdown of both the engine and TUI.
func StartWithTUI[OUT any](ctx context.Context, gosApp *app[OUT], dashboard tui.IDashboard) error {
	engineCtx, engineCancel := context.WithCancel(ctx)
	defer engineCancel()

	if gosApp.hub != nil {
		go gosApp.hub.Start(engineCtx)
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- gosApp.Engine.Start(engineCtx)
	}()

	tuiErr := dashboard.Run()
	engineCancel() // tui exits -> cancel engine immediately

	// wait for engine to finish cleaning up
	select {
	case err := <-errCh:
		if tuiErr != nil {
			return tuiErr
		}
		return err
	case <-time.After(10 * time.Second):
		if tuiErr != nil {
			return tuiErr
		}
		return context.DeadlineExceeded
	}
}
