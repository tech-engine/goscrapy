package signal

import (
	"context"
	"errors"
	"testing"
)

func TestSignalBus_ConnectAndEmit(t *testing.T) {
	bus := New()
	ctx := context.Background()

	t.Run("SpiderIdle", func(t *testing.T) {
		called := false
		bus.Connect(SpiderIdle, func(c context.Context) {
			called = true
		})

		bus.EmitSpiderIdle(ctx)
		if !called {
			t.Error("handler not called")
		}
	})

	t.Run("ItemScraped", func(t *testing.T) {
		var capturedItem any
		item := "test-item"
		bus.Connect(ItemScraped, func(c context.Context, i any) {
			capturedItem = i
		})

		bus.EmitItemScraped(ctx, item)
		if capturedItem != item {
			t.Errorf("expected %v, got %v", item, capturedItem)
		}
	})

	t.Run("MultipleHandlers", func(t *testing.T) {
		count := 0
		bus.Connect(EngineStarted, func(c context.Context) { count++ })
		bus.Connect(EngineStarted, func(c context.Context) { count++ })

		bus.EmitEngineStarted(ctx)
		if count != 2 {
			t.Errorf("expected 2, got %d", count)
		}
	})
}

func TestSignalBus_Robustness(t *testing.T) {
	bus := New()
	ctx := context.Background()

	t.Run("WrongSignatureIgnored", func(t *testing.T) {
		bus.Connect(SpiderIdle, func() {
			t.Error("should not be called")
		})

		bus.EmitSpiderIdle(ctx)
	})

	t.Run("SpiderErrorSignature", func(t *testing.T) {
		err := errors.New("test-error")
		called := false
		bus.Connect(SpiderError, func(c context.Context, e error) {
			called = true
			if e != err {
				t.Error("error mismatch")
			}
		})

		bus.EmitSpiderError(ctx, err)
		if !called {
			t.Error("handler not called")
		}
	})
}
