package gosm

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/tech-engine/goscrapy/pkg/core"
	"github.com/tidwall/gjson"
)

// plan cache to avoid reflection overhead on every call
var planCache sync.Map

type fieldPlan struct {
	index     int
	jsonTag   string
	cssTag    string // raw selector
	cssAttr   string // attr name if @ is used
	xPathTag  string // raw xpath
	xPathAttr string // attr name if @ is used
	isStruct  bool
	isPtr     bool

	// pre-bound setter for speed
	setter func(v reflect.Value, res any)
}

type structPlan struct {
	fields []fieldPlan
}

// Map populates target from source using tags (gos, gos_css, gos_xpath)
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

	// normalize raw json once so we don't re-parse per field
	source = normalizeSource(source)

	return executePlan(source, elem, plan.(*structPlan))
}

func normalizeSource(source any) any {
	switch s := source.(type) {
	case []byte:
		if len(s) > 0 && (s[0] == '{' || s[0] == '[') {
			return gjson.ParseBytes(s)
		}
	case string:
		if len(s) > 0 && (s[0] == '{' || s[0] == '[') {
			return gjson.Parse(s)
		}
	}
	return source
}

func buildPlan(t reflect.Type) *structPlan {
	plan := &structPlan{
		fields: make([]fieldPlan, 0, t.NumField()),
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		// skip unexported
		if !field.IsExported() {
			continue
		}

		fp := fieldPlan{index: i}

		kind := field.Type.Kind()
		if kind == reflect.Struct || (kind == reflect.Ptr && field.Type.Elem().Kind() == reflect.Struct) {
			fp.isStruct = true
			fp.isPtr = kind == reflect.Ptr
		}

		fp.jsonTag = field.Tag.Get("gos")

		if raw := field.Tag.Get("gos_css"); raw != "" {
			fp.cssTag, fp.cssAttr = parseAttrCSS(raw)
		}

		if raw := field.Tag.Get("gos_xpath"); raw != "" {
			fp.xPathTag, fp.xPathAttr = parseAttrXPath(raw)
		}

		fp.setter = bindSetter(field.Type)

		if fp.isStruct || fp.jsonTag != "" || fp.cssTag != "" || fp.xPathTag != "" {
			plan.fields = append(plan.fields, fp)
		}
	}

	return plan
}

func unwrap(val any) any {
	if slice, ok := val.([]string); ok {
		if len(slice) > 0 {
			return slice[0]
		}
		return ""
	}
	return val
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
		itemSetter := bindSetter(elemType)
		return func(v reflect.Value, val any) {
			if slice, ok := val.([]string); ok {
				if elemType.Kind() == reflect.String {
					v.Set(reflect.ValueOf(slice))
					return
				}
				// non-string slice: convert elements
				newSlice := reflect.MakeSlice(t, len(slice), len(slice))
				for i, s := range slice {
					itemSetter(newSlice.Index(i), s)
				}
				v.Set(newSlice)
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
			if fp.jsonTag != "" {
				if res, ok := getGJSON(source, fp.jsonTag); ok && res.Exists() {
					subSource = res
				}
			}
			if err := Map(subSource, targetField.Interface()); err != nil {
				return err
			}
			continue
		}

		// check tags in priority order
		if fp.jsonTag != "" {
			if res, ok := getGJSON(source, fp.jsonTag); ok && res.Exists() {
				fp.setter(fieldVal, res)
				continue
			}
		}
		if fp.cssTag != "" {
			if res, ok := getCSS(source, fp.cssTag, fp.cssAttr); ok {
				fp.setter(fieldVal, res)
				continue
			}
		}
		if fp.xPathTag != "" {
			if res, ok := getXPath(source, fp.xPathTag, fp.xPathAttr); ok {
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
	case core.IResponseReader:
		body := s.Bytes()
		if len(body) > 0 && (body[0] == '{' || body[0] == '[') {
			return gjson.ParseBytes(body).Get(path), true
		}
	}
	return gjson.Result{}, false
}

// split "selector@attr" into (selector, attr)
func parseAttrCSS(tag string) (string, string) {
	if i := strings.LastIndex(tag, "@"); i > 0 {
		return tag[:i], tag[i+1:]
	}
	return tag, ""
}

// split "xpath@attr" safely (avoids @ in predicates)
func parseAttrXPath(tag string) (string, string) {
	lastBracket := strings.LastIndex(tag, "]")
	searchFrom := lastBracket + 1
	if i := strings.Index(tag[searchFrom:], "@"); i >= 0 {
		pos := searchFrom + i
		return tag[:pos], tag[pos+1:]
	}
	return tag, ""
}

func getCSS(source any, selector string, attr string) ([]string, bool) {
	resp, ok := source.(core.IResponseReader)
	if !ok {
		return nil, false
	}
	var vals []string
	if attr != "" {
		vals = resp.Css(selector).Attr(attr)
	} else {
		vals = resp.Css(selector).Text()
	}
	return vals, len(vals) > 0
}

func getXPath(source any, path string, attr string) ([]string, bool) {
	resp, ok := source.(core.IResponseReader)
	if !ok {
		return nil, false
	}
	var vals []string
	if attr != "" {
		vals = resp.Xpath(path).Attr(attr)
	} else {
		vals = resp.Xpath(path).Text()
	}
	return vals, len(vals) > 0
}
