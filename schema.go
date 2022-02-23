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
	Dialect    *querybuilder.Dialect
	Fields     []*Field
	GetPK      func(obj Entity) interface{}
	SetPK      func(obj Entity, value interface{})
	Values     func(obj Entity, withPK bool) []interface{}
}

func GetSchema[T Entity]() *Schema {
	v := new(T)
	return schemaOf(*v)
}
func (o *Schema) Columns(withPK bool) []string {
	var cols []string
	for _, field := range o.Fields {
		if field.Virtual {
			continue
		}
		if !withPK && field.IsPK {
			continue
		}
		if o.Dialect.AddTableNameInSelectColumns {
			cols = append(cols, o.Table+"."+field.Name)
		} else {
			cols = append(cols, field.Name)
		}
	}
	return cols
}

func (e *Schema) pkName() string {
	for _, field := range e.Fields {
		if field.IsPK {
			return field.Name
		}
	}
	return ""
}

type Field struct {
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
		if key == "name" {
			tag.Name = value
		} else if key == "pk" {
			tag.PK = true
		} else if key == "virtual" {
			tag.Virtual = true
		}
	}
	return tag
}

func genericFieldsOf(obj interface{}) []*Field {
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

	var fms []*Field
	for i := 0; i < t.NumField(); i++ {
		ft := t.Field(i)
		tagParsed := fieldMetadataFromTag(ft.Tag.Get("orm"))
		fm := &Field{}
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
func genericGetPkValue(o Entity, withPK bool) []interface{} {
	t := reflect.TypeOf(o)
	v := reflect.ValueOf(o)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}
	fields := o.Schema().Fields
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
	for i, field := range obj.Schema().Fields {
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

func getPkValue(obj Entity) interface{} {
	if obj.Schema().GetPK != nil {
		c := obj.Schema()
		return c.GetPK(obj)
	}

	t := reflect.TypeOf(obj)
	val := reflect.ValueOf(obj)
	if t.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	fields := obj.Schema().Fields
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
	// Fill user defined schemaOf members
	if userSchema.GetPK != nil {
		schema.GetPK = userSchema.GetPK
	}
	if userSchema.Connection != "" {
		schema.Connection = userSchema.Connection
	}
	if userSchema.Table != "" {
		schema.Table = userSchema.Table
	}
	if userSchema.SetPK != nil {
		schema.SetPK = userSchema.SetPK
	}
	if userSchema.Fields != nil {
		schema.Fields = userSchema.Fields
	}
	if userSchema.Values != nil {
		schema.Values = userSchema.Values
	}
	if userSchema.Dialect != nil {
		schema.Dialect = userSchema.Dialect
	}

	// Fill in the blanks
	if schema.GetPK == nil {
		schema.GetPK = getPkValue
	}

	if schema.SetPK == nil {
		schema.SetPK = genericSetPkValue
	}

	if schema.Values == nil {
		schema.Values = genericGetPkValue
	}

	if schema.Table == "" {
		schema.Table = initTableName(v)
	}

	if schema.Connection == "" {
		schema.Connection = "default"
	}
	if schema.Fields == nil {
		schema.Fields = genericFieldsOf(v)
	}

	return schema
}

func (e *Schema) getTable() string {
	return e.Table
}

func (e *Schema) getConnection() *sql.DB {
	return e.getDB().conn
}

func (e *Schema) getDialect() *querybuilder.Dialect {
	return e.Dialect
}

func (e *Schema) getDB() *DB {
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
	return s.getDB().getSchema(s.Table)
}
