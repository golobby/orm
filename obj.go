package orm

import (
	"fmt"
	"reflect"
	"strings"
	"unsafe"

	"github.com/golobby/orm/binder"
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
		name, exists := ft.Tag.Lookup("sqlname")
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
		if tag, exists := f.Tag.Lookup("sqlname"); exists {
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
				if name, exist := t.Field(i).Tag.Lookup("sqlname"); exist {
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
		if tag, exists := f.Tag.Lookup("sqlname"); exists {
			m[tag] = thisFieldValue.Interface()
		} else {
			m[f.Name] = thisFieldValue.Interface()
		}
	}
	return m
}

type ObjectMetadata struct {
	// Name of the table that the object represents
	Table  string
	Fields []*FieldMetadata
}

func (o *ObjectMetadata) Columns() []string {
	var cols []string
	for _, field := range o.Fields {
		cols = append(cols, field.SQLName)
	}
	return cols
}

type RelationType uint8

const (
	RelationTypeOneToOne = iota + 1
	RelationTypeOneToMany
	RelationTypeManyToOne
	RelationTypeManyToMany
)

func relationTypeFromStr(s string) RelationType {
	if s == "one2one" {
		return RelationTypeOneToOne
	} else if s == "one2many" {
		return RelationTypeOneToMany
	} else if s == "many2one" {
		return RelationTypeManyToOne
	} else if s == "many2many" {
		return RelationTypeManyToMany
	}
	panic("no relation type matched for " + s)
}

type RelationMetadata struct {
	Type        RelationType
	Table       string
	LeftColumn  string
	RightColumn string
}

type FieldMetadata struct {
	SQLName          string
	IsRel            bool
	RelationMetadata *RelationMetadata
}
type HasFields interface {
	Fields() []*FieldMetadata
}

func fieldsOf(obj interface{}) []*FieldMetadata {
	hasFields, is := obj.(HasFields)
	if is {
		return hasFields.Fields()
	}
	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}
	var fms []*FieldMetadata
	for i := 0; i < t.NumField(); i++ {
		ft := t.Field(i)
		fm := &FieldMetadata{}
		// Resolve this field name in database table
		if sqlName, exists := ft.Tag.Lookup("sqlname"); exists {
			fm.SQLName = sqlName
		} else {
			fm.SQLName = ft.Name
		}
		if _, exists := ft.Tag.Lookup("rel"); exists {
			fm.IsRel = true
			fm.RelationMetadata = &RelationMetadata{}
			if table, exists := ft.Tag.Lookup("foreigntable"); exists {
				fm.RelationMetadata.Table = table
			} else {
				// if no tag use fields own name as right table name
				fm.RelationMetadata.Table = ft.Name
			}
			if typ, exists := ft.Tag.Lookup("reltype"); exists {
				fm.RelationMetadata.Type = relationTypeFromStr(typ)
			} else {
				panic("cannot infer relation type yet for " + ft.Name)
			}
			if leftCol, exists := ft.Tag.Lookup("left"); exists {
				fm.RelationMetadata.LeftColumn = leftCol
			} else {
				fm.RelationMetadata.LeftColumn = ft.Name
			}
			if rightCol, exists := ft.Tag.Lookup("right"); exists {
				fm.RelationMetadata.RightColumn = rightCol
			} else {
				panic("cannot infer right side of join yet for " + ft.Name)
			}
		}
		fms = append(fms, fm)
	}
	return fms
}

func ObjectMetadataFrom(v interface{}) *ObjectMetadata {
	return &ObjectMetadata{
		Table:  ObjectHelpers.Table(v),
		Fields: fieldsOf(v),
	}
}
