package binder

import (
	"database/sql"
	"reflect"
	"strings"
	"unsafe"
)

func getTypeName(t reflect.Type) string {
	parts := strings.Split(t.Name(), ".")
	return strings.ToLower(parts[len(parts)-1])
}
func makePtrsOf(v reflect.Value, cts []*sql.ColumnType) []interface{} {
	t := v.Type()

	if t.Kind() == reflect.Ptr {
		v = v.Elem()
		t = t.Elem()
	}
	tyName := getTypeName(v.Type())
	scanInto := []interface{}{}
	for index := 0; index < len(cts); index++ {
		ct := cts[index]
		for i := 0; i < t.NumField(); i++ {
			ft := t.Field(i)
			if ft.Type.Kind() == reflect.Ptr {
				return append(scanInto, makePtrsOf(v.Field(i).Elem(), cts)...)
			} else if ft.Type.Kind() == reflect.Struct {
				return append(scanInto, makePtrsOf(v.Field(i), cts)...)
			} else {
				name, exists := ft.Tag.Lookup("sqlname")
				if exists {
					if ct.Name() == name || ct.Name() == tyName+"_"+name {
						ptr := reflect.NewAt(t.Field(i).Type, unsafe.Pointer(v.Field(i).UnsafeAddr()))
						actualPtr := ptr.Elem().Addr().Interface()
						scanInto = append(scanInto, actualPtr)
						newcts := append(cts[:index], cts[index+1:]...)
						return append(scanInto, makePtrsOf(v, newcts)...)
					}
				}
			}

		}
	}
	return scanInto

}

type FromRows interface {
	FromRows(rows *sql.Rows)
}

// Bind binds given rows to the given object at v.
func Bind(rows *sql.Rows, v interface{}) error {
	if fr, isFr := v.(FromRows); isFr {
		fr.FromRows(rows)
		return nil
	}
	cts, err := rows.ColumnTypes()
	if err != nil {
		return err
	}

	t := reflect.TypeOf(v)
	vt := reflect.ValueOf(v)

	if t.Kind() == reflect.Ptr {
		vt = vt.Elem()
		t = t.Elem()
	}

	inputs := [][]interface{}{}
	if t.Kind() != reflect.Slice {
		inputs = append(inputs, makePtrsOf(reflect.ValueOf(v), cts))
	} else {
		for i := 0; i < vt.Len(); i++ {
			p := vt.Index(i).Elem()
			if p.Type().Kind() == reflect.Ptr {
				p = p.Elem()
			}
			newCts := make([]*sql.ColumnType, len(cts))
			copy(newCts, cts)
			ptrs := makePtrsOf(p, newCts)
			inputs = append(inputs, ptrs)
		}
	}

	i := 0
	for rows.Next() && i < len(inputs) {
		err = rows.Scan(inputs[i]...)
		if err != nil {
			return err
		}
		i++
	}

	return nil

}
