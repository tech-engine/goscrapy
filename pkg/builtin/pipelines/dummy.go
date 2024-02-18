package pipelines

import (
	"reflect"

	"github.com/tech-engine/goscrapy/pkg/core"
)

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
	dataType := reflect.TypeOf(dummyRecord{})
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

	inputType := reflect.TypeOf(o)

	if inputType.Kind() != reflect.Struct {
		panic("Record is not a struct")
	}

	inputValue := reflect.ValueOf(o)

	slice := make([]any, inputType.NumField())

	for i := 0; i < inputType.NumField(); i++ {
		slice[i] = inputValue.Field(i).Interface()
	}
	return slice
}

func (o *dummyRecord) Job() core.IJob {
	return o.J
}

type dummyRecord struct {
	Id   string    `json:"id" csv:"id"`
	Name string    `json:"name" csv:"name"`
	J    *dummyJob `json:"-" csv:"-"`
}
