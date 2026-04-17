package testutils

import (
	"reflect"

	"github.com/tech-engine/goscrapy/pkg/core"
)

type DummyJob struct {
	id string
}

func (j *DummyJob) Id() string {
	return "DummyJob"
}

type DummyRecord struct {
	Id   string    `json:"id" csv:"id"`
	Name string    `json:"name" csv:"name"`
	J    *DummyJob `json:"-" csv:"-"`
}

func (o *DummyRecord) Record() *DummyRecord {
	return o
}

func (o *DummyRecord) RecordKeys() []string {
	dataType := reflect.TypeOf(DummyRecord{})
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

func (o *DummyRecord) RecordFlat() []any {
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

func (o *DummyRecord) Job() core.IJob {
	return o.J
}
