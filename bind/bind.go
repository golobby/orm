package bind

import (
	"database/sql"
	"reflect"
	"unsafe"
)

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

	for rows.Next() {
		err = rows.Scan(scanInto...)
		if err != nil {
			return err
		}
	}

	return nil

}
