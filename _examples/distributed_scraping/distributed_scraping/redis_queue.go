// This is currently an demo example only on how a custom ITaskQueue using could be implemented using redis.
// If it works good, will be moved to the main goscrapy package. But it's hasn't been tested heavy
// and so have not much idea on how it would behave.
package distributed_scraping

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/tech-engine/goscrapy/pkg/core"
	"github.com/tech-engine/goscrapy/pkg/engine"
	"github.com/tech-engine/goscrapy/pkg/scheduler"
)

// lua scripts for atomicity
const (
	// pull script
	// RPOP from main queue (returns UUID)
	// if UUID exists, ZADD to active queue with score = now + visibilityTimeout
	// return UUID
	pullScript = `
		local taskId = redis.call("RPOP", KEYS[1])
		if taskId then
			redis.call("ZADD", KEYS[2], ARGV[1], taskId)
			return taskId
		end
		return nil
	`

	// reaper script
	// find tasks in active queue with score <= now
	// remove them from active queue
	// push them back to main queue
	reaperScript = `
		local expiredTasks = redis.call("ZRANGEBYSCORE", KEYS[1], "-inf", ARGV[1])
		if #expiredTasks > 0 then
			redis.call("ZREM", KEYS[1], unpack(expiredTasks))
			for i = 1, #expiredTasks do
				redis.call("LPUSH", KEYS[2], expiredTasks[i])
			end
		end
		return #expiredTasks
	`
)

var (
	pullRedisScript   = redis.NewScript(pullScript)
	reaperRedisScript = redis.NewScript(reaperScript)
)

type redisTaskQueue struct {
	client            *redis.Client
	queueKey          string
	activeKey         string
	dataKey           string
	visibilityTimeout time.Duration
	ctx               context.Context
	cancel            context.CancelFunc
}

type RedisOptions struct {
	Addr              string
	Username          string
	Password          string
	DB                int
	Key               string
	VisibilityTimeout time.Duration // time before an active task is considered orphaned
	ReaperInterval    time.Duration // interval to check for orphaned tasks
}

func NewRedisTaskQueue(opts RedisOptions) scheduler.ITaskQueue {
	rdb := redis.NewClient(&redis.Options{
		Addr:     opts.Addr,
		Username: opts.Username,
		Password: opts.Password,
		DB:       opts.DB,
		Protocol: 2, // force RESP2 to prevent 'Conn has unread data' desyncs
	})

	if opts.Key == "" {
		opts.Key = "goscrapy:tasks"
	}
	if opts.VisibilityTimeout == 0 {
		opts.VisibilityTimeout = 5 * time.Minute
	}
	if opts.ReaperInterval == 0 {
		opts.ReaperInterval = 10 * time.Second
	}

	ctx, cancel := context.WithCancel(context.Background())

	q := &redisTaskQueue{
		client:            rdb,
		queueKey:          opts.Key,
		activeKey:         opts.Key + ":active",
		dataKey:           opts.Key + ":data",
		visibilityTimeout: opts.VisibilityTimeout,
		ctx:               ctx,
		cancel:            cancel,
	}

	go q.startReaper(opts.ReaperInterval)

	return q
}

// stop allows graceful shutdown of the reaper process
func (q *redisTaskQueue) Stop() {
	q.cancel()
	q.client.Close()
}

type redisTask struct {
	Request      *core.Request `json:"request"`
	CallbackName string        `json:"callback_name"`
}

func (q *redisTaskQueue) Push(ctx context.Context, task scheduler.QueuedTask) error {
	taskID := uuid.New().String()

	data, err := json.Marshal(redisTask{
		Request:      task.Request,
		CallbackName: task.CallbackName,
	})
	if err != nil {
		return err
	}

	pipe := q.client.Pipeline()
	pipe.HSet(ctx, q.dataKey, taskID, data)
	pipe.LPush(ctx, q.queueKey, taskID)
	_, err = pipe.Exec(ctx)
	return err
}

func (q *redisTaskQueue) Pull(ctx context.Context) (scheduler.QueuedTask, engine.TaskHandle, error) {
	for {
		select {
		case <-ctx.Done():
			return scheduler.QueuedTask{}, nil, ctx.Err()
		default:
		}

		expiryScore := time.Now().Add(q.visibilityTimeout).Unix()

		result, err := pullRedisScript.Run(ctx, q.client, []string{q.queueKey, q.activeKey}, expiryScore).Result()
		if err != nil && err != redis.Nil {
			return scheduler.QueuedTask{}, nil, err
		}

		if result != nil {
			taskID, ok := result.(string)
			if !ok {
				return scheduler.QueuedTask{}, nil, fmt.Errorf("unexpected lua result type")
			}

			// fetch actual task data
			data, err := q.client.HGet(ctx, q.dataKey, taskID).Result()
			if err != nil {
				// If data is missing but ID was in queue, it's corrupted. Ack to clean it up.
				q.Ack(ctx, taskID)
				continue // Try next task
			}

			var rt redisTask
			if err := json.Unmarshal([]byte(data), &rt); err != nil {
				q.Ack(ctx, taskID) // clean up corrupted data
				continue           // Try next task
			}

			task := scheduler.QueuedTask{
				Request:      rt.Request,
				CallbackName: rt.CallbackName,
			}

			return task, taskID, nil
		}

		// backoff if queue is empty to prevent tight loop
		select {
		case <-ctx.Done():
			return scheduler.QueuedTask{}, nil, ctx.Err()
		case <-time.After(500 * time.Millisecond):
		}
	}
}

func (q *redisTaskQueue) Ack(ctx context.Context, handle engine.TaskHandle) error {
	taskID, ok := handle.(string)
	if !ok {
		return fmt.Errorf("invalid redis task handle")
	}

	pipe := q.client.Pipeline()
	pipe.ZRem(ctx, q.activeKey, taskID)
	pipe.HDel(ctx, q.dataKey, taskID)
	_, err := pipe.Exec(ctx)
	return err
}

func (q *redisTaskQueue) Nack(ctx context.Context, handle engine.TaskHandle) error {
	taskID, ok := handle.(string)
	if !ok {
		return fmt.Errorf("invalid redis task handle")
	}

	// move back from active to main queue immediately
	pipe := q.client.Pipeline()
	pipe.ZRem(ctx, q.activeKey, taskID)
	pipe.LPush(ctx, q.queueKey, taskID)
	_, err := pipe.Exec(ctx)
	return err
}

// background goroutine to recover orphaned tasks
func (q *redisTaskQueue) startReaper(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-q.ctx.Done():
			return
		case <-ticker.C:
			nowScore := time.Now().Unix()
			// suppress errors in background loop
			reaperRedisScript.Run(q.ctx, q.client, []string{q.activeKey, q.queueKey}, nowScore)
		}
	}
}
