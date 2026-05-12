package utils

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCancelOrSend(t *testing.T) {
	t.Run("send success", func(t *testing.T) {
		ch := make(chan int, 1)
		ctx := context.Background()
		ok := CancelOrSend(ctx, ch, 42)
		assert.True(t, ok)
		assert.Equal(t, 42, <-ch)
	})

	t.Run("send cancelled", func(t *testing.T) {
		ch := make(chan int) // unbuffered, will block
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		ok := CancelOrSend(ctx, ch, 42)
		assert.False(t, ok)
	})
}

func TestCancelOrReceive(t *testing.T) {
	t.Run("receive success", func(t *testing.T) {
		ch := make(chan int, 1)
		ch <- 42
		ctx := context.Background()
		val, ok, closed := CancelOrReceive(ctx, ch)
		assert.True(t, ok)
		assert.False(t, closed)
		assert.Equal(t, 42, val)
	})

	t.Run("receive cancelled", func(t *testing.T) {
		ch := make(chan int) // empty, will block
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, ok, closed := CancelOrReceive(ctx, ch)
		assert.False(t, ok)
		assert.False(t, closed)
	})

	t.Run("receive closed", func(t *testing.T) {
		ch := make(chan int)
		close(ch)
		ctx := context.Background()
		_, ok, closed := CancelOrReceive(ctx, ch)
		assert.False(t, ok)
		assert.True(t, closed)
	})
}
