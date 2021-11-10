package orm

import (
	"fmt"
	"github.com/golobby/orm/binder"
	"reflect"
	"strings"
	"unsafe"
)

type Entity interface {
	Columns
	PKValue
	PKColumn
	SetPKValue
	Table
	ToMap
	InsertColumnsAndValues
	binder.FromRows
}
type objectHelpers struct {
	// Returns a list of string which are the columns that a struct repreasent based on binder tags.
	Columns func(v interface{}) []string
	// Returns a string which is the table name ( by convention is TYPEs ) of given object
	Table func(v interface{}) string
	// Returns a list of args of the given object, useful for passing as args of sql exec or query
	Values func(v interface{}) []interface{}
	// Returns the primary key for given object.
	PKColumn func(v interface{}) string
	// Sets primary key for object
	SetPK func(obj interface{}, pk interface{})
	// Gets value of primary key of given obj
	PKValue func(obj interface{}) interface{}
	// Returns a Key-Value paired of struct.
	ToMap                    func(obj interface{}) map[string]interface{}
	InsertColumnsAndValuesOf func(obj interface{}) ([]string, []interface{})
}

// ObjectHelpers are set of functions that extract type informations from a struct, it's better to use `ObjectMetadata` if possible also
// implement Entity interface for better performance.
var ObjectHelpers = &objectHelpers{
	Columns:                  columnsOf,
	Table:                    tableName,
	Values:                   valuesOf,
	PKColumn:                 primaryKeyOf,
	SetPK:                    setPrimaryKeyFor,
	PKValue:                  primaryKeyValue,
	ToMap:                    keyValueOf,
	InsertColumnsAndValuesOf: colsAndValsForInsert,
}

type InsertColumnsAndValues interface {
	InsertColumnsAndValues() ([]string, []interface{})
}

func colsAndValsForInsert(o interface{}) ([]string, []interface{}) {
	hv, is := o.(InsertColumnsAndValues)
	if is {
		return hv.InsertColumnsAndValues()
	}
	t := reflect.TypeOf(o)
	v := reflect.ValueOf(o)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}
	var cols []string
	var values []interface{}

	for i := 0; i < t.NumField(); i++ {
		ft := t.Field(i)
		fv := v.Field(i)
		name, exists := ft.Tag.Lookup("bind")
		_, isPK := ft.Tag.Lookup("pk")
		if isPK {
			continue
		}
		if fv.IsZero() {
			continue
		}
		if exists {
			cols = append(cols, name)
		} else {
			cols = append(cols, ft.Name)
		}
		values = append(values, fv.Interface())
	}
	return cols, values
}

// Values defines how a type should return it's args for sql arguments, if not implemented sql will fallback to reflection based approach
type Values interface {
	Values() []interface{}
}

func valuesOf(o interface{}) []interface{} {
	hv, isHasValues := o.(Values)
	if isHasValues {
		return hv.Values()
	}
	v := reflect.ValueOf(o)
	if v.Type().Kind() == reflect.Ptr {
		v = v.Elem()
	}
	var values []interface{}
	for i := 0; i < v.NumField(); i++ {
		fv := v.Field(i)
		if fv.IsZero() {
			continue
		}
		values = append(values, fv.Interface())
	}
	return values
}

// Table defines how a type should return it's coresponding table name, if not implemented sql will fallback to reflection based approach
type Table interface {
	Table() string
}

func tableName(v interface{}) string {
	hv, isTableName := v.(Table)
	if isTableName {
		return hv.Table()
	}
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	parts := strings.Split(t.Name(), ".")
	return strings.ToLower(parts[len(parts)-1]) + "s"
}

// Columns defines a type columns list, if not implemented sql will fallback to reflection based approach
type Columns interface {
	Columns() []string
}

func columnsOf(v interface{}) []string {
	hv, isHasColumns := v.(Columns)
	if isHasColumns {
		return hv.Columns()
	}
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	columns := []string{}

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if tag, exists := f.Tag.Lookup("bind"); exists {
			columns = append(columns, tag)
		} else {
			columns = append(columns, f.Name)
		}
	}
	return columns
}

// PKColumn defines a type PK column name, if not implemented sql will fallback to reflection based approach
type PKColumn interface {
	PKColumn() string
}

func primaryKeyOf(v interface{}) string {
	hv, isHasPK := v.(PKColumn)
	if isHasPK {
		return hv.PKColumn()
	}
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	for i := 0; i < t.NumField(); i++ {
		if tag, exists := t.Field(i).Tag.Lookup("pk"); exists {
			if tag == "true" {
				if name, exist := t.Field(i).Tag.Lookup("bind"); exist {
					return name
				}
				return t.Field(i).Name
			}
		}
	}
	return ""
}

type PKValue interface {
	PKValue() interface{}
}

func primaryKeyValue(v interface{}) interface{} {
	hv, isPKValue := v.(PKValue)
	if isPKValue {
		return hv.PKValue()
	}
	t := reflect.TypeOf(v)
	val := reflect.ValueOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		val = val.Elem()
	}
	for i := 0; i < t.NumField(); i++ {
		if tag, exists := t.Field(i).Tag.Lookup("pk"); exists {
			if tag == "true" {

				return val.Field(i).Interface()
			}
		}
	}
	return ""
}

type SetPKValue interface {
	SetPKValue(pk interface{})
}

func setPrimaryKeyFor(v interface{}, value interface{}) {
	hv, isSetPK := v.(SetPKValue)
	if isSetPK {
		hv.SetPKValue(value)
		return
	}
	t := reflect.TypeOf(v)
	val := reflect.ValueOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		val = val.Elem()
	}
	pkIdx := -1
	for i := 0; i < t.NumField(); i++ {
		if tag, exists := t.Field(i).Tag.Lookup("pk"); exists {
			if tag == "true" {
				pkIdx = i
			}
		}
	}
	ptr := reflect.NewAt(t.Field(pkIdx).Type, unsafe.Pointer(val.Field(pkIdx).UnsafeAddr())).Elem()
	toSetValue := reflect.ValueOf(value)
	if t.AssignableTo(ptr.Type()) {
		fmt.Println("no converting needed")
		ptr.Set(toSetValue)
	} else {
		if toSetValue.CanConvert(ptr.Type()) {
			ptr.Set(toSetValue.Convert(ptr.Type()))
		} else {
			panic(fmt.Sprintf("value of type %s is not assignable to %s and cannot convert also.", t, ptr.Type()))
		}
	}
}

type ToMap interface {
	ToMap() map[string]interface{}
}

func keyValueOf(obj interface{}) map[string]interface{} {
	m := map[string]interface{}{}
	hv, isKeyValue := obj.(ToMap)
	if isKeyValue {
		return hv.ToMap()
	}
	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		thisFieldValue := v.Field(i)
		if thisFieldValue.IsZero() {
			continue
		}
		if tag, exists := f.Tag.Lookup("bind"); exists {
			m[tag] = thisFieldValue.Interface()
		} else {
			m[f.Name] = thisFieldValue.Interface()
		}
	}
	return m
}

type Relations interface {
	Relations() []*ObjectMetadata
}

func relationsOf(obj interface{}) []*ObjectMetadata {
	r, is := obj.(Relations)
	if is {
		return r.Relations()
	}
	var mds []*ObjectMetadata
	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}
	for i := 0; i < t.NumField(); i++ {
		ft := t.Field(i)
		if rel, exists := ft.Tag.Lookup("fk"); exists && rel == "true" {
			mds = append(mds, ObjectMetadataFrom(v.Field(i).Interface()))
		}
	}
	return mds
}

type ObjectMetadata struct {
	// Name of the table that the object represents
	Table string
	// List of columns that this object has.
	Columns func(...string) []string
	// primary key of this struct
	PrimaryKey string
	Relations  []*ObjectMetadata
}

func ObjectMetadataFrom(v interface{}) *ObjectMetadata {
	return &ObjectMetadata{
		Table: ObjectHelpers.Table(v),
		Columns: func(blacklist ...string) []string {
			allColumns := ObjectHelpers.Columns(v)
			blacklisted := strings.Join(blacklist, ";")
			columns := []string{}
			for _, col := range allColumns {
				if !strings.Contains(blacklisted, col) {
					columns = append(columns, col)
				}
			}
			return columns
		},
		PrimaryKey: primaryKeyOf(v),
		Relations:  relationsOf(v),
	}
}
