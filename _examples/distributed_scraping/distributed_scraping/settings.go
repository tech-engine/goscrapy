package distributed_scraping

import (
	"os"

	"github.com/tech-engine/goscrapy/pkg/builtin/middlewares"
	"github.com/tech-engine/goscrapy/pkg/builtin/pipelines/csv"
	"github.com/tech-engine/goscrapy/pkg/engine"
	"github.com/tech-engine/goscrapy/pkg/middlewaremanager"
)

// Default: INFO
const GOS_LOG_LEVEL = "INFO"

// Concurrency settings
const AUTOSCALER_MAX_WORKERS = "50"
const AUTOSCALER_MIN_WORKERS = "10"

var MIDDLEWARES = []middlewaremanager.Middleware{
	middlewares.Retry(),
	middlewares.MultiCookieJar,
}

var export2CSV = csv.New[*Record](csv.Options{
	Filename: "distributed_results.csv",
})

var PIPELINES = []engine.IPipeline[*Record]{
	export2CSV,
}

func init() {
	var settings = map[string]string{
		"GOS_LOG_LEVEL":          GOS_LOG_LEVEL,
		"AUTOSCALER_MAX_WORKERS": AUTOSCALER_MAX_WORKERS,
		"AUTOSCALER_MIN_WORKERS": AUTOSCALER_MIN_WORKERS,
	}

	for key, value := range settings {
		if value != "" {
			os.Setenv(key, value)
		}
	}
}

// User define settings here
// Redis settings for distributed scraping
const REDIS_ADDR = "localhost:6379"
const REDIS_USER = "default"
const REDIS_PASSWORD = ""
const REDIS_KEY = "goscrapy:distributed_example"
