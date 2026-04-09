package pipelinemanager

import (
	"context"
	"sync"

	"github.com/tech-engine/goscrapy/internal/cmap"
	rp "github.com/tech-engine/goscrapy/internal/resource_pool"
	"github.com/tech-engine/goscrapy/internal/types"
	"github.com/tech-engine/goscrapy/pkg/core"
	"github.com/tech-engine/goscrapy/pkg/logger"
	"golang.org/x/sync/errgroup"
)

type PipelineManager[OUT any] struct {
	opts
	itemPool    *rp.Pooler[cmap.CMap[string, any]]
	outputQueue chan core.IOutput[OUT]
	pipelines   []IPipeline[OUT]
	logger      core.ILogger
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
		itemPool:    rp.NewPooler(rp.WithSize[cmap.CMap[string, any]](opts.itemPoolSize)),
		logger:      logger.GetLogger(), // default to global logger
	}
}

func (pm *PipelineManager[OUT]) Add(pipeline ...IPipeline[OUT]) {
	pm.pipelines = append(pm.pipelines, pipeline...)
}

func (pm *PipelineManager[OUT]) WithLogger(logger core.ILogger) {
	pm.logger = logger
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

	for i := uint64(0); i < pm.opts.maxProcessItemConcurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case item, ok := <-pm.outputQueue:
					if !ok {
						return
					}
					pm.processItem(item)
				case <-ctx.Done():
					// after cancellation, keep draining outputQueue until empty.
					for item := range pm.outputQueue {
						pm.processItem(item)
					}
					return
				}
			}
		}()
	}

	<-ctx.Done()
	// no more items will be pushed, scheduler has finished
	close(pm.outputQueue)
	return ctx.Err()
}

// Stoping manager would call the close function of every pipeline
func (pm *PipelineManager[OUT]) stop() {
	var wg sync.WaitGroup

	wg.Add(len(pm.pipelines))
	defer wg.Wait()

	for _, p := range pm.pipelines {
		go func() {
			defer wg.Done()
			p.Close()
		}()

	}
}

func (pm *PipelineManager[OUT]) Push(original core.IOutput[OUT]) {
	if len(pm.pipelines) <= 0 {
		return
	}
	pm.logger.Debug("📦 PipelineManager: item pushed to queue")
	pm.outputQueue <- original
}

// Below function passes each yield output through our pipelines
func (pm *PipelineManager[OUT]) processItem(original core.IOutput[OUT]) {

	// call sync pipelines
	var (
		pItem *cmap.CMap[string, any] // pipeline item
		err   error
	)

	pItem = pm.itemPool.Acquire()

	if pItem == nil {
		pItem = cmap.NewCMap[string, any](cmap.WithSize(int(pm.itemPoolSize)))
	}

	defer func() {
		pItem.Clear()
		pm.itemPool.Release(pItem)
	}()

	for _, pipeline := range pm.pipelines {

		// we check if pipeline is a group by checking
		if err = pipeline.ProcessItem(IPipelineItem(pItem), original); err != nil {
			pm.logger.Errorf("❌ Pipeline error: %v", err)
			return
		}
	}
	pm.logger.Debug("✅ PipelineManager: item processed successfully")
}
