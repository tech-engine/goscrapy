package distributed_scraping

import (
	"context"

	"github.com/tech-engine/goscrapy/pkg/gos"
)

func New(ctx context.Context) (*Spider, error) {
	//Create the Redis backed Task Queue (implemented locally in redis_queue.go)
	redisQueue := NewRedisTaskQueue(RedisOptions{
		Addr:     REDIS_ADDR,
		Username: REDIS_USER,
		Password: REDIS_PASSWORD,
		Key:      REDIS_KEY,
	})

	// Initialize the GOS application with the custom queue
	app, err := gos.New[*Record](&gos.Config{
		TaskQueue: redisQueue,
	})
	if err != nil {
		return nil, err
	}

	app.WithMiddlewares(MIDDLEWARES...)
	app.WithPipelines(PIPELINES...)

	spider := &Spider{
		ICoreSpider: app,
		baseUrl:     "https://books.toscrape.com",
	}

	app.RegisterSpider(spider)

	go func() {
		_ = app.Start(ctx)
	}()

	return spider, nil
}
