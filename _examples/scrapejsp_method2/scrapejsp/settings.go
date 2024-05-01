package scrapejsp

import (
	"github.com/tech-engine/goscrapy/pkg/builtin/middlewares"
	"github.com/tech-engine/goscrapy/pkg/builtin/pipelines"
	"github.com/tech-engine/goscrapy/pkg/middlewaremanager"
	pm "github.com/tech-engine/goscrapy/pkg/pipeline_manager"
)

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
func myCustomPipelineGroup() *pm.Group[*Record] {
	pipelineGroup := pm.NewGroup[*Record]()
	pipelineGroup.Add(export2CSV)
	// pipelineGroup.Add(export2Json)
	return pipelineGroup
}

// Pipelines here
// Executed in the order they appear.
var PIPELINES = []pm.IPipeline[*Record]{
	export2CSV,
	// export2Json,
	myCustomPipelineGroup(),
}
