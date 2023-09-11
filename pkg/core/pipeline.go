package core

import (
	"context"
	"sync"

	metadata "github.com/tech-engine/goscrapy/pkg/meta_data"
)

func WithRequired[J Job, IN any, OUT any, OR Output[J, OUT]](r bool) PipelineOption[J, IN, OUT, OR] {
	return func(p Pipeline[J, IN, OUT, OR]) {
		p.Required(r)
	}
}

func WithAsync[J Job, IN any, OUT any, OR Output[J, OUT]](r bool) PipelineOption[J, IN, OUT, OR] {
	return func(p Pipeline[J, IN, OUT, OR]) {
		p.Async(r)
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
		wg    sync.WaitGroup
		input IN
		err   error
	)
	for _, pipeline := range p.pipelines {
		// if pipeline set to async will it be run in a separate goroutine
		if pipeline.Async() {
			wg.Add(1)
			go func(_wg *sync.WaitGroup) {
				defer _wg.Done()
				pipeline.ProcessItem(input, original, metadata)
			}(&wg)
			continue
		}

		input, err = pipeline.ProcessItem(input, original, metadata)

		if err != nil {
			break
		}
	}
	wg.Wait()

	return input, err
}

// runs when spider starts job
func (p *PipelineManager[J, IN, OUT, OR]) start(ctx context.Context) error {

	var wg sync.WaitGroup

	errCh := make(chan error, len(p.pipelines))

	wg.Add(len(p.pipelines))
	for _, pipeline := range p.pipelines {

		go func(_wg *sync.WaitGroup, pipeline Pipeline[J, IN, OUT, OR]) {
			defer _wg.Done()

			if err := pipeline.Open(ctx); err != nil {
				errCh <- err
			}
		}(&wg, pipeline)
	}

	go func() {
		wg.Wait()
		close(errCh)
	}()

	for err := range errCh {
		return err
	}

	return nil
}

// runs when the spider is done
func (p *PipelineManager[J, IN, OUT, OR]) stop() {
	for _, pipeline := range p.pipelines {
		pipeline.Close()
	}
}
