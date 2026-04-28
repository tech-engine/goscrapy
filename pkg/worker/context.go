package worker

import (
	"context"
)

// valueOnlyCtx unifies a lifecycle context with a separate value context
type valueOnlyCtx struct {
	context.Context                 // lifecycle context
	valCtx          context.Context // value context
}

func (v *valueOnlyCtx) Value(key any) any {
	// check value context first
	if val := v.valCtx.Value(key); val != nil {
		return val
	}
	// fallback to lifecycle context
	return v.Context.Value(key)
}

// mergeContexts unifies worker and request contexts while preserving traces/values
func mergeContexts(workerCtx, reqCtx context.Context) (context.Context, context.CancelFunc) {
	if reqCtx == nil {
		return context.WithCancel(workerCtx)
	}

	// use fast path for un-cancellable contexts to avoid lock contention
	if reqCtx.Done() == nil {
		return context.WithCancel(&valueOnlyCtx{
			Context: workerCtx,
			valCtx:  reqCtx,
		})
	}

	// slow path for cancellable contexts
	execCtx, cancel := context.WithCancel(reqCtx)

	stop := context.AfterFunc(workerCtx, func() {
		cancel()
	})

	cleanup := func() {
		stop()
		cancel()
	}

	return execCtx, cleanup
}
