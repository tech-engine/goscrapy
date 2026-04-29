package distributed_scraping

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tech-engine/goscrapy/pkg/core"
	"github.com/tech-engine/goscrapy/pkg/scheduler"
)

func TestRedisTaskQueue_Functional(t *testing.T) {
	// Use the real credentials provided in settings.go
	opts := RedisOptions{
		Addr:              REDIS_ADDR,
		Username:          REDIS_USER,
		Password:          REDIS_PASSWORD,
		Key:               REDIS_KEY + ":test", // Use a test-specific key
		VisibilityTimeout: 20 * time.Second,    // Short timeout for testing reaper
		ReaperInterval:    1 * time.Second,     // Frequent checks for testing
	}

	queue := NewRedisTaskQueue(opts).(*redisTaskQueue)
	defer queue.Stop()

	ctx := context.Background()

	// Clean up any old test data
	queue.client.Del(ctx, queue.queueKey, queue.activeKey, queue.dataKey)

	t.Run("Push and Pull", func(t *testing.T) {
		task := &scheduler.QueuedTask{
			Request:      &core.Request{Method_: "GET"},
			CallbackName: "parse",
		}
		task.Request.Url("http://example.com")

		err := queue.Push(ctx, *task)
		assert.NoError(t, err)

		// Pull the task
		pulledTask, handle, err := queue.Pull(ctx)
		fmt.Println("pull task: ", pulledTask, handle, err)
		time.Sleep(10 * time.Second)
		assert.NoError(t, err)
		assert.NotNil(t, pulledTask)
		assert.Equal(t, "GET", pulledTask.Request.Method_)

		// Verify it's in the active ZSet
		exists, _ := queue.client.ZScore(ctx, queue.activeKey, handle.(string)).Result()
		assert.True(t, exists > 0)

		// Ack the task
		err = queue.Ack(ctx, handle)
		assert.NoError(t, err)

		// Verify it's gone from everything
		zcount, _ := queue.client.ZCard(ctx, queue.activeKey).Result()
		assert.Equal(t, int64(0), zcount)
		hcount, _ := queue.client.HLen(ctx, queue.dataKey).Result()
		assert.Equal(t, int64(0), hcount)
	})

	t.Run("Nack recovery", func(t *testing.T) {
		task := &scheduler.QueuedTask{
			Request:      &core.Request{Method_: "POST"},
			CallbackName: "parse",
		}
		task.Request.Url("http://example.com")

		queue.Push(ctx, *task)
		_, handle, _ := queue.Pull(ctx)

		// Nack it
		err := queue.Nack(ctx, handle)
		assert.NoError(t, err)

		// It should be back in the main queue immediately
		pulledAgain, handleAgain, _ := queue.Pull(ctx)
		assert.NotNil(t, pulledAgain)
		assert.Equal(t, "POST", pulledAgain.Request.Method_)

		// MUST ACK to clean up
		queue.Ack(ctx, handleAgain)
	})

	t.Run("Reaper Orphan Recovery", func(t *testing.T) {
		// This tests if a task is automatically moved back to pending if not Acked
		task := &scheduler.QueuedTask{
			Request:      &core.Request{Method_: "PUT"},
			CallbackName: "parse",
		}
		task.Request.Url("http://example.com")

		queue.Push(ctx, *task)
		_, _, _ = queue.Pull(ctx) // Pull it, but DON'T Ack or Nack

		t.Log("Waiting for visibility timeout (2s) + reaper interval (1s)...")
		time.Sleep(4 * time.Second)

		// The reaper should have moved it back to the main queue
		pulledByReaper, handleByReaper, _ := queue.Pull(ctx)
		assert.NotNil(t, pulledByReaper, "Task should have been recovered by reaper")
		assert.Equal(t, "PUT", pulledByReaper.Request.Method_)

		// MUST ACK to clean up
		queue.Ack(ctx, handleByReaper)
	})
}
