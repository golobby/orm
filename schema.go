package orm

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"unsafe"

	"github.com/gertd/go-pluralize"
	"github.com/iancoleman/strcase"
)

type EntityConfigurator struct {
	connection       string
	table            string
	this             Entity
	relations        map[string]interface{}
	resolveRelations []func()
}

func newEntityConfigurator() *EntityConfigurator {
	return &EntityConfigurator{}
}

func (e *EntityConfigurator) Table(name string) *EntityConfigurator {
	e.table = name
	return e
}

func (r *EntityConfigurator) HasMany(property Entity, config HasManyConfig) *EntityConfigurator {
	if r.relations == nil {
		r.relations = map[string]interface{}{}
	}
	r.resolveRelations = append(r.resolveRelations, func() {
		if config.PropertyForeignKey != "" && config.PropertyTable != "" {
			r.relations[config.PropertyTable] = config
			return
		}
		configurator := newEntityConfigurator()
		property.ConfigureEntity(configurator)

		if config.PropertyTable == "" {
			config.PropertyTable = configurator.table
		}

		if config.PropertyForeignKey == "" {
			config.PropertyForeignKey = pluralize.NewClient().Singular(r.table) + "_id"
		}

		r.relations[configurator.table] = config

		return
	})
	return r
}

func (r *EntityConfigurator) HasOne(property Entity, config HasOneConfig) *EntityConfigurator {
	if r.relations == nil {
		r.relations = map[string]interface{}{}
	}
	r.resolveRelations = append(r.resolveRelations, func() {
		if config.PropertyForeignKey != "" && config.PropertyTable != "" {
			r.relations[config.PropertyTable] = config
			return
		}

		configurator := newEntityConfigurator()
		property.ConfigureEntity(configurator)

		if config.PropertyTable == "" {
			config.PropertyTable = configurator.table
		}
		if config.PropertyForeignKey == "" {
			config.PropertyForeignKey = pluralize.NewClient().Singular(r.table) + "_id"
		}

		r.relations[configurator.table] = config
		return
	})
	return r
}

func (r *EntityConfigurator) BelongsTo(owner Entity, config BelongsToConfig) *EntityConfigurator {
	if r.relations == nil {
		r.relations = map[string]interface{}{}
	}
	r.resolveRelations = append(r.resolveRelations, func() {
		if config.ForeignColumnName != "" && config.LocalForeignKey != "" && config.OwnerTable != "" {
			r.relations[config.OwnerTable] = config
			return
		}
		ownerConfigurator := newEntityConfigurator()
		owner.ConfigureEntity(ownerConfigurator)
		if config.OwnerTable == "" {
			config.OwnerTable = ownerConfigurator.table
		}
		if config.LocalForeignKey == "" {
			config.LocalForeignKey = pluralize.NewClient().Singular(ownerConfigurator.table) + "_id"
		}
		if config.ForeignColumnName == "" {
			config.ForeignColumnName = "id"
		}
		r.relations[ownerConfigurator.table] = config
	})
	return r
}

func (r *EntityConfigurator) BelongsToMany(owner Entity, config BelongsToManyConfig) *EntityConfigurator {
	if r.relations == nil {
		r.relations = map[string]interface{}{}
	}
	r.resolveRelations = append(r.resolveRelations, func() {
		ownerConfigurator := newEntityConfigurator()
		owner.ConfigureEntity(ownerConfigurator)

		if config.ForeignLookupColumn == "" {
			var pkName string
			for _, field := range genericFieldsOf(owner) {
				if field.IsPK {
					pkName = field.Name
				}
			}
			config.ForeignLookupColumn = pkName

		}
		if config.ForeignTable == "" {
			config.ForeignTable = ownerConfigurator.table
		}
		if config.IntermediateTable == "" {
			panic("cannot infer intermediate table yet")
		}
		if config.IntermediatePropertyID == "" {
			config.IntermediatePropertyID = pluralize.NewClient().Singular(ownerConfigurator.table) + "_id"
		}
		if config.IntermediateOwnerID == "" {
			config.IntermediateOwnerID = pluralize.NewClient().Singular(r.table) + "_id"
		}

		r.relations[ownerConfigurator.table] = config
	})
	return r
}

func getConnectionFor(e Entity) *Connection {
	configurator := newEntityConfigurator()
	e.ConfigureEntity(configurator)

	if len(globalORM) > 1 && (configurator.connection == "" || configurator.table == "") {
		panic("need table and Connection name when having more than 1 Connection registered")
	}
	if len(globalORM) == 1 {
		for _, db := range globalORM {
			return db
		}
	}
	if db, exists := globalORM[fmt.Sprintf("%s", configurator.connection)]; exists {
		return db
	}
	panic("no db found")
}

func getSchemaFor(e Entity) *schema {
	configurator := newEntityConfigurator()
	c := getConnectionFor(e)
	e.ConfigureEntity(configurator)
	return c.getSchema(configurator.table)
}

type schema struct {
	Connection string
	Table      string
	dialect    *Dialect
	fields     []*field
	relations  map[string]interface{}
	setPK      func(o Entity, value interface{})
	getPK      func(o Entity) interface{}
}

func (o *schema) Columns(withPK bool) []string {
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

func (e *schema) pkName() string {
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
		if key == "col" {
			tag.Name = value
		} else if key == "pk" {
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
	fields := getSchemaFor(o).fields
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
	for i, field := range getSchemaFor(obj).fields {
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

	fields := getSchemaFor(obj).fields
	for i, field := range fields {
		if field.IsPK {
			return val.Field(i).Interface()
		}
	}
	return ""
}

func schemaOf(v Entity) *schema {
	userSchema := newEntityConfigurator()
	v.ConfigureEntity(userSchema)
	for _, relation := range userSchema.resolveRelations {
		relation()
	}
	schema := &schema{}
	if userSchema.connection != "" {
		schema.Connection = userSchema.connection
	}
	if userSchema.table != "" {
		schema.Table = userSchema.table
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
	if schema.getPK == nil {
		schema.getPK = genericGetPKValue
	}

	if schema.setPK == nil {
		schema.setPK = genericSetPkValue
	}

	schema.relations = userSchema.relations

	return schema
}

func (e *schema) getTable() string {
	return e.Table
}

func (e *schema) getSQLDB() *sql.DB {
	return e.getConnection().Connection
}

func (e *schema) getDialect() *Dialect {
	return e.dialect
}

func (e *schema) getConnection() *Connection {
	if len(globalORM) > 1 && (e.Connection == "" || e.Table == "") {
		panic("need table and Connection name when having more than 1 Connection registered")
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
