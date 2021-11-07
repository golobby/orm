package orm

import (
	"reflect"
	"strings"
	"unsafe"
)

type objectHelpers struct {
	// Returns a list of string which are the columns that a struct repreasent based on binder tags.
	// Best usage would be to generate these column names in startup.
	ColumnsOf func(v interface{}) []string
	// Returns a string which is the table name ( by convention is TYPEs ) of given object
	TableName func(v interface{}) string
	// Returns a list of values of the given object, useful for passing as args of sql exec or query
	ValuesOf func(v interface{}) []interface{}
	// Returns the primary key for given object.
	PrimaryKeyOf func(v interface{}) string
	// Sets primary key for object
	SetPK func(obj interface{}, pk interface{})
	// Gets value of primary key of given obj
	PKValue func(obj interface{}) interface{}
	// Returns a Key-Value paired of struct.
	KeyValue func(obj interface{}) map[string]interface{}
}

// ObjectHelpers are set of functions that extract type informations from a struct, it's better to use `ObjectMetadata`
var ObjectHelpers = &objectHelpers{
	ColumnsOf:    columnsOf,
	TableName:    tableName,
	ValuesOf:     valuesOf,
	PrimaryKeyOf: primaryKeyOf,
	SetPK:        setPrimaryKeyFor,
	PKValue:      primaryKeyValue,
	KeyValue:     keyValueOf,
}

// HasValues defines how a type should return it's values for sql arguments, if not implemented sql will fallback to reflection based approach
type HasValues interface {
	Values() []interface{}
}

func valuesOf(o interface{}) []interface{} {
	hv, isHasValues := o.(HasValues)
	if isHasValues {
		return hv.Values()
	}
	v := reflect.ValueOf(o)
	if v.Type().Kind() == reflect.Ptr {
		v = v.Elem()
	}
	values := []interface{}{}
	for i := 0; i < v.NumField(); i++ {
		values = append(values, v.Field(i).Interface())
	}
	return values
}

// HasTableName defines how a type should return it's coresponding table name, if not implemented sql will fallback to reflection based approach
type HasTableName interface {
	TableName() string
}

func tableName(v interface{}) string {
	hv, isTableName := v.(HasTableName)
	if isTableName {
		return hv.TableName()
	}
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	parts := strings.Split(t.Name(), ".")
	return strings.ToLower(parts[len(parts)-1]) + "s"
}

// HasColumns defines a type columns list, if not implemented sql will fallback to reflection based approach
type HasColumns interface {
	Columns() []string
}

func columnsOf(v interface{}) []string {
	hv, isHasColumns := v.(HasColumns)
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

// HasPK defines a type PK column name, if not implemented sql will fallback to reflection based approach
type HasPK interface {
	PK() string
}

func primaryKeyOf(v interface{}) string {
	hv, isHasPK := v.(HasPK)
	if isHasPK {
		return hv.PK()
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

type SetPK interface {
	SetPK(pk interface{})
}

func setPrimaryKeyFor(v interface{}, value interface{}) {
	hv, isSetPK := v.(SetPK)
	if isSetPK {
		hv.SetPK(value)
		return
	}
	t := reflect.TypeOf(v)
	val := reflect.ValueOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	pkIdx := -1
	for i := 0; i < t.NumField(); i++ {
		if tag, exists := t.Field(i).Tag.Lookup("pk"); exists {
			if tag == "true" {
				pkIdx = i
			}
		}
	}
	ptr := reflect.NewAt(t.Field(pkIdx).Type, unsafe.Pointer(val.Field(pkIdx).UnsafeAddr()))
	ptr.Set(reflect.ValueOf(value))
}

type KeyValue interface {
	KeyValue() map[string]interface{}
}

func keyValueOf(obj interface{}) map[string]interface{} {
	m := map[string]interface{}{}
	hv, isKeyValue := obj.(KeyValue)
	if isKeyValue {
		return hv.KeyValue()
	}
	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if tag, exists := f.Tag.Lookup("bind"); exists {
			m[tag] = v.Field(i)
		} else {
			m[f.Name] = v.Field(i)
		}
	}
	return m
}

type ObjectMetadata struct {
	// Name of the table that the object represents
	Table string
	// List of columns that this object has.
	Columns func(...string) []string
	// primary key of this struct
	PrimaryKey string
	// index of the relation fields
	RelationField map[string]int
}

func ObjectMetadataFrom(v interface{}) *ObjectMetadata {
	return &ObjectMetadata{
		Table: ObjectHelpers.TableName(v),
		Columns: func(blacklist ...string) []string {
			allColumns := ObjectHelpers.ColumnsOf(v)
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
	}
}
