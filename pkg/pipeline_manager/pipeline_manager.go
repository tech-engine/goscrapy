package pipelinemanager

import (
	"context"
	"sync"

	"github.com/tech-engine/goscrapy/internal/cmap"
	rp "github.com/tech-engine/goscrapy/internal/resource_pool"
	"github.com/tech-engine/goscrapy/internal/types"
	"github.com/tech-engine/goscrapy/pkg/core"
	"golang.org/x/sync/errgroup"
)

type PipelineManager[OUT any] struct {
	opts
	itemPool    *rp.Pooler[cmap.CMap]
	outputQueue chan core.IOutput[OUT]
	pipelines   []IPipeline[OUT]
}

func New[OUT any](optFuncs ...types.OptFunc[opts]) *PipelineManager[OUT] {

	// set default options
	opts := defaultOpts()

	// set custom options
	for _, fn := range optFuncs {
		fn(&opts)
	}

	return &PipelineManager[OUT]{
		opts:        opts,
		outputQueue: make(chan core.IOutput[OUT], opts.outputQueueBuffSize),
		pipelines:   make([]IPipeline[OUT], 0),
		itemPool:    rp.NewPooler[cmap.CMap](rp.WithSize[cmap.CMap](opts.itemPoolSize)),
	}
}

func (pm *PipelineManager[OUT]) Add(pipeline ...IPipeline[OUT]) {
	pm.pipelines = append(pm.pipelines, pipeline...)
}

// runs after the spider's Open func and calls all open function of pipelines
func (pm *PipelineManager[OUT]) Start(ctx context.Context) error {

	var (
		group    *errgroup.Group
		groupCtx context.Context
		err      error
	)

	if err = ctx.Err(); err != nil {
		return err
	}

	// Below code ensures that we return an error in case any of the pipelines open
	// funtion returns and error if opts.openMust has been set to true

	group, groupCtx = errgroup.WithContext(ctx)

	for _, pipeline := range pm.pipelines {
		pipeline := pipeline
		group.Go(func() error {
			return pipeline.Open(groupCtx)
		})
	}

	// we return early as there would be no point in processing items on the
	if err = group.Wait(); err != nil {
		return err
	}

	// upon exiting, stop pipeline manager
	defer pm.stop()

	// Below we listen on the outputQueue for new yield outputs

	var wg sync.WaitGroup
	defer wg.Wait()

	// This semaphone will make sure only a fixed number of goroutines
	// are spun up to process items from output queue
	semaphone := make(chan struct{}, pm.opts.maxProcessItemConcurrency)

	for {
		select {
		case semaphone <- struct{}{}:

			wg.Add(1)
			go func() {

				defer wg.Done()
				defer func() { <-semaphone }()

				// this select is to make sure this goroutine doesn't get's blocked
				// waiting for items on queue and get a chance to exit when on context cancellation
				select {
				case item := <-pm.outputQueue:
					if ctx.Err() != nil {
						return
					}
					pm.processItem(item)
				case <-ctx.Done():
					// currently not needed but we also consider closing pm.outputQueue channel in future
					return
				}

			}()

		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// Stoping manager would call the close function of every pipeline
func (pm *PipelineManager[OUT]) stop() {
	var wg sync.WaitGroup

	wg.Add(len(pm.pipelines))
	defer wg.Wait()

	for _, pipeline := range pm.pipelines {

		go func(p IPipeline[OUT]) {
			defer wg.Done()
			p.Close()
		}(pipeline)

	}
}

func (pm *PipelineManager[OUT]) Push(original core.IOutput[OUT]) {
	if len(pm.pipelines) <= 0 {
		return
	}
	pm.outputQueue <- original
}

// Below function passes each yield output through our pipelines
func (pm *PipelineManager[OUT]) processItem(original core.IOutput[OUT]) {

	// call sync pipelines
	var (
		pItem IPipelineItem // pipeline item
		err   error
	)

	pItem = pm.itemPool.Acquire()
	defer func() {
		pItem.Clear()
		pm.itemPool.Release(pItem.(*cmap.CMap))
	}()

	if pItem == nil {
		pItem = cmap.NewCMap()
	}

	for _, pipeline := range pm.pipelines {

		// we check if pipeline is a group by checking
		if err = pipeline.ProcessItem(pItem, original); err != nil {
			return
		}
	}
}
