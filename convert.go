package bind

import (
	"fmt"
	"reflect"
	"strconv"
	"time"
)

// setScalar 将字符串 s 解析后设置到反射值 v 中。
// v 的类型可以是：string、bool、int*、uint*、float*、time.Duration、time.Time 或上述类型的指针。
func setScalar(v reflect.Value, s string) error {
	// 处理指针类型：分配新值后递归设置
	if v.Kind() == reflect.Ptr {
		if s == "" {
			return nil
		}
		elem := reflect.New(v.Type().Elem())
		if err := setScalar(elem.Elem(), s); err != nil {
			return err
		}
		v.Set(elem)
		return nil
	}

	// 特殊类型：time.Duration
	if v.Type() == reflect.TypeOf(time.Duration(0)) {
		d, err := time.ParseDuration(s)
		if err != nil {
			return fmt.Errorf("cannot parse %q as time.Duration: %w", s, err)
		}
		v.SetInt(int64(d))
		return nil
	}

	// 特殊类型：time.Time（RFC3339 格式）
	if v.Type() == reflect.TypeOf(time.Time{}) {
		t, err := time.Parse(time.RFC3339, s)
		if err != nil {
			return fmt.Errorf("cannot parse %q as time.Time (expected RFC3339): %w", s, err)
		}
		v.Set(reflect.ValueOf(t))
		return nil
	}

	switch v.Kind() {
	case reflect.String:
		v.SetString(s)

	case reflect.Bool:
		b, err := strconv.ParseBool(s)
		if err != nil {
			return fmt.Errorf("cannot parse %q as bool: %w", s, err)
		}
		v.SetBool(b)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(s, 10, v.Type().Bits())
		if err != nil {
			return fmt.Errorf("cannot parse %q as %s: %w", s, v.Type(), err)
		}
		v.SetInt(n)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n, err := strconv.ParseUint(s, 10, v.Type().Bits())
		if err != nil {
			return fmt.Errorf("cannot parse %q as %s: %w", s, v.Type(), err)
		}
		v.SetUint(n)

	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(s, v.Type().Bits())
		if err != nil {
			return fmt.Errorf("cannot parse %q as %s: %w", s, v.Type(), err)
		}
		v.SetFloat(f)

	default:
		return fmt.Errorf("unsupported scalar type: %s", v.Type())
	}
	return nil
}
