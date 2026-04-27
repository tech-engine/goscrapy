package with_tui_stats

import (
	"os"

	"github.com/tech-engine/goscrapy/pkg/builtin/middlewares"
	"github.com/tech-engine/goscrapy/pkg/builtin/pipelines/csv"

	// "github.com/tech-engine/goscrapy/pkg/builtin/pipelines/json"
	"github.com/tech-engine/goscrapy/pkg/engine"
	"github.com/tech-engine/goscrapy/pkg/middlewaremanager"
)

// Pipeline Manager settings

// Default: 24
const PIPELINEMANAGER_ITEM_SIZE = ""

// Default: 5000
const PIPELINEMANAGER_OUTPUT_QUEUE_BUF_SIZE = ""

// Default: 150
const PIPELINEMANAGER_MAX_PROCESS_ITEM_CONCURRENCY = ""

// Middlewares here
// Executed in reverse order from bottom to top.
var MIDDLEWARES = []middlewaremanager.Middleware{
	middlewares.Stats(HttpStats),
	middlewares.Retry(),
	middlewares.MultiCookieJar,
	middlewares.DupeFilter,
}

var export2CSV = csv.New[*Record](csv.Options{
	Filename: "itstimeitsnowornever.csv",
})

// use export 2 json pipeline
// var export2Json = json.New[*Record](json.Options{
// 	Filename:  "itstimeitsnowornever.json",
// 	Immediate: true,
// })

// add pipeline to group
//func myCustomPipelineGroup() *pm.Group[*Record] {
//	pipelineGroup := pm.NewGroup[*Record]()
//	pipelineGroup.Add(export2CSV)
//	// pipelineGroup.Add(export2Json)
//	return pipelineGroup
//}

// Pipelines here
// Executed in the order they appear.
var PIPELINES = []engine.IPipeline[*Record]{
	export2CSV,
	// export2Json,
	// myCustomPipelineGroup(),
}

func init() {
	var settings = map[string]string{
		"GOS_LOG_LEVEL":                                GOS_LOG_LEVEL,
		"MIDDLEWARE_HTTP_TIMEOUT_MS":                   MIDDLEWARE_HTTP_TIMEOUT_MS,
		"MIDDLEWARE_HTTP_MAX_IDLE_CONN":                MIDDLEWARE_HTTP_MAX_IDLE_CONN,
		"MIDDLEWARE_HTTP_MAX_CONN_PER_HOST":            MIDDLEWARE_HTTP_MAX_CONN_PER_HOST,
		"MIDDLEWARE_HTTP_MAX_IDLE_CONN_PER_HOST":       MIDDLEWARE_HTTP_MAX_IDLE_CONN_PER_HOST,
		"MIDDLEWARE_HTTP_RETRY_MAX_RETRIES":            MIDDLEWARE_HTTP_RETRY_MAX_RETRIES,
		"MIDDLEWARE_HTTP_RETRY_CODES":                  MIDDLEWARE_HTTP_RETRY_CODES,
		"MIDDLEWARE_HTTP_RETRY_BASE_DELAY":             MIDDLEWARE_HTTP_RETRY_BASE_DELAY,		"SCHEDULER_CONCURRENCY":                        SCHEDULER_CONCURRENCY,		"PIPELINEMANAGER_ITEM_SIZE":                    PIPELINEMANAGER_ITEM_SIZE,
		"PIPELINEMANAGER_OUTPUT_QUEUE_BUF_SIZE":        PIPELINEMANAGER_OUTPUT_QUEUE_BUF_SIZE,
		"PIPELINEMANAGER_MAX_PROCESS_ITEM_CONCURRENCY": PIPELINEMANAGER_MAX_PROCESS_ITEM_CONCURRENCY,
	}

	for key, value := range settings {
		if value != "" {
			os.Setenv(key, value)
		}
	}
}
