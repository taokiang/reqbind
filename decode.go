package bind

import (
	"fmt"
	"reflect"
)

// decodeInto 是 Decode[T] 的具体实现。
func decodeInto[T any](ctx RequestContext) (T, error) {
	var zero T

	// 获取 T 的 reflect.Type（不受零值影响）
	rt := reflect.TypeOf((*T)(nil)).Elem()

	isPtr := rt.Kind() == reflect.Ptr
	baseType := rt
	if isPtr {
		baseType = rt.Elem()
	}

	if baseType.Kind() != reflect.Struct {
		return zero, fmt.Errorf("bind: Decode requires a struct or *struct type, got %s", rt)
	}

	// 获取（缓存的）字段元数据
	fields, err := cachedFields(baseType)
	if err != nil {
		return zero, err
	}

	// 创建可寻址的结构体值
	pv := reflect.New(baseType) // *BaseType
	sv := pv.Elem()             // BaseType（可寻址，因为是通过指针访问）

	for _, fm := range fields {
		fv := sv.Field(fm.Index)

		switch fm.Source {
		case fieldSourcePath:
			raw := ctx.PathParam(fm.Key)
			if raw == "" && fm.HasDefault {
				raw = fm.Default
			}
			if raw != "" {
				if err := setScalar(fv, raw); err != nil {
					return zero, fmt.Errorf("bind: path[%q]: %w", fm.Key, err)
				}
			}

		case fieldSourceHeader:
			raw := ctx.Header(fm.Key)
			if raw == "" && fm.HasDefault {
				raw = fm.Default
			}
			if raw != "" {
				if err := setScalar(fv, raw); err != nil {
					return zero, fmt.Errorf("bind: header[%q]: %w", fm.Key, err)
				}
			}

		case fieldSourceQuery:
			raw := ctx.Query(fm.Key)
			if raw == "" && fm.HasDefault {
				raw = fm.Default
			}
			if raw != "" {
				if err := setScalar(fv, raw); err != nil {
					return zero, fmt.Errorf("bind: query[%q]: %w", fm.Key, err)
				}
			}

		case fieldSourceBody:
			body := ctx.Body()
			if body == nil {
				continue
			}
			if err := decodeBody(body, fm.Key, fv); err != nil {
				return zero, fmt.Errorf("bind: body[%q]: %w", fm.Key, err)
			}
		}
	}

	if isPtr {
		// T 是指针类型，返回 *BaseType
		return pv.Interface().(T), nil
	}
	// T 是值类型，返回 BaseType
	return sv.Interface().(T), nil
}
