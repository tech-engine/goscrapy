package corespider

import (
	"context"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/tech-engine/goscrapy/pkg/core"
	"github.com/tech-engine/goscrapy/pkg/engine"
	"github.com/tech-engine/goscrapy/pkg/executor"
	httpnative "github.com/tech-engine/goscrapy/pkg/executor_adapters/http_native"
	"github.com/tech-engine/goscrapy/pkg/middlewaremanager"
	pipelinemanager "github.com/tech-engine/goscrapy/pkg/pipeline_manager"
	"github.com/tech-engine/goscrapy/pkg/scheduler"
)

func New[OUT any]() *CoreSpiderBuilder[OUT] {

	c := &CoreSpiderBuilder[OUT]{}

	c.HttpClient = createDefaultHTTPClient()

	c.MiddlewareManager = middlewaremanager.New(c.HttpClient)

	c.ExecutorAdapter = httpnative.NewHTTPClientAdapter(
		c.MiddlewareManager.HTTPClient(),
	)

	c.Executor = executor.New(c.ExecutorAdapter)

	c.Scheduler = scheduler.New(c.Executor)

	c.PipelineManager = pipelinemanager.New[OUT]()

	c.Engine = engine.New(c.Scheduler, c.PipelineManager)

	c.Core = core.New[OUT](c.Engine)

	return c
}

func (c *CoreSpiderBuilder[OUT]) Start(ctx context.Context) error {
	return c.Engine.Start(ctx)
}

// createDefaultHTTPClient creates a default http client with defaults.
// If default values are set in the env it will pick the defaults from the env.
func createDefaultHTTPClient() *http.Client {
	cli := &http.Client{
		Timeout: MIDDLEWARE_DEFAULT_HTTP_TIMEOUT_MS * time.Millisecond,
	}

	t := http.DefaultTransport.(*http.Transport).Clone()

	t.MaxIdleConns = MIDDLEWARE_DEFAULT_HTTP_MAX_IDLE_CONN
	t.MaxConnsPerHost = MIDDLEWARE_DEFAULT_HTTP_MAX_CONN_PER_HOST
	t.MaxIdleConnsPerHost = MIDDLEWARE_DEFAULT_HTTP_MAX_IDLE_CONN_PER_HOST

	value, ok := os.LookupEnv("MIDDLEWARE_HTTP_MAX_IDLE_CONN")

	if ok {
		maxIdleConn, err := strconv.Atoi(value)
		if err == nil {
			t.MaxIdleConns = maxIdleConn
		}
	}

	value, ok = os.LookupEnv("MIDDLEWARE_HTTP_MAX_CONN_PER_HOST")

	if ok {
		maxConnPerHost, err := strconv.Atoi(value)
		if err == nil {
			t.MaxConnsPerHost = maxConnPerHost
		}
	}

	value, ok = os.LookupEnv("MIDDLEWARE_HTTP_MAX_IDLE_CONN_PER_HOST")

	if ok {
		maxIdleConnPerHost, err := strconv.Atoi(value)
		if err == nil {
			t.MaxConnsPerHost = maxIdleConnPerHost
		}
	}

	value, ok = os.LookupEnv("MIDDLEWARE_DEFAULT_HTTP_TIMEOUT_MS")

	if ok {
		timeoutMs, err := strconv.Atoi(value)
		if err == nil {
			cli.Timeout = time.Duration(timeoutMs) * time.Millisecond
		}
	}

	cli.Transport = t

	return cli
}
