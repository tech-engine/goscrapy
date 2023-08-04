package core

import "context"

type SpiderDelegation[IN Job, OUT any] struct {
	m *manager[IN, OUT]
}

func (s *SpiderDelegation[IN, OUT]) ExRequest(ctx context.Context, req *Request, cb ResponseCallback) {
	s.m.exRequest(ctx, req, cb)
}

func (s *SpiderDelegation[IN, OUT]) NewRequest() *Request {
	req := s.m.requestPool.Acquire()
	if req == nil {
		req = &Request{
			method: "GET",
		}
	}
	return req
}

func (s *SpiderDelegation[IN, OUT]) Yield(output Output[IN, OUT]) {
	s.m.outputCh <- output
}
