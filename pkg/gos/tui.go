package gos

import (
	"context"
	"fmt"
	"time"

	"github.com/tech-engine/goscrapy/pkg/tui"
)

// Handles the graceful initialization and shutdown of both the engine and TUI.
func StartWithTUI[OUT any](ctx context.Context, gosApp *app[OUT], dashboard tui.IDashboard) error {
	// engineCtx, engineCancel := context.WithCancel(ctx)
	// defer engineCancel()
	gosApp.cancelableSignal = newCancelableSignal(ctx)
	defer gosApp.cancelableSignal.cancel()

	if gosApp.hub != nil {
		go gosApp.hub.Start(gosApp.cancelableSignal.ctx)
	}

	errCh := make(chan error, 1)
	go func() {
		errCh <- gosApp.Engine.Start(gosApp.cancelableSignal.ctx)
	}()

	tuiErr := dashboard.Run()
	gosApp.cancelableSignal.cancel()

	// wait for engine to finish cleaning up
	select {
	case err := <-errCh:
		if tuiErr != nil {
			gosApp.lastErr = tuiErr
			return gosApp.lastErr
		}
		return err
	case <-time.After(10 * time.Second):
		if tuiErr != nil {
			gosApp.lastErr = tuiErr
			return gosApp.lastErr
		}
		gosApp.lastErr = fmt.Errorf("tui quit timeout")
		return gosApp.lastErr
	}
}
