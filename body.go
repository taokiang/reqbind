package bind

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/url"
	"reflect"
	"strings"
)

// decodeBody 根据 format 将 body 解码到字段值 fv 中。
func decodeBody(body io.ReadCloser, format string, fv reflect.Value) error {
	defer body.Close()

	switch strings.ToLower(format) {
	case "json":
		return decodeJSON(body, fv)
	case "xml":
		return decodeXML(body, fv)
	case "form":
		return decodeForm(body, fv)
	case "text":
		return decodeText(body, fv)
	default:
		return fmt.Errorf("unsupported body format: %q", format)
	}
}

func decodeJSON(r io.Reader, fv reflect.Value) error {
	ptr := resolvePtr(fv)
	return json.NewDecoder(r).Decode(ptr)
}

func decodeXML(r io.Reader, fv reflect.Value) error {
	ptr := resolvePtr(fv)
	return xml.NewDecoder(r).Decode(ptr)
}

func decodeForm(r io.Reader, fv reflect.Value) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("reading body: %w", err)
	}
	values, err := url.ParseQuery(string(data))
	if err != nil {
		return fmt.Errorf("parsing form body: %w", err)
	}
	return mapFormValues(values, fv)
}

func decodeText(r io.Reader, fv reflect.Value) error {
	data, err := io.ReadAll(r)
	if err != nil {
		return fmt.Errorf("reading body: %w", err)
	}
	fv.SetString(string(data))
	return nil
}

// resolvePtr 处理指针类型字段：若 fv 是指针则分配并返回新值的指针；
// 否则直接返回 fv.Addr().Interface()。
func resolvePtr(fv reflect.Value) any {
	if fv.Kind() == reflect.Ptr {
		if fv.IsNil() {
			fv.Set(reflect.New(fv.Type().Elem()))
		}
		return fv.Interface()
	}
	return fv.Addr().Interface()
}

// mapFormValues 将 url.Values 映射到结构体字段，使用 `form:"key"` 标签（或字段名小写）。
func mapFormValues(values url.Values, fv reflect.Value) error {
	t := fv.Type()
	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		if !sf.IsExported() {
			continue
		}
		key := sf.Tag.Get("form")
		if key == "" {
			key = strings.ToLower(sf.Name)
		}
		val := values.Get(key)
		if val == "" {
			continue
		}
		if err := setScalar(fv.Field(i), val); err != nil {
			return fmt.Errorf("form field %q: %w", key, err)
		}
	}
	return nil
}
