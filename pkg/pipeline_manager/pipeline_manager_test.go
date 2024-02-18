package pipelinemanager

import (
	"context"
	"reflect"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/tech-engine/goscrapy/pkg/core"
)

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

func (p *doublePipeline[OUT]) ProcessItem(item IPipelineItem, original core.IOutput[OUT]) error {
	rec := original.RecordFlat()
	item.Set("id", rec[0])
	item.Set("age", rec[1].(int)*2)
	return nil
}

// dummy pipeline 2
type dummyPipeline2[OUT any] struct {
	FId  int
	FAge int
}

func newDummyPipeline2[OUT any]() *dummyPipeline2[OUT] {
	return &dummyPipeline2[OUT]{}
}

func (p *dummyPipeline2[OUT]) Open(ctx context.Context) error {
	return nil
}

func (p *dummyPipeline2[OUT]) Close() {
}

func (p *dummyPipeline2[OUT]) ProcessItem(item IPipelineItem, original core.IOutput[OUT]) error {
	id, _ := item.Get("id")
	age, _ := item.Get("age")

	p.FId, _ = id.(int)
	p.FAge, _ = age.(int)
	return nil
}

func TestPipelineManager(t *testing.T) {
	// create a pipeline manager
	var wg sync.WaitGroup
	pipelineManager := New[*dummyRecord]()
	// add a dummy test pipeline
	readPipeline := newDummyPipeline2[*dummyRecord]()
	pipelineManager.Add(
		newDoublePipeline[*dummyRecord](),
		readPipeline,
	)
	// start the pipeline
	wg.Add(1)
	go func() {
		wg.Done()
		pipelineManager.Start(context.Background())
	}()
	// push item to pipeline
	pipelineManager.Push(&dummyRecord{Id: 1, Age: 19})
	// verify what we pushed is what we get
	assert.Equalf(t, 1, readPipeline.FId, "expected id=1, got=%s", readPipeline.FId)
	assert.Equalf(t, 38, readPipeline.FAge, "expected age=1, got=%s", readPipeline.FAge)
	wg.Wait()
}
