package worker

import "github.com/tech-engine/goscrapy/pkg/core"

type IExecutor interface {
	Execute(req *core.Request, res core.IResponseWriter) error
}
