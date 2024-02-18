package pipelinemanager

import (
	"context"

	"github.com/tech-engine/goscrapy/pkg/core"
)

// We have added it here as PipelineManager is the one that passes IPipelineItems to pipelines
// and so must be aware of IPipelineItem.
type IPipelineItem interface {
	Get(string) (any, bool)
	Set(string, any) error
	Del(string)
	Keys() []any
	Clear()
}

type IPipeline[OUT any] interface {
	Open(context.Context) error
	Close()
	ProcessItem(IPipelineItem, core.IOutput[OUT]) error
}
