package scrapeThisSite

import "reflect"

// do not modify this file

type Output struct {
	records []Record
	err     error
	job     *Job
}

func (o *Output) Records() []Record {
	return o.records
}

func (o *Output) RecordKeys() []string {
	dataType := reflect.TypeOf(Record{})
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

func (o *Output) RecordsFlat() [][]any {
	records := make([][]any, 0, len(o.records))

	var inputType reflect.Type

	for i, record := range o.records {
		if i == 0 {
			inputType = reflect.TypeOf(record)

			if inputType.Kind() != reflect.Struct {
				panic("Record is not a struct")
			}
		}

		inputValue := reflect.ValueOf(record)

		slice := make([]any, inputType.NumField())

		for i := 0; i < inputType.NumField(); i++ {
			slice[i] = inputValue.Field(i).Interface()
		}

		records = append(records, slice)
	}
	return records
}

func (o *Output) Error() error {
	return o.err
}

func (o *Output) Job() *Job {
	return o.job
}

func (o *Output) IsEmpty() bool {
	if o == nil || o.records == nil {
		return true
	}
	return len(o.records) <= 0
}
