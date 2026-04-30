package gosm

import (
	"fmt"
	"reflect"
	"strconv"
	"sync"

	"github.com/tech-engine/goscrapy/pkg/core"
	"github.com/tidwall/gjson"
)

// map[reflect.Type]*structPlan
var planCache sync.Map

type fieldPlan struct {
	index    int
	tagJSON  string
	tagCSS   string
	tagXPath string
	isStruct bool
	isPtr    bool

	// pre-bound setter to avoid switch in hot loop
	setter func(v reflect.Value, res any)
}

type structPlan struct {
	fields []fieldPlan
}

// Map populates a struct with data from a response or gjson result using tags.
// Supported tags: gos (gjson path), gos_css (CSS selector), gos_xpath (XPath).
func Map(source any, target any) error {
	v := reflect.ValueOf(target)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("target must be a pointer to a struct")
	}

	elem := v.Elem()
	t := elem.Type()

	plan, ok := planCache.Load(t)
	if !ok {
		plan, _ = planCache.LoadOrStore(t, buildPlan(t))
	}

	return executePlan(source, elem, plan.(*structPlan))
}

func buildPlan(t reflect.Type) *structPlan {
	plan := &structPlan{
		fields: make([]fieldPlan, 0, t.NumField()),
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// skip unexported fields to avoid reflect panics
		if !field.IsExported() {
			continue
		}

		fp := fieldPlan{index: i}

		kind := field.Type.Kind()
		if kind == reflect.Struct || (kind == reflect.Ptr && field.Type.Elem().Kind() == reflect.Struct) {
			fp.isStruct = true
			fp.isPtr = kind == reflect.Ptr
		}

		fp.tagJSON = field.Tag.Get("gos")
		fp.tagCSS = field.Tag.Get("gos_css")
		fp.tagXPath = field.Tag.Get("gos_xpath")

		// bind setter for this field's type
		fp.setter = bindSetter(field.Type)

		if fp.isStruct || fp.tagJSON != "" || fp.tagCSS != "" || fp.tagXPath != "" {
			plan.fields = append(plan.fields, fp)
		}
	}

	return plan
}

func bindSetter(t reflect.Type) func(reflect.Value, any) {
	kind := t.Kind()

	switch kind {
	case reflect.Ptr:
		elemType := t.Elem()
		innerSetter := bindSetter(elemType)
		return func(v reflect.Value, val any) {
			if v.IsNil() {
				v.Set(reflect.New(elemType))
			}
			innerSetter(v.Elem(), val)
		}

	case reflect.Slice:
		elemType := t.Elem()
		itemSetter := bindSetter(elemType) // pre-compute once, not per-call
		return func(v reflect.Value, val any) {
			if slice, ok := val.([]string); ok {
				if elemType.Kind() == reflect.String {
					v.Set(reflect.ValueOf(slice))
				}
				return
			}
			if res, ok := val.(gjson.Result); ok && res.IsArray() {
				items := res.Array()
				newSlice := reflect.MakeSlice(t, len(items), len(items))
				for i, item := range items {
					itemSetter(newSlice.Index(i), item)
				}
				v.Set(newSlice)
			}
		}
	}

	// unwrap extracts a single value from a []string (CSS/XPath results)
	unwrap := func(val any) any {
		if slice, ok := val.([]string); ok {
			if len(slice) > 0 {
				return slice[0]
			}
			return ""
		}
		return val
	}

	// scalar types
	switch kind {
	case reflect.String:
		return func(v reflect.Value, val any) {
			val = unwrap(val)
			if res, ok := val.(gjson.Result); ok {
				v.SetString(res.String())
			} else if s, ok := val.(string); ok {
				v.SetString(s)
			}
		}
	case reflect.Float32, reflect.Float64:
		return func(v reflect.Value, val any) {
			val = unwrap(val)
			if res, ok := val.(gjson.Result); ok {
				v.SetFloat(res.Float())
			} else if s, ok := val.(string); ok {
				f, _ := strconv.ParseFloat(s, 64)
				v.SetFloat(f)
			}
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return func(v reflect.Value, val any) {
			val = unwrap(val)
			if res, ok := val.(gjson.Result); ok {
				v.SetInt(res.Int())
			} else if s, ok := val.(string); ok {
				i, _ := strconv.ParseInt(s, 10, 64)
				v.SetInt(i)
			}
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return func(v reflect.Value, val any) {
			val = unwrap(val)
			if res, ok := val.(gjson.Result); ok {
				v.SetUint(res.Uint())
			} else if s, ok := val.(string); ok {
				u, _ := strconv.ParseUint(s, 10, 64)
				v.SetUint(u)
			}
		}
	case reflect.Bool:
		return func(v reflect.Value, val any) {
			val = unwrap(val)
			if res, ok := val.(gjson.Result); ok {
				v.SetBool(res.Bool())
			} else if s, ok := val.(string); ok {
				b, _ := strconv.ParseBool(s)
				v.SetBool(b)
			}
		}
	}

	// unsupported types get a no-op setter
	return func(v reflect.Value, val any) {}
}

func executePlan(source any, elem reflect.Value, plan *structPlan) error {
	for _, fp := range plan.fields {
		fieldVal := elem.Field(fp.index)

		if fp.isStruct {
			targetField := fieldVal
			if fp.isPtr {
				if fieldVal.IsNil() {
					fieldVal.Set(reflect.New(fieldVal.Type().Elem()))
				}
				targetField = fieldVal
			} else {
				targetField = fieldVal.Addr()
			}

			subSource := source
			if fp.tagJSON != "" {
				if res, ok := getGJSON(source, fp.tagJSON); ok && res.Exists() {
					subSource = res
				}
			}
			_ = Map(subSource, targetField.Interface())
			continue
		}

		// try tags in priority order
		if fp.tagJSON != "" {
			if res, ok := getGJSON(source, fp.tagJSON); ok && res.Exists() {
				fp.setter(fieldVal, res)
				continue
			}
		}
		if fp.tagCSS != "" {
			if res, ok := getCSS(source, fp.tagCSS); ok {
				fp.setter(fieldVal, res)
				continue
			}
		}
		if fp.tagXPath != "" {
			if res, ok := getXPath(source, fp.tagXPath); ok {
				fp.setter(fieldVal, res)
				continue
			}
		}
	}
	return nil
}

func getGJSON(source any, path string) (gjson.Result, bool) {
	switch s := source.(type) {
	case gjson.Result:
		return s.Get(path), true
	case []byte:
		return gjson.ParseBytes(s).Get(path), true
	case string:
		return gjson.Parse(s).Get(path), true
	case core.IResponseReader:
		body := s.Bytes()
		if len(body) > 0 && (body[0] == '{' || body[0] == '[') {
			return gjson.ParseBytes(body).Get(path), true
		}
	}
	return gjson.Result{}, false
}

func getCSS(source any, selector string) ([]string, bool) {
	if resp, ok := source.(core.IResponseReader); ok {
		texts := resp.Css(selector).Text()
		return texts, len(texts) > 0
	}
	return nil, false
}

func getXPath(source any, path string) ([]string, bool) {
	if resp, ok := source.(core.IResponseReader); ok {
		texts := resp.Xpath(path).Text()
		return texts, len(texts) > 0
	}
	return nil, false
}
