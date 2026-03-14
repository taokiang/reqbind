package bind

import (
	"fmt"
	"reflect"
	"sync"
)

// fieldSource 表示字段值的来源。
type fieldSource uint8

const (
	fieldSourcePath   fieldSource = iota + 1
	fieldSourceHeader             // nolint
	fieldSourceQuery
	fieldSourceBody
)

// fieldMeta 保存一个结构体字段的绑定元数据。
type fieldMeta struct {
	Index      int         // 在结构体中的字段索引
	Source     fieldSource // 值来源
	Key        string      // path/header/query 的 key，或 body 的格式（"json"/"xml"/"form"/"text"）
	Default    string      // default 标签的值
	HasDefault bool
	Type       reflect.Type
}

// fieldCache 缓存已解析的结构体字段元数据，避免每次请求重复反射。
var fieldCache sync.Map // map[reflect.Type][]fieldMeta

// cachedFields 返回类型 t（必须是 struct）的字段元数据，带缓存。
func cachedFields(t reflect.Type) ([]fieldMeta, error) {
	if v, ok := fieldCache.Load(t); ok {
		return v.([]fieldMeta), nil
	}
	fields, err := parseFields(t)
	if err != nil {
		return nil, err
	}
	fieldCache.Store(t, fields)
	return fields, nil
}

// parseFields 解析结构体类型 t 的所有绑定字段。
func parseFields(t reflect.Type) ([]fieldMeta, error) {
	var fields []fieldMeta
	bodyCount := 0

	for i := 0; i < t.NumField(); i++ {
		sf := t.Field(i)
		if !sf.IsExported() {
			continue
		}

		fm, ok, err := parseFieldMeta(i, sf)
		if err != nil {
			return nil, fmt.Errorf("bind: field %s: %w", sf.Name, err)
		}
		if !ok {
			continue
		}

		if fm.Source == fieldSourceBody {
			bodyCount++
			if bodyCount > 1 {
				return nil, fmt.Errorf("bind: struct %s has more than one body field", t.Name())
			}
		}

		fields = append(fields, fm)
	}
	return fields, nil
}

// parseFieldMeta 解析单个字段的标签，返回 (meta, found, error)。
func parseFieldMeta(index int, sf reflect.StructField) (fieldMeta, bool, error) {
	fm := fieldMeta{
		Index: index,
		Type:  sf.Type,
	}

	switch {
	case sf.Tag.Get("path") != "":
		fm.Source = fieldSourcePath
		fm.Key = sf.Tag.Get("path")
	case sf.Tag.Get("header") != "":
		fm.Source = fieldSourceHeader
		fm.Key = sf.Tag.Get("header")
	case sf.Tag.Get("query") != "":
		fm.Source = fieldSourceQuery
		fm.Key = sf.Tag.Get("query")
	case sf.Tag.Get("body") != "":
		fm.Source = fieldSourceBody
		fm.Key = sf.Tag.Get("body")
		if err := validateBodyField(sf); err != nil {
			return fm, false, err
		}
	default:
		return fm, false, nil
	}

	if v, ok := sf.Tag.Lookup("default"); ok {
		fm.Default = v
		fm.HasDefault = true
	}

	return fm, true, nil
}

// validateBodyField 验证 body 字段类型合法性。
func validateBodyField(sf reflect.StructField) error {
	t := sf.Type
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	format := sf.Tag.Get("body")
	switch format {
	case "json", "xml", "form":
		if t.Kind() != reflect.Struct {
			return fmt.Errorf("body:%q field must be a struct or *struct, got %s", format, sf.Type)
		}
	case "text":
		if t.Kind() != reflect.String {
			return fmt.Errorf("body:\"text\" field must be string, got %s", sf.Type)
		}
	default:
		return fmt.Errorf("unknown body format %q (supported: json, xml, form, text)", format)
	}
	return nil
}
