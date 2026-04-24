package scheduler

import "github.com/tech-engine/goscrapy/pkg/core"

type schedulerWork struct {
	next    core.ResponseCallback
	request *core.Request
}

func (s *schedulerWork) Reset() {
	s.next = nil
	s.request = nil
}
