// GOS is the app build using the goscrapy components.
// This where the components are stitch together to form a complete GOS application
// This is the package that majority is not all spider will be using.
// Unless you may want to tweak the individual components.
package gos

import (
	"context"
	"net/http"

	"github.com/tech-engine/goscrapy/pkg/core"
	"github.com/tech-engine/goscrapy/pkg/middlewaremanager"
)

// Any custom spider created using GoScrapy Framework must implement ICoreSpider[OUT any] interface
type ICoreSpider[OUT any] interface {
	// Parse schedules a request and passed the processed response to the callback
	Parse(req *core.Request, cb core.ResponseCallback)
	// Request creates a new request
	Request(context.Context) *core.Request
	Yield(core.IOutput[OUT])
	Logger() core.ILogger
	// Wait blocks until the spider finishes its work or receives a termination signal (Ctrl+C).
	// If the optional 'autoExit' parameter is set to true, the framework will
	// automatically initiate a graceful shutdown once it detects that all
	// scraping tasks and pipeline operations are complete.
	Wait(...bool) error
}



type IMiddlewareManager interface {
	HTTPClient() *http.Client
	Add(...middlewaremanager.Middleware)
}
