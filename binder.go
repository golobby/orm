package orm

import (
	"database/sql"
	"reflect"
	"unsafe"
)

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

// Bind binds given rows to the given object at v.
func (o *ObjectMetadata) Bind(rows *sql.Rows, v interface{}) error {
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

	var inputs [][]interface{}
	if t.Kind() != reflect.Slice {
		inputs = append(inputs, o.ptrsFor(reflect.ValueOf(v), cts))
	} else {
		for i := 0; i < vt.Len(); i++ {
			p := vt.Index(i).Elem()
			if p.Type().Kind() == reflect.Ptr {
				p = p.Elem()
			}
			newCts := make([]*sql.ColumnType, len(cts))
			copy(newCts, cts)
			ptrs := o.ptrsFor(p, newCts)
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