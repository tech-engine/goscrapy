package scrapeThisSite

// do not modify this file

import (
	"context"
	"net/url"

	"github.com/tech-engine/goscrapy/pkg/core"
)

type Spider struct {
	baseUrl   *url.URL
	ctx       context.Context
	delegator core.Delegator[*Job, []Record]
}

func (s *Spider) NewJob(id string) *Job {
	return NewJob(id)
}

func (s *Spider) SetDelegator(delegator core.Delegator[*Job, []Record]) {
	s.delegator = delegator
}

func (s *Spider) NewRequest() *core.Request {
	return s.delegator.NewRequest()
}

func (s *Spider) Request(ctx context.Context, req *core.Request, cb core.ResponseCallback) {
	s.delegator.ExRequest(ctx, req, cb)
}

func (s *Spider) yield(output *Output) {
	s.delegator.Yield(output)
}

func (s *Spider) jobFromContext(ctx context.Context) (*Job, error) {
	meta, ok := ctx.Value("META_DATA").(map[string]any)

	if !ok {
		return nil, ERR_EXTRACTING_META
	}

	job, ok := meta["JOB"].(*Job)

	if !ok {
		return nil, ERR_EXTRACTING_JOB
	}
	return job, nil
}
