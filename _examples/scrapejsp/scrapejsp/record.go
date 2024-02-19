package scrapejsp

import (
	"reflect"

	"github.com/tech-engine/goscrapy/pkg/core"
)

// do not modify this file

type Record struct {
	J         *Job   `json:"-" csv:"-"` // JobId is required
	UserId    int    `csv:"userId" json:"userId"`
	Id        int    `csv:"id" json:"id"`
	Title     string `csv:"title" json:"title"`
	Completed bool   `csv:"completed" json:"completed"`
}

func (r *Record) Record() *Record {
	return r
}

func (r *Record) RecordKeys() []string {
	dataType := reflect.TypeOf(*r)
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

func (r *Record) RecordFlat() []any {

	inputType := reflect.TypeOf(*r)

	if inputType.Kind() != reflect.Struct {
		panic("Record is not a struct")
	}

	inputValue := reflect.ValueOf(*r)

	slice := make([]any, inputType.NumField())

	for i := 0; i < inputType.NumField(); i++ {
		slice[i] = inputValue.Field(i).Interface()
	}
	return slice
}

func (r *Record) Job() core.IJob {
	return r.J
}
