package cpio

import (
	"fmt"
	"reflect"
)

func wrapError(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("cpio: %v", err)
}

func copyStructField(dst, src interface{}) {
	v1 := reflect.ValueOf(dst).Elem()
	v2 := reflect.ValueOf(src).Elem()
	t1 := v1.Type()
	t2 := v2.Type()
	for i := 0; i < t1.NumField(); i++ {
		f1 := t1.Field(i)
		f2, found := t2.FieldByName(f1.Name)
		if !found || len(f2.Index) == 0 {
			continue
		}

		a1 := v1.Field(i)
		a2 := v2.Field(f2.Index[0])

		switch a1.Kind() {
		case reflect.Array:
			switch a2.Kind() {
			case reflect.Array:
				reflect.Copy(a1, a2)
			case reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				buf := a1.Slice(0, a1.Len()).Bytes()
				str := fmt.Sprint(a2.Uint())
				for i := range buf {
					buf[i] = 0
				}
				copy(buf, str)
			default:
				panic(fmt.Errorf("unhandled type: (%v, %v) (%v, %v)",
					f1.Name, a1.Kind(), f2.Name, a2.Kind()))
			}
		default:
			panic(fmt.Errorf("unhandled type: (%v, %v)", f1.Name, a1.Kind()))
		}
	}
}
