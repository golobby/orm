package orm

import (
	"database/sql"
	"fmt"
	"reflect"
	"unsafe"
)

// ptrsFor does for each field in struct:
// if field is primitive just allocate and add pointer
// if field is struct call recursively and add all pointers
func (o *ObjectMetadata) ptrsFor(v reflect.Value, cts []*sql.ColumnType) []interface{} {
	t := v.Type()

	if t.Kind() == reflect.Ptr {
		v = v.Elem()
		t = t.Elem()
	}
	tableName := o.Table
	var scanInto []interface{}
	for index := 0; index < len(cts); index++ {
		ct := cts[index]
		for i := 0; i < t.NumField(); i++ {
			ft := t.Field(i)

			if ft.Type.Kind() == reflect.Ptr {
				return append(scanInto, o.Fields[i].RelationMetadata.objectMetadata.ptrsFor(v.Field(i).Elem(), cts)...)
			} else if ft.Type.Kind() == reflect.Struct {
				return append(scanInto, o.Fields[i].RelationMetadata.objectMetadata.ptrsFor(v.Field(i), cts)...)
			} else {
				fieldName := o.Fields[i].Name
				if ct.Name() == fieldName || ct.Name() == tableName+"."+fieldName {
					ptr := reflect.NewAt(t.Field(i).Type, unsafe.Pointer(v.Field(i).UnsafeAddr()))
					actualPtr := ptr.Elem().Addr().Interface()
					scanInto = append(scanInto, actualPtr)
					newcts := append(cts[:index], cts[index+1:]...)
					return append(scanInto, o.ptrsFor(v, newcts)...)
				}
			}

		}
	}
	return scanInto
}

// Bind binds given rows to the given object at obj. obj should be a pointer
func (o *ObjectMetadata) Bind(rows *sql.Rows, obj interface{}) error {
	cts, err := rows.ColumnTypes()
	if err != nil {
		return err
	}

	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)
	if t.Kind() != reflect.Ptr {
		return fmt.Errorf("obj should be a ptr")
	}
	t = t.Elem()
	v = v.Elem()
	if t.Kind() == reflect.Slice {
		t = t.Elem()
		for rows.Next() {
			var rowValue reflect.Value
			if t.Kind() == reflect.Ptr {
				rowValue = reflect.New(t.Elem())
			} else {
				rowValue = reflect.New(t)
			}
			newCts := make([]*sql.ColumnType, len(cts))
			copy(newCts, cts)
			ptrs := o.ptrsFor(rowValue, newCts)
			err = rows.Scan(ptrs...)
			if err != nil {
				return err
			}
			v = reflect.Append(v, rowValue)
		}
	} else {
		for rows.Next() {
			ptrs := o.ptrsFor(v, cts)
			err = rows.Scan(ptrs...)
			if err != nil {
				return err
			}
		}
	}
	// v is either struct or slice
	reflect.ValueOf(obj).Elem().Set(v)
	return nil
}