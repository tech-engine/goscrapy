package worker

import (
	"context"
	"github.com/tech-engine/goscrapy/pkg/core"
)

type result struct {
	request      *core.Request
	response     core.IResponseReader
	callbackName string
	taskHandle   core.TaskHandle
	err          error
	cancel       context.CancelFunc
}

func (r *result) Request() *core.Request        { return r.request }
func (r *result) Response() core.IResponseReader { return r.response }
func (r *result) CallbackName() string           { return r.callbackName }
func (r *result) TaskHandle() core.TaskHandle    { return r.taskHandle }
func (r *result) Error() error                   { return r.err }
func (r *result) Release() {
	if r.cancel != nil {
		r.cancel()
	}
}

func (r *result) Reset() {
	r.request = nil
	r.response = nil
	r.callbackName = ""
	r.taskHandle = nil
	r.err = nil
	r.cancel = nil
}
