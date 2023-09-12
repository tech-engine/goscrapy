package core

import (
	"context"
	"sync"

	metadata "github.com/tech-engine/goscrapy/pkg/meta_data"
)

func NewPipelineManager[J Job, IN any, OUT any, OR Output[J, OUT]]() *PipelineManager[J, IN, OUT, OR] {
	return &PipelineManager[J, IN, OUT, OR]{}
}

func (pm *PipelineManager[J, IN, OUT, OR]) add(pipeline *pipeline[J, IN, OUT, OR]) *PipelineManager[J, IN, OUT, OR] {
	pm.pipelines = append(pm.pipelines, pipeline)
	return pm
}

func (pm *PipelineManager[J, IN, OUT, OR]) do(original OR, metadata metadata.MetaData) (IN, error) {
	var (
		wg    sync.WaitGroup
		input IN
		err   error
	)
	for _, pipeline := range pm.pipelines {
		// if pipeline set to async will it be run in a separate goroutine
		if pipeline.async {
			wg.Add(1)
			go func(_wg *sync.WaitGroup) {
				defer _wg.Done()
				pipeline.p.ProcessItem(input, original, metadata)
			}(&wg)
			continue
		}

		input, err = pipeline.p.ProcessItem(input, original, metadata)

		if err != nil {
			break
		}
	}
	wg.Wait()

	return input, err
}

// runs when spider starts job
func (pm *PipelineManager[J, IN, OUT, OR]) start(ctx context.Context) error {

	var wg sync.WaitGroup

	errCh := make(chan error, len(pm.pipelines))

	wg.Add(len(pm.pipelines))
	for _, pipeline := range pm.pipelines {

		go func(_wg *sync.WaitGroup, pipeline Pipeline[J, IN, OUT, OR]) {
			defer _wg.Done()

			if err := pipeline.Open(ctx); err != nil {
				errCh <- err
			}
		}(&wg, pipeline.p)
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
func (pm *PipelineManager[J, IN, OUT, OR]) stop() {
	for _, pipeline := range pm.pipelines {
		pipeline.p.Close()
	}
}

func (p *pipeline[J, IN, OUT, OR]) WithRequired() *pipeline[J, IN, OUT, OR] {
	p.required = true
	return p
}

func (p *pipeline[J, IN, OUT, OR]) WithAsync() *pipeline[J, IN, OUT, OR] {
	p.required = true
	return p
}
