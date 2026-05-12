package utils

import "context"

// CancelOrSend returns true if value sent, false if context is cancelled.
func CancelOrSend[T any](ctx context.Context, ch chan<- T, val T) bool {
	select {
	case <-ctx.Done():
		return false
	case ch <- val:
		return true
	}
}

// CancelOrReceive returns value and true if received.
// Returns ok=false, closed=true if channel is closed.
// Returns ok=false, closed=false if context is cancelled.
func CancelOrReceive[T any](ctx context.Context, in <-chan T) (val T, ok bool, closed bool) {
	select {
	case <-ctx.Done():
		return val, false, false
	case val, ok = <-in:
		return val, ok, !ok
	}
}
