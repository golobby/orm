package orm

import (
	"fmt"
	"github.com/golobby/orm/ds"
	"reflect"
	"strings"
	"unsafe"
)

type Entity interface {
	PKValue
	PKColumn
	SetPKValue
	Table
	ToMap
	InsertColumnsAndValues
	FromRows
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

// PKColumn defines a type PK column name, if not implemented sql will fallback to reflection based approach
type PKColumn interface {
	PKColumn() string
}

func pkName(v interface{}) string {
	hv, isHasPK := v.(PKColumn)
	if isHasPK {
		return hv.PKColumn()
	}
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	for i := 0; i < t.NumField(); i++ {
		ft := t.Field(i)
		if orm, exists := ft.Tag.Lookup("orm"); exists {
			m := fieldMetadataFromTag(orm)
			if m.PK {
				if m.Name == "" {
					return ft.Name
				} else {
					return m.Name
				}
			}
		}
	}
	return ""
}

type PKValue interface {
	PKValue() interface{}
}

func pkValue(v interface{}) interface{} {
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
		ft := t.Field(i)
		if orm, exists := ft.Tag.Lookup("orm"); exists {
			m := fieldMetadataFromTag(orm)
			if m.PK {
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
		ft := t.Field(i)
		if orm, exists := ft.Tag.Lookup("orm"); exists {
			m := fieldMetadataFromTag(orm)
			if m.PK {
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

func keyValueOf(obj interface{}) []ds.KV {
	var kvs []ds.KV
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
			kvs = append(kvs, ds.KV{
				Key:   tag,
				Value: thisFieldValue.Interface(),
			})
		} else {
			kvs = append(kvs, ds.KV{
				Key:   f.Name,
				Value: thisFieldValue.Interface(),
			})
		}
	}
	return kvs
}

type ObjectMetadata struct {
	// Name of the table that the object represents
	Table  string
	Fields []*FieldMetadata
}

func (o *ObjectMetadata) Columns() []string {
	var cols []string
	for _, field := range o.Fields {
		if field.IsRel {
			continue
		}
		cols = append(cols, o.Table+"."+field.SQLName)
	}
	return cols
}

type RelationType uint8

const (
	RelationTypeHasOne = iota + 1
	RelationTypeHasMany
)

func relationTypeFromStr(s string) RelationType {
	if s == "1" || s == "one" {
		return RelationTypeHasOne
	} else if s == "n" || s == "many" {
		return RelationTypeHasMany
	}
	panic("no relation type matched for " + s)
}

type RelationMetadata struct {
	Type           RelationType
	Table          string
	LeftColumn     string
	RightColumn    string
	objectMetadata *ObjectMetadata
}

type FieldMetadata struct {
	SQLName          string
	IsPK             bool
	IsRel            bool
	RelationMetadata *RelationMetadata
}

type FieldTag struct {
	Name  string
	PK    bool
	InRel bool
	With  string
	Left  string
	Right string
}

type HasFields interface {
	Fields() []*FieldMetadata
}

func fieldMetadataFromTag(t string) FieldTag {
	if t == "" {
		return FieldTag{}
	}
	tuples := strings.Split(t, " ")
	var tag FieldTag
	kv := map[string]string{}
	for _, tuple := range tuples {
		parts := strings.Split(tuple, "=")
		key := parts[0]
		value := parts[1]
		kv[key] = value
		if key == "name" {
			tag.Name = value
		} else if key == "in_rel" {
			tag.InRel = value == "true"
		} else if key == "with" {
			tag.With = value
		} else if key == "left" {
			tag.Left = value
		} else if key == "right" {
			tag.Right = value
		} else if key == "pk" {
			tag.PK = true
		}
	}
	return tag
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
		tagParsed := fieldMetadataFromTag(ft.Tag.Get("orm"))
		fm := &FieldMetadata{}
		if tagParsed.Name != "" {
			fm.SQLName = tagParsed.Name
		} else {
			fm.SQLName = ft.Name
		}
		if tagParsed.PK {
			fm.IsPK = true
		}
		if tagParsed.InRel == true {
			fm.IsRel = true

			fm.RelationMetadata = &RelationMetadata{}
			fm.RelationMetadata.objectMetadata = ObjectMetadataFrom(reflect.New(ft.Type).Interface())

			if tagParsed.With != "" {
				fm.RelationMetadata.Table = tagParsed.With
			}
			if tagParsed.Left != "" {
				fm.RelationMetadata.LeftColumn = tagParsed.Left
			}
			if tagParsed.Right != "" {
				fm.RelationMetadata.RightColumn = tagParsed.Right
			}
		}
		fms = append(fms, fm)
	}
	return fms
}

func ObjectMetadataFrom(v interface{}) *ObjectMetadata {
	return &ObjectMetadata{
		Table:  tableName(v),
		Fields: fieldsOf(v),
	}
}
