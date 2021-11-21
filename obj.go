package orm

import (
	"fmt"
	"reflect"
	"strings"
	"unsafe"

	"github.com/gertd/go-pluralize"

	"github.com/iancoleman/strcase"
)

// IEntity interface is for sake of documentation, if you want to change orm behaviour for:
// Table name generation -> implement Table for your model
// GetPKValue -> returns value of primary key of model, implementing this helps with performance.
// SetPKValue -> sets the value of primary key of mode, implementing this helps with performance.
type IEntity interface {
	Table
	GetPKValue
	SetPKValue
	Values
}

// Values returns a slice containing all values of current object to be used in insert or updates.
type Values interface {
	Values() []interface{}
}

func (s *Repository) valuesOf(o interface{}, withPK bool) []interface{} {
	vls, is := o.(Values)
	if is {
		return vls.Values()
	}
	t := reflect.TypeOf(o)
	v := reflect.ValueOf(o)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}
	pkIdx := -1
	for i, field := range s.metadata.Fields {
		if field.IsPK {
			pkIdx = i
		}

	}

	var values []interface{}

	for i := 0; i < t.NumField(); i++ {
		if !withPK && i == pkIdx {
			continue
		}
		if s.metadata.Fields[i].Virtual {
			continue
		}
		values = append(values, v.Field(i).Interface())
	}
	return values
}

// Table defines how a type should return it's coresponding table name, if not implemented sql will fallback to reflection based approach (plural snake case).
type Table interface {
	Table() string
}

func tableName(v interface{}) string {
	if s, ok := v.(string); ok {
		return strcase.ToSnake(pluralize.NewClient().Plural(s))
	}
	hv, isTableName := v.(Table)
	if isTableName {
		return hv.Table()
	}
	t := reflect.TypeOf(v)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	parts := strings.Split(t.Name(), ".")
	name := parts[len(parts)-1]
	return strcase.ToSnake(pluralize.NewClient().Plural(name))
}

func (o *ObjectMetadata) pkName() string {
	for _, field := range o.Fields {
		if field.IsPK {
			return field.Name
		}
	}
	return ""
}

type GetPKValue interface {
	PKValue() interface{}
}

func (s *Repository) getPkValue(v interface{}) interface{} {
	hv, isPKValue := v.(GetPKValue)
	if isPKValue {
		return hv.PKValue()
	}
	t := reflect.TypeOf(v)
	val := reflect.ValueOf(v)
	if t.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	for i, field := range s.metadata.Fields {
		if field.IsPK {
			return val.Field(i).Interface()
		}
	}
	return ""
}

type SetPKValue interface {
	SetPKValue(pk interface{})
}

func (s *Repository) setPkValue(v interface{}, value interface{}) {
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
	for i, field := range s.metadata.Fields {
		if field.IsPK {
			pkIdx = i
		}
	}
	ptr := reflect.NewAt(t.Field(pkIdx).Type, unsafe.Pointer(val.Field(pkIdx).UnsafeAddr())).Elem()
	toSetValue := reflect.ValueOf(value)
	if t.Field(pkIdx).Type.AssignableTo(ptr.Type()) {
		ptr.Set(toSetValue)
	} else {
		panic(fmt.Sprintf("value of type %s is not assignable to %s", t.Field(pkIdx).Type.String(), ptr.Type()))
	}
}

func (s *Repository) toMap(obj interface{}) []KV {
	var kvs []KV
	vs := s.valuesOf(obj, true)
	cols := s.metadata.Columns(true)
	for i, col := range cols {
		kvs = append(kvs, KV{
			Key:   col,
			Value: vs[i],
		})
	}
	return kvs
}

type ObjectMetadata struct {
	// Name of the table that the object represents
	Table   string
	Type    reflect.Type
	dialect *Dialect
	Fields  []*FieldMetadata
}

func (o *ObjectMetadata) Columns(withPK bool) []string {
	var cols []string
	for _, field := range o.Fields {
		if field.Virtual {
			continue
		}
		if !withPK && field.IsPK {
			continue
		}
		if o.dialect.AddTableNameInSelectColumns {
			cols = append(cols, o.Table+"."+field.Name)
		} else {
			cols = append(cols, field.Name)
		}
	}
	return cols
}

func ObjectMetadataFrom(v interface{}, dialect *Dialect) *ObjectMetadata {
	return &ObjectMetadata{
		Table:   tableName(v),
		dialect: dialect,
		Type:    reflect.TypeOf(v),
		Fields:  fieldsOf(v, dialect),
	}
}
