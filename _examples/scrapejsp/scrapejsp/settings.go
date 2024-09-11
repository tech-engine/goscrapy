package scrapejsp

import (
	"os"

	"github.com/tech-engine/goscrapy/pkg/builtin/middlewares"
	"github.com/tech-engine/goscrapy/pkg/builtin/pipelines"
	"github.com/tech-engine/goscrapy/pkg/middlewaremanager"
	pm "github.com/tech-engine/goscrapy/pkg/pipeline_manager"
)

// HTTP Transport settings

// Default: 10000
const MIDDLEWARE_HTTP_TIMEOUT_MS = ""

// Default: 100
const MIDDLEWARE_HTTP_MAX_IDLE_CONN = ""

// Default: 100
const MIDDLEWARE_HTTP_MAX_CONN_PER_HOST = ""

// Default: 100
const MIDDLEWARE_HTTP_MAX_IDLE_CONN_PER_HOST = ""

// Inbuilt Retry middleware settings

// Default: 3
const MIDDLEWARE_HTTP_RETRY_MAX_RETRIES = ""

// Default: 500, 502, 503, 504, 522, 524, 408, 429
const MIDDLEWARE_HTTP_RETRY_CODES = ""

// Default: 1s
const MIDDLEWARE_HTTP_RETRY_BASE_DELAY = ""

// Default: 1000000
const SCHEDULER_REQ_RES_POOL_SIZE = ""

// Default: 3
const SCHEDULER_NUM_WORKERS_MULTIPLIER = ""

// Default: 1000000
const SCHEDULER_WORK_QUEUE_SIZE = ""

// Pipeline Manager settings

// Default: 10000
const PIPELINEMANAGER_ITEMPOOL_SIZE = ""

// Default: 24
const PIPELINEMANAGER_ITEM_SIZE = ""

// Default: 0
const PIPELINEMANAGER_OUTPUT_QUEUE_BUF_SIZE = ""

// Default: 1000
const PIPELINEMANAGER_MAX_PROCESS_ITEM_CONCURRENCY = ""

// Middlewares here
// Executed in reverse order from bottom to top.
var MIDDLEWARES = []middlewaremanager.Middleware{
	middlewares.Retry(),
	middlewares.MultiCookieJar,
	middlewares.DupeFilter,
}

var export2CSV = pipelines.Export2CSV[*Record](pipelines.Export2CSVOpts{
	Filename: "itstimeitsnowornever.csv",
})

// use export 2 json pipeline
// var export2Json = pipelines.Export2JSON[*Record](pipelines.Export2JSONOpts{
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
var PIPELINES = []pm.IPipeline[*Record]{
	export2CSV,
	// export2Json,
	// myCustomPipelineGroup(),
}

func init() {
	var settings = map[string]string{
		"MIDDLEWARE_HTTP_TIMEOUT_MS":                   MIDDLEWARE_HTTP_TIMEOUT_MS,
		"MIDDLEWARE_HTTP_MAX_IDLE_CONN":                MIDDLEWARE_HTTP_MAX_IDLE_CONN,
		"MIDDLEWARE_HTTP_MAX_CONN_PER_HOST":            MIDDLEWARE_HTTP_MAX_CONN_PER_HOST,
		"MIDDLEWARE_HTTP_MAX_IDLE_CONN_PER_HOST":       MIDDLEWARE_HTTP_MAX_IDLE_CONN_PER_HOST,
		"MIDDLEWARE_HTTP_RETRY_MAX_RETRIES":            MIDDLEWARE_HTTP_RETRY_MAX_RETRIES,
		"MIDDLEWARE_HTTP_RETRY_CODES":                  MIDDLEWARE_HTTP_RETRY_CODES,
		"MIDDLEWARE_HTTP_RETRY_BASE_DELAY":             MIDDLEWARE_HTTP_RETRY_BASE_DELAY,
		"SCHEDULER_REQ_RES_POOL_SIZE":                  SCHEDULER_REQ_RES_POOL_SIZE,
		"SCHEDULER_NUM_WORKERS_MULTIPLIER":             SCHEDULER_NUM_WORKERS_MULTIPLIER,
		"SCHEDULER_WORK_QUEUE_SIZE":                    SCHEDULER_WORK_QUEUE_SIZE,
		"PIPELINEMANAGER_ITEMPOOL_SIZE":                PIPELINEMANAGER_ITEMPOOL_SIZE,
		"PIPELINEMANAGER_ITEM_SIZE":                    PIPELINEMANAGER_ITEM_SIZE,
		"PIPELINEMANAGER_OUTPUT_QUEUE_BUF_SIZE":        PIPELINEMANAGER_OUTPUT_QUEUE_BUF_SIZE,
		"PIPELINEMANAGER_MAX_PROCESS_ITEM_CONCURRENCY": PIPELINEMANAGER_MAX_PROCESS_ITEM_CONCURRENCY,
	}

	for key, value := range settings {
		if value != "" {
			os.Setenv(key, value)
		}
	}
}
