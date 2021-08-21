package bind

import (
	"database/sql"
	"reflect"
	"unsafe"
)

func makeScanIntoList(v interface{}, cts []*sql.ColumnType) []interface{} {
	t := reflect.TypeOf(v)

	vt := reflect.ValueOf(v)
	if t.Kind() == reflect.Ptr {
		vt = vt.Elem()
		t = t.Elem()
	}

	scanInto := []interface{}{}

	for _, ct := range cts {
		for i := 0; i < t.NumField(); i++ {
			ft := t.Field(i)
			name, exists := ft.Tag.Lookup("bind")
			if exists {
				if ct.Name() == name {
					ptr := reflect.NewAt(t.Field(i).Type, unsafe.Pointer(vt.Field(i).UnsafeAddr()))
					actualPtr := ptr.Interface()
					scanInto = append(scanInto, actualPtr)
				}
			}
		}
	}
	return scanInto

}

// Bind binds given rows to the given object at v.
func Bind(rows *sql.Rows, v interface{}) error {
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
		inputs = append(inputs, makeScanIntoList(v, cts))
	} else {
		for i := 0; i < vt.Len(); i++ {
			p := vt.Index(i)
			if p.Type().Kind() == reflect.Ptr {
				p = p.Elem()
			}
			inputs = append(inputs, makeScanIntoList(p.Interface(), cts))
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
