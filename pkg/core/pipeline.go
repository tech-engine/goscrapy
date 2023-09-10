package core

import (
	"context"

	metadata "github.com/tech-engine/goscrapy/pkg/meta_data"
)

func IsRequired[J Job, IN any, OUT any, OR Output[J, OUT]](r bool) PipelineOption[J, IN, OUT, OR] {
	return func(p Pipeline[J, IN, OUT, OR]) {
		p.SetRequired(r)
	}
}

func IsAsync[J Job, IN any, OUT any, OR Output[J, OUT]](r bool) PipelineOption[J, IN, OUT, OR] {
	return func(p Pipeline[J, IN, OUT, OR]) {
		p.SetAsync(r)
	}
}

func NewPipelineManager[J Job, IN any, OUT any, OR Output[J, OUT]]() *PipelineManager[J, IN, OUT, OR] {
	return &PipelineManager[J, IN, OUT, OR]{}
}

func (p *PipelineManager[J, IN, OUT, OR]) add(pipeline Pipeline[J, IN, OUT, OR]) *PipelineManager[J, IN, OUT, OR] {
	p.pipelines = append(p.pipelines, pipeline)
	return p
}

func (p *PipelineManager[J, IN, OUT, OR]) do(original OR, metadata metadata.MetaData) (IN, error) {
	var (
		input IN
		err   error
	)
	for _, pipeline := range p.pipelines {
		input, err = pipeline.ProcessItem(input, original, metadata)
		if err != nil {
			break
		}
	}

	return input, err
}

// runs when spider starts job
func (p *PipelineManager[J, IN, OUT, OR]) start(ctx context.Context) error {
	for _, pipeline := range p.pipelines {
		if err := pipeline.Open(ctx); err != nil {
			return err
		}
	}
	return nil
}

// runs when the spider is done
func (p *PipelineManager[J, IN, OUT, OR]) stop() {
	for _, pipeline := range p.pipelines {
		pipeline.Close()
	}
}
