package pipelinemanager

import (
	"context"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/tech-engine/goscrapy/pkg/core"
	"github.com/tech-engine/goscrapy/pkg/engine"
	"github.com/tech-engine/goscrapy/pkg/logger"
)

type safeDummyRecord struct {
	mu      sync.Mutex
	id, age int
}

func (s *safeDummyRecord) Set(id, age int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.id = id
	s.age = age
}

func (s *safeDummyRecord) GetVal() [2]int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return [2]int{s.id, s.age}
}

type dummyRecord struct {
	Id, Age int
}

type dummyJob struct {
	id string
}

func (j *dummyJob) Id() string {
	return "dummyJob"
}

func (o *dummyRecord) Record() *dummyRecord {
	return o
}

func (o *dummyRecord) RecordKeys() []string {
	dataType := reflect.TypeOf(*o)
	if dataType.Kind() != reflect.Struct {
		panic("Record is not a struct")
	}

	numFields := dataType.NumField()
	keys := make([]string, numFields)

	for i := 0; i < numFields; i++ {
		field := dataType.Field(i)
		csvTag := field.Tag.Get("csv")
		keys[i] = csvTag
	}

	return keys
}

func (o *dummyRecord) RecordFlat() []any {

	inputType := reflect.TypeOf(*o)

	if inputType.Kind() != reflect.Struct {
		panic("Record is not a struct")
	}

	inputValue := reflect.ValueOf(*o)

	slice := make([]any, inputType.NumField())

	for i := 0; i < inputType.NumField(); i++ {
		slice[i] = inputValue.Field(i).Interface()
	}
	return slice
}

func (o *dummyRecord) Job() core.IJob {
	return nil
}

// dummy pipeline 1
type doublePipeline[OUT any] struct {
}

func newDoublePipeline[OUT any]() *doublePipeline[OUT] {
	return &doublePipeline[OUT]{}
}

func (p *doublePipeline[OUT]) Open(ctx context.Context) error {
	return nil
}

func (p *doublePipeline[OUT]) Close() {
}

func (p *doublePipeline[OUT]) ProcessItem(item engine.IPipelineItem, original core.IOutput[OUT]) error {
	rec := original.RecordFlat()
	item.Set("id", rec[0])
	item.Set("age", rec[1].(int)*2)
	return nil
}

// dummy pipeline 2
type dummyPipeline2[OUT any] struct {
	safeRecord safeDummyRecord
}

func newDummyPipeline2[OUT any]() *dummyPipeline2[OUT] {
	return &dummyPipeline2[OUT]{
		safeRecord: safeDummyRecord{},
	}
}

func (p *dummyPipeline2[OUT]) Open(ctx context.Context) error {
	return nil
}

func (p *dummyPipeline2[OUT]) Close() {
}

func (p *dummyPipeline2[OUT]) ProcessItem(item engine.IPipelineItem, original core.IOutput[OUT]) error {
	id, _ := item.Get("id")
	age, _ := item.Get("age")
	p.safeRecord.Set(id.(int), age.(int))

	return nil
}

func TestPipelineManager(t *testing.T) {
	// create a pipeline manager
	var wg sync.WaitGroup
	logger := logger.NewNoopLogger()
	cfg := DefaultConfig[*dummyRecord]()
	cfg.Logger = logger
	pipelineManager := New(cfg)

	// add a dummy test pipeline
	readPipeline := newDummyPipeline2[*dummyRecord]()
	pipelineManager.Add(
		newDoublePipeline[*dummyRecord](),
		readPipeline,
	)
	// start the pipeline with a cancellable context
	ctx, cancel := context.WithCancel(context.Background())
	wg.Add(1)
	go func() {
		defer wg.Done()
		pipelineManager.Start(ctx)
	}()
	// push item to pipeline
	pipelineManager.Push(&dummyRecord{Id: 1, Age: 19})
	// verify what we pushed is what we get
	assert.Eventually(t, func() bool {
		val := readPipeline.safeRecord.GetVal()
		return val[0] == 1 && val[1] == 38
	}, time.Second*2, time.Millisecond*5)

	safeRecord := readPipeline.safeRecord.GetVal()
	assert.Equalf(t, 1, safeRecord[0], "expected id=1, got=%d", safeRecord[0])
	assert.Equalf(t, 38, safeRecord[1], "expected age=38, got=%d", safeRecord[1])

	// signal the manager to stop and wait for it
	cancel()
	wg.Wait()
}

func TestPipelineManager_LoggerCoverage(t *testing.T) {
	t.Run("DefaultLogger_DoesNotPanic", func(t *testing.T) {
		pm := New[*dummyRecord](nil)
		assert.NotNil(t, pm.logger)

		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		assert.NotPanics(t, func() {
			_ = pm.Start(ctx)
		})
	})
}

func TestPipelineManager_Pooling(t *testing.T) {
	pm := New[*dummyRecord](nil)

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		pm.Start(ctx)
	}()

	// Push 10 items to verify pooling doesn't corrupt state
	for i := 0; i < 10; i++ {
		pm.Push(&dummyRecord{Id: i, Age: 20})
	}

	time.Sleep(100 * time.Millisecond)
	cancel()
	wg.Wait()
}
