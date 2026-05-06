package pipelinemanager

import (
	"context"
	"errors"
	"os"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/tech-engine/goscrapy/internal/cmap"
	"github.com/tech-engine/goscrapy/pkg/core"
	"github.com/tech-engine/goscrapy/pkg/engine"
	"github.com/tech-engine/goscrapy/pkg/logger"
	"github.com/tech-engine/goscrapy/pkg/signal"
	"golang.org/x/sync/errgroup"
)

type Config[OUT any] struct {
	ItemSize                  uint64
	OutputQueueBuffSize       uint64
	MaxProcessItemConcurrency uint64
	Logger                    core.ILogger
	Signals                   signal.ItemBus[OUT]
}

func DefaultConfig[OUT any]() *Config[OUT] {
	c := &Config[OUT]{
		ItemSize:                  PIPELINEMANAGER_ITEM_SIZE,
		OutputQueueBuffSize:       PIPELINEMANAGER_OUTPUT_QUEUE_BUF_SIZE,
		MaxProcessItemConcurrency: PIPELINEMANAGER_MAX_PROCESS_ITEM_CONCURRENCY,
	}

	if envVal, ok := os.LookupEnv("PIPELINEMANAGER_ITEM_SIZE"); ok {
		if v, err := strconv.ParseUint(envVal, 10, 64); err == nil && v > 0 {
			c.ItemSize = v
		}
	}

	if envVal, ok := os.LookupEnv("PIPELINEMANAGER_OUTPUT_QUEUE_BUF_SIZE"); ok {
		if v, err := strconv.ParseUint(envVal, 10, 64); err == nil && v > 0 {
			c.OutputQueueBuffSize = v
		}
	}

	if envVal, ok := os.LookupEnv("PIPELINEMANAGER_MAX_PROCESS_ITEM_CONCURRENCY"); ok {
		if v, err := strconv.ParseUint(envVal, 10, 64); err == nil && v > 0 {
			c.MaxProcessItemConcurrency = v
		}
	}

	return c
}

type PipelineManager[OUT any] struct {
	itemPool                  sync.Pool
	outputQueue               chan core.IOutput[OUT]
	pipelines                 []engine.IPipeline[OUT]
	logger                    core.ILogger
	maxProcessItemConcurrency uint64
	signals                   signal.ItemBus[OUT]
	wg                        sync.WaitGroup
	stopped                   atomic.Bool
}

func New[OUT any](config *Config[OUT]) *PipelineManager[OUT] {
	if config == nil {
		config = DefaultConfig[OUT]()
	}

	if config.Logger == nil {
		config.Logger = logger.EnsureLogger(nil).WithName("PipelineManager")
	}

	pm := &PipelineManager[OUT]{
		outputQueue:               make(chan core.IOutput[OUT], config.OutputQueueBuffSize),
		pipelines:                 make([]engine.IPipeline[OUT], 0),
		logger:                    config.Logger,
		maxProcessItemConcurrency: config.MaxProcessItemConcurrency,
		signals:                   config.Signals,
	}

	pm.itemPool.New = func() any {
		return cmap.NewCMap[string, any](cmap.WithSize(int(config.ItemSize)))
	}

	return pm
}

func (pm *PipelineManager[OUT]) Add(pipeline ...engine.IPipeline[OUT]) {
	pm.pipelines = append(pm.pipelines, pipeline...)
}

// func (pm *PipelineManager[OUT]) WithLogger(loggerIn core.ILogger) engine.IPipelineManager[OUT] {
// 	loggerIn = logger.EnsureLogger(loggerIn)
// 	pm.logger = loggerIn.WithName("PipelineManager")
// 	return pm
// }

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

	// start workers
	concurrency := max(pm.maxProcessItemConcurrency, 1)

	pm.wg.Add(int(concurrency))

	for range concurrency {
		go func() {
			defer pm.wg.Done()
			for res := range pm.outputQueue {
				pm.processItem(res)
			}
			// for {
			// 	select {
			// 	case out, ok := <-pm.outputQueue:
			// 		if !ok {
			// 			return
			// 		}
			// 		pm.processItem(out)
			// 	case <-ctx.Done():
			// 		return
			// 	}
			// }
		}()
	}

	<-ctx.Done()

	// signal stop
	pm.Stop()

	return nil
}

// Stop signals the pipeline manager to shut down by closing the input channel.
// This implements engine.IPipelineManager.
func (pm *PipelineManager[OUT]) Stop() {
	if pm.stopped.Swap(true) {
		return
	}
	pm.logger.Debug("Stopping...")
	close(pm.outputQueue)

	// ensure everything is closed on exit
	pm.stopPipelines()
	// wait for all goroutines to finish (after channel is closed)
	pm.wg.Wait()

	pm.logger.Infof("stopped")
}

// stopPipelines calls the close function of every pipeline
func (pm *PipelineManager[OUT]) stopPipelines() {

	pm.wg.Add(len(pm.pipelines))

	for _, p := range pm.pipelines {
		go func() {
			defer pm.wg.Done()
			p.Close()
		}()

	}
}

func (pm *PipelineManager[OUT]) Push(original core.IOutput[OUT]) {
	// pm.stopped needed to prevent panic when spider tries to push
	// after pm.outputQueue is closed
	if len(pm.pipelines) == 0 || pm.stopped.Load() {
		return
	}

	pm.outputQueue <- original
}

// processItem passes each yield output through our pipelines
func (pm *PipelineManager[OUT]) processItem(original core.IOutput[OUT]) {

	// call sync pipelines
	var (
		pItem *cmap.CMap[string, any] // pipeline item
		err   error
	)

	pItem = pm.itemPool.Get().(*cmap.CMap[string, any])

	defer func() {
		pItem.Clear()
		pm.itemPool.Put(pItem)
	}()

	for _, pipeline := range pm.pipelines {

		// we check if pipeline is a group by checking
		if err = pipeline.ProcessItem(engine.IPipelineItem(pItem), original); err != nil {
			if errors.Is(err, engine.ErrDropItem) {
				pm.logger.Infof("Item dropped by pipeline: %v", err)
				if pm.signals != nil {
					pm.signals.EmitItemDropped(context.Background(), original.Record(), err)
				}
			} else {
				pm.logger.Errorf("Pipeline error: %v", err)
				if pm.signals != nil {
					pm.signals.EmitItemError(context.Background(), original.Record(), err)
				}
			}
			return
		}
	}

	if pm.signals != nil {
		pm.signals.EmitItemScraped(context.Background(), original.Record())
	}
	pm.logger.Debug("Item processed successfully")
}
