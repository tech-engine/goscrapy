package scrapejsp

import (
	"github.com/tech-engine/goscrapy/pkg/builtin/middlewares"
	"github.com/tech-engine/goscrapy/pkg/builtin/pipelines"
	"github.com/tech-engine/goscrapy/pkg/middlewaremanager"
	pm "github.com/tech-engine/goscrapy/pkg/pipeline_manager"
)

// Middlewares here
var MIDDLEWARES = []middlewaremanager.Middleware{
	middlewares.Retry(),
	middlewares.MultiCookieJar,
	middlewares.DupeFilter,
}

var export2CSV = pipelines.Export2CSV[*Record](pipelines.Export2CSVOpts{
	Filename: "itstimeitsnowornever.csv",
})

// Pipelines here
var PIPELINES = []pm.IPipeline[*Record]{
	export2CSV,
	// export2Json,
}