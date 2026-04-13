package pipelinemanager

import (
	"context"
	"sync"

	"github.com/tech-engine/goscrapy/internal/cmap"
	rp "github.com/tech-engine/goscrapy/internal/resource_pool"
	"github.com/tech-engine/goscrapy/internal/types"
	"github.com/tech-engine/goscrapy/pkg/core"
	"github.com/tech-engine/goscrapy/pkg/engine"
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
		logger:      logger.EnsureLogger(nil).WithName("PipelineManager"),
	}
}

func (pm *PipelineManager[OUT]) Add(pipeline ...IPipeline[OUT]) {
	pm.pipelines = append(pm.pipelines, pipeline...)
}

func (pm *PipelineManager[OUT]) WithLogger(loggerIn core.ILogger) engine.IPipelineManager[OUT] {
	loggerIn = logger.EnsureLogger(loggerIn)
	pm.logger = loggerIn.WithName("PipelineManager")
	return pm
}

// runs after the spider's Open func and calls all open function of pipelines
func (pm *PipelineManager[OUT]) Start(ctx context.Context) error {
	if len(pm.pipelines) == 0 {
		pm.logger.Warn("No pipelines registered, items will be dropped")
	}

	pm.logger.Infof("Starting pipeline manager with %d pipelines", len(pm.pipelines))

	// open all pipelines
	group, groupCtx := errgroup.WithContext(ctx)
	for _, pipeline := range pm.pipelines {
		group.Go(func() error {
			return pipeline.Open(groupCtx)
		})
	}

	if err := group.Wait(); err != nil {
		pm.logger.Errorf("Failed to open pipelines: %v", err)
		return err
	}

	// ensure everything is closed on exit
	defer pm.stop()

	var wg sync.WaitGroup

	// start workers
	concurrency := pm.opts.maxProcessItemConcurrency
	if concurrency == 0 {
		concurrency = 1
	}

	wg.Add(int(concurrency))
	// wait for all goroutines
	defer wg.Wait()

	for i := uint64(0); i < concurrency; i++ {
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case out, ok := <-pm.outputQueue:
					if !ok {
						return
					}
					pm.processItem(out)
				}
			}
		}()
	}

	// wait for framework shutdown
	<-ctx.Done()

	// no more items will be pushed, scheduler has finished
	// so we close queue to signal workers
	close(pm.outputQueue)

	// draining remaining items in outputQueue
	for out := range pm.outputQueue {
		pm.processItem(out)
	}

	pm.logger.Infof("stopped")
	return nil
}

// stop calls the close function of every pipeline
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

	select {
	case pm.outputQueue <- original:
		pm.logger.Debug("📦 Item pushed to pipeline queue")
	default:
		pm.logger.Warn("⚠️ Pipeline queue full, blocking push")
		pm.outputQueue <- original
	}
}

// processItem passes each yield output through our pipelines
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

	pm.logger.Debug("✅ Item processed successfully")
}
