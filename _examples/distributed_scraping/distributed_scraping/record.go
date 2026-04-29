package distributed_scraping

import (
	"reflect"

	"github.com/tech-engine/goscrapy/pkg/core"
)

type Record struct {
	J     *Job   `json:"-" csv:"-"`
	Title string `json:"title" csv:"title"`
	Price string `json:"price" csv:"price"`
}

func (r *Record) Record() *Record {
	return r
}

func (r *Record) RecordKeys() []string {
	dataType := reflect.TypeOf(*r)
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

