package google_finance

import (
	"os"

	"github.com/tech-engine/goscrapy/pkg/builtin/pipelines/json"
	"github.com/tech-engine/goscrapy/pkg/engine"
	"github.com/tech-engine/goscrapy/pkg/middlewaremanager"
)

var MIDDLEWARES = []middlewaremanager.Middleware{}

var export2Json = json.New[*Record](json.Options{
	Filename:  "itstimeitsnowornever.json",
	Immediate: true,
})

var PIPELINES = []engine.IPipeline[*Record]{
	export2Json,
}

func init() {
	os.Setenv("GOS_LOG_LEVEL", "DEBUG")
}
