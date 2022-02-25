package orm

import (
	"database/sql"
	"fmt"
	"github.com/golobby/orm/querybuilder"
	"github.com/iancoleman/strcase"
	"reflect"
	"strings"
	"unsafe"
)

type Schema struct {
	Connection string
	Table      string
	dialect    *querybuilder.Dialect
	fields     []*field
	pkOffset   uintptr
	pkType     string
}

func GetSchema[T Entity]() *Schema {
	v := new(T)
	return (*v).Schema().Get()
}
func (o *Schema) Columns(withPK bool) []string {
	var cols []string
	for _, field := range o.fields {
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

func (e *Schema) pkName() string {
	for _, field := range e.fields {
		if field.IsPK {
			return field.Name
		}
	}
	return ""
}

type field struct {
	Name    string
	IsPK    bool
	Virtual bool
	Type    reflect.Type
}

type fieldTag struct {
	Name    string
	Virtual bool
	PK      bool
}

func fieldMetadataFromTag(t string) fieldTag {
	if t == "" {
		return fieldTag{}
	}
	tuples := strings.Split(t, " ")
	var tag fieldTag
	kv := map[string]string{}
	for _, tuple := range tuples {
		parts := strings.Split(tuple, "=")
		key := parts[0]
		value := parts[1]
		kv[key] = value
		if key == "dbCol" {
			tag.Name = value
		} else if key == "dbPK" {
			tag.PK = true
		}
		if tag.Name == "_" {
			tag.Virtual = true
		}
	}
	return tag
}

func genericFieldsOf(obj interface{}) []*field {
	t := reflect.TypeOf(obj)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()

	}
	if t.Kind() == reflect.Slice {
		t = t.Elem()
		for t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
	}

	var fms []*field
	for i := 0; i < t.NumField(); i++ {
		ft := t.Field(i)
		tagParsed := fieldMetadataFromTag(ft.Tag.Get("orm"))
		fm := &field{}
		fm.Type = ft.Type
		if tagParsed.Name != "" {
			fm.Name = tagParsed.Name
		} else {
			fm.Name = strcase.ToSnake(ft.Name)
		}
		if tagParsed.PK || strings.ToLower(ft.Name) == "id" {
			fm.IsPK = true
		}
		if tagParsed.Virtual || ft.Type.Kind() == reflect.Struct || ft.Type.Kind() == reflect.Slice || ft.Type.Kind() == reflect.Ptr {
			fm.Virtual = true
		}
		fms = append(fms, fm)
	}
	return fms
}
func genericValuesOf(o Entity, withPK bool) []interface{} {
	t := reflect.TypeOf(o)
	v := reflect.ValueOf(o)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}
	fields := o.Schema().Get().fields
	pkIdx := -1
	for i, field := range fields {
		if field.IsPK {
			pkIdx = i
		}

	}

	var values []interface{}

	for i := 0; i < t.NumField(); i++ {
		if !withPK && i == pkIdx {
			continue
		}
		if fields[i].Virtual {
			continue
		}
		values = append(values, v.Field(i).Interface())
	}
	return values
}

func genericSetPkValue(obj Entity, value interface{}) {
	t := reflect.TypeOf(obj)
	val := reflect.ValueOf(obj)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		val = val.Elem()
	}
	pkIdx := -1
	for i, field := range obj.Schema().Get().fields {
		if field.IsPK {
			pkIdx = i
		}
	}
	ptr := reflect.NewAt(t.Field(pkIdx).Type, unsafe.Pointer(val.Field(pkIdx).UnsafeAddr())).Elem()
	toSetValue := reflect.ValueOf(value)
	if t.Field(pkIdx).Type.AssignableTo(ptr.Type()) {
		ptr.Set(toSetValue)
	} else {
		if t.Field(pkIdx).Type.ConvertibleTo(ptr.Type()) {
			ptr.Set(toSetValue.Convert(ptr.Type()))
		} else {
			panic(fmt.Sprintf("value of type %s is not assignable to %s", t.Field(pkIdx).Type.String(), ptr.Type()))
		}
	}
}

func genericGetPKValue(obj Entity) interface{} {
	t := reflect.TypeOf(obj)
	val := reflect.ValueOf(obj)
	if t.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	fields := obj.Schema().Get().fields
	for i, field := range fields {
		if field.IsPK {
			return val.Field(i).Interface()
		}
	}
	return ""
}

func schemaOf(v Entity) *Schema {
	userSchema := v.Schema()
	schema := &Schema{}
	if userSchema.Connection != "" {
		schema.Connection = userSchema.Connection
	}
	if userSchema.Table != "" {
		schema.Table = userSchema.Table
	}

	if userSchema.fields != nil {
		schema.fields = userSchema.fields
	}

	if userSchema.dialect != nil {
		schema.dialect = userSchema.dialect
	}

	if schema.Table == "" {
		schema.Table = initTableName(v)
	}

	if schema.Connection == "" {
		schema.Connection = "default"
	}
	if schema.fields == nil {
		schema.fields = genericFieldsOf(v)
	}
	var pkIDX int
	for idx, f := range schema.fields {
		if f.IsPK {
			pkIDX = idx
		}
	}
	typ := reflect.TypeOf(v)
	if typ.Kind() == reflect.Ptr || typ.Kind() == reflect.Interface {
		typ = typ.Elem()
	}
	schema.pkOffset = typ.Field(pkIDX).Offset
	schema.pkType = typ.Field(pkIDX).Type.String()
	return schema
}

func (e *Schema) getTable() string {
	return e.Table
}

func (e *Schema) getSQLDB() *sql.DB {
	return e.getConnection().Connection
}

func (e *Schema) getDialect() *querybuilder.Dialect {
	return e.dialect
}

func (e *Schema) getConnection() *Connection {
	if len(globalORM) > 1 && (e.Connection == "" || e.getTable() == "") {
		panic("need Table and Connection name when having more than 1 Connection registered")
	}
	if len(globalORM) == 1 {
		for _, db := range globalORM {
			return db
		}
	}
	if db, exists := globalORM[fmt.Sprintf("%s", e.Connection)]; exists {
		return db
	}
	panic("no db found")

}

func (s *Schema) Get() *Schema {
	return s.getConnection().getSchema(s.Table)
}
