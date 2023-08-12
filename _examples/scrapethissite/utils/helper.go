package utils

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

func OnTerminate(fn func()) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-ctx.Done()
	stop()
	fn()
}
