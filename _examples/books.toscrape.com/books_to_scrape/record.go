package books_to_scrape

import (
	"reflect"

	"github.com/tech-engine/goscrapy/pkg/core"
)

/*
   json and csv struct field tags are required, if you want the Record to be exported
   or processed by builtin pipelines
*/

type Record struct {
	J *Job `json:"-" csv:"-"` // JobId is required
	// add you custom fields here
	Title       string `json:"title" csv:"title"`
	Price       string `json:"price" csv:"price"`
	Stock       string `json:"stock" csv:"stock"`
	Rating      string `json:"rating" csv:"rating"`
	Description string `json:"description" csv:"description"`
	Upc         string `json:"upc" csv:"upc"`
	ProductType string `json:"product_type" csv:"product_type"`
	Reviews     string `json:"reviews" csv:"reviews"`
}

// modify below code only if you know what you are doing
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
