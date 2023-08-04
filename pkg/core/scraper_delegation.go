package core

import "context"

type ScraperDelegation[IN Job, OUT any] struct {
	m *manager[IN, OUT]
}

func (s *ScraperDelegation[IN, OUT]) ExRequest(ctx context.Context, req *Request, cb ResponseCallback) {
	s.m.exRequest(ctx, req, cb)
}

func (s *ScraperDelegation[IN, OUT]) NewRequest() *Request {
	req := s.m.requestPool.Acquire()
	if req == nil {
		req = &Request{
			method: "GET",
		}
	}
	return req
}

func (s *ScraperDelegation[IN, OUT]) Yield(output Output[IN, OUT]) {
	s.m.outputCh <- output
}
