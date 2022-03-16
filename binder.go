package orm

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"reflect"
	"unsafe"
)

// makeNewPointersOf creates a map of [field name] -> pointer to fill it
// recursively. it will go down until reaches a driver.Valuer implementation, it will stop there.
func (b *binder) makeNewPointersOf(v reflect.Value) interface{} {
	m := map[string]interface{}{}
	actualV := v
	for actualV.Type().Kind() == reflect.Ptr {
		actualV = actualV.Elem()
	}
	if actualV.Type().Kind() == reflect.Struct {
		for i := 0; i < actualV.NumField(); i++ {
			f := actualV.Field(i)
			if (f.Type().Kind() == reflect.Struct || f.Type().Kind() == reflect.Ptr) && !f.Type().Implements(reflect.TypeOf((*driver.Valuer)(nil)).Elem()) {
				f = reflect.NewAt(actualV.Type().Field(i).Type, unsafe.Pointer(actualV.Field(i).UnsafeAddr()))
				fm := b.makeNewPointersOf(f).(map[string]interface{})
				for k, p := range fm {
					m[k] = p
				}
			} else {
				var fm *field
				fm = b.s.getField(actualV.Type().Field(i))
				if fm == nil {
					fm = fieldMetadata(actualV.Type().Field(i), b.s.columnConstraints)[0]
				}
				m[fm.Name] = reflect.NewAt(actualV.Field(i).Type(), unsafe.Pointer(actualV.Field(i).UnsafeAddr())).Interface()
			}
		}
	} else {
		return v.Addr().Interface()
	}

	return m
}

// ptrsFor first allocates for all struct fields recursively until reaches a driver.Value impl
// then it will put them in a map with their correct field name as key, then loops over cts
// and for each one gets appropriate one from the map and adds it to pointer list.
func (b *binder) ptrsFor(v reflect.Value, cts []*sql.ColumnType) []interface{} {
	ptrs := b.makeNewPointersOf(v)
	var scanInto []interface{}
	if reflect.TypeOf(ptrs).Kind() == reflect.Map {
		nameToPtr := ptrs.(map[string]interface{})
		for _, ct := range cts {
			if nameToPtr[ct.Name()] != nil {
				scanInto = append(scanInto, nameToPtr[ct.Name()])
			}
		}
	} else {
		scanInto = append(scanInto, ptrs)
	}

	return scanInto
}

type binder struct {
	s *schema
}

func newBinder(s *schema) *binder {
	return &binder{s: s}
}

// bind binds given rows to the given object at obj. obj should be a pointer
func (b *binder) bind(rows *sql.Rows, obj interface{}) error {
	cts, err := rows.ColumnTypes()
	if err != nil {
		return err
	}

	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)
	if t.Kind() != reflect.Ptr {
		return fmt.Errorf("obj should be a ptr")
	}
	// since passed input is always a pointer one deref is necessary
	t = t.Elem()
	v = v.Elem()
	if t.Kind() == reflect.Slice {
		// getting slice elemnt type -> slice[t]
		t = t.Elem()
		for rows.Next() {
			var rowValue reflect.Value
			// Since reflect.SetupConnections returns a pointer to the type, we need to unwrap it to get actual
			rowValue = reflect.New(t).Elem()
			// till we reach a not pointer type continue newing the underlying type.
			for rowValue.IsZero() && rowValue.Type().Kind() == reflect.Ptr {
				rowValue = reflect.New(rowValue.Type().Elem()).Elem()
			}
			newCts := make([]*sql.ColumnType, len(cts))
			copy(newCts, cts)
			ptrs := b.ptrsFor(rowValue, newCts)
			err = rows.Scan(ptrs...)
			if err != nil {
				return err
			}
			for rowValue.Type() != t {
				tmp := reflect.New(rowValue.Type())
				tmp.Elem().Set(rowValue)
				rowValue = tmp
			}
			v = reflect.Append(v, rowValue)
		}
	} else {
		for rows.Next() {
			ptrs := b.ptrsFor(v, cts)
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

func bindToMap(rows *sql.Rows) ([]map[string]interface{}, error) {
	cts, err := rows.ColumnTypes()
	if err != nil {
		return nil, err
	}
	var ms []map[string]interface{}
	for rows.Next() {
		var ptrs []interface{}
		for _, ct := range cts {
			ptrs = append(ptrs, reflect.New(ct.ScanType()).Interface())
		}

		err = rows.Scan(ptrs...)
		if err != nil {
			return nil, err
		}
		m := map[string]interface{}{}
		for i, ptr := range ptrs {
			m[cts[i].Name()] = reflect.ValueOf(ptr).Elem().Interface()
		}

		ms = append(ms, m)
	}
	return ms, nil
}
