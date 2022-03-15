package orm

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"github.com/gertd/go-pluralize"
	"github.com/iancoleman/strcase"
	"reflect"
	"strings"
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

func (e *EntityConfigurator) Connection(name string) *EntityConfigurator {
	e.connection = name
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

		if config.OwnerLookupColumn == "" {
			var pkName string
			for _, field := range genericFieldsOf(owner) {
				if field.IsPK {
					pkName = field.Name
				}
			}
			config.OwnerLookupColumn = pkName

		}
		if config.OwnerTable == "" {
			config.OwnerTable = ownerConfigurator.table
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

	if len(globalConnections) > 1 && (configurator.connection == "" || configurator.table == "") {
		panic("need table and Connection name when having more than 1 Connection registered")
	}
	if len(globalConnections) == 1 {
		for _, db := range globalConnections {
			return db
		}
	}
	if db, exists := globalConnections[fmt.Sprintf("%s", configurator.connection)]; exists {
		return db
	}
	panic("no db found")
}

func getSchemaFor(e Entity) *schema {
	configurator := newEntityConfigurator()
	c := getConnectionFor(e)
	e.ConfigureEntity(configurator)
	s := c.getSchema(configurator.table)
	if s == nil {
		s = schemaOf(e)
		c.setSchema(e, s)
	}
	return s
}

type schema struct {
	Connection string
	Table      string
	fields     []*field
	relations  map[string]interface{}
	setPK      func(o Entity, value interface{})
	getPK      func(o Entity) interface{}
}

func (s *schema) getDialect() *Dialect {
	return GetConnection(s.Connection).Dialect
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
		if o.getDialect().AddTableNameInSelectColumns {
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
	Name        string
	IsPK        bool
	Virtual     bool
	IsCreatedAt bool
	IsUpdatedAt bool
	IsDeletedAt bool
	Type        reflect.Type
}

type fieldTag struct {
	Name        string
	Virtual     bool
	PK          bool
	IsCreatedAt bool
	IsUpdatedAt bool
	IsDeletedAt bool
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
		} else if key == "created_at" {
			tag.IsCreatedAt = true
		} else if key == "updated_at" {
			tag.IsUpdatedAt = true
		} else if key == "deleted_at" {
			tag.IsDeletedAt = true
		}
		if tag.Name == "_" {
			tag.Virtual = true
		}
	}
	return tag
}
func fieldMetadata(ft reflect.StructField) []*field {
	tagParsed := fieldMetadataFromTag(ft.Tag.Get("orm"))
	var fms []*field
	baseFm := &field{}
	baseFm.Type = ft.Type
	fms = append(fms, baseFm)
	if tagParsed.Name != "" {
		baseFm.Name = tagParsed.Name
	} else {
		baseFm.Name = strcase.ToSnake(ft.Name)
	}
	if tagParsed.PK || strings.ToLower(ft.Name) == "id" {
		baseFm.IsPK = true
	}
	if tagParsed.IsCreatedAt || strings.ToLower(ft.Name) == "createdat" {
		baseFm.IsCreatedAt = true
	}
	if tagParsed.IsUpdatedAt || strings.ToLower(ft.Name) == "updatedat" {
		baseFm.IsUpdatedAt = true
	}
	if tagParsed.IsDeletedAt || strings.ToLower(ft.Name) == "deletedat" {
		baseFm.IsDeletedAt = true
	}
	if tagParsed.Virtual {
		baseFm.Virtual = true
	}
	if ft.Type.Kind() == reflect.Struct || ft.Type.Kind() == reflect.Ptr {
		t := ft.Type
		if ft.Type.Kind() == reflect.Ptr {
			t = ft.Type.Elem()
		}
		if !t.Implements(reflect.TypeOf((*driver.Valuer)(nil)).Elem()) {
			for i := 0; i < t.NumField(); i++ {
				fms = append(fms, fieldMetadata(t.Field(i))...)
			}
			fms = fms[1:]
		}
	}
	return fms
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
		fm := fieldMetadata(ft)
		fms = append(fms, fm...)
	}
	return fms
}

func valuesOfField(vf reflect.Value) []interface{} {
	var values []interface{}
	if vf.Type().Kind() == reflect.Struct || vf.Type().Kind() == reflect.Ptr {
		t := vf.Type()
		if vf.Type().Kind() == reflect.Ptr {
			t = vf.Type().Elem()
		}
		if !t.Implements(reflect.TypeOf((*driver.Valuer)(nil)).Elem()) {
			//go into
			// it does not implement driver.Valuer interface
			for i := 0; i < vf.NumField(); i++ {
				vif := vf.Field(i)
				values = append(values, valuesOfField(vif)...)
			}
		} else {
			values = append(values, vf.Interface())
		}
	} else {
		values = append(values, vf.Interface())
	}
	return values
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
		vf := v.Field(i)
		values = append(values, valuesOfField(vf)...)
	}
	return values
}

func genericSetPkValue(obj Entity, value interface{}) {
	genericSet(obj, getSchemaFor(obj).pkName(), value)
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

func (s *schema) createdAt() *field {
	for _, f := range s.fields {
		if f.IsCreatedAt {
			return f
		}
	}
	return nil
}
func (s *schema) updatedAt() *field {
	for _, f := range s.fields {
		if f.IsUpdatedAt {
			return f
		}
	}
	return nil
}

func (s *schema) deletedAt() *field {
	for _, f := range s.fields {
		if f.IsDeletedAt {
			return f
		}
	}
	return nil
}
func pointersOf(v reflect.Value) map[string]interface{} {
	m := map[string]interface{}{}
	actualV := v
	for actualV.Type().Kind() == reflect.Ptr {
		actualV = actualV.Elem()
	}
	for i := 0; i < actualV.NumField(); i++ {
		f := actualV.Field(i)
		if (f.Type().Kind() == reflect.Struct || f.Type().Kind() == reflect.Ptr) && !f.Type().Implements(reflect.TypeOf((*driver.Valuer)(nil)).Elem()) {
			fm := pointersOf(f)
			for k, p := range fm {
				m[k] = p
			}
		} else {
			fm := fieldMetadata(actualV.Type().Field(i))[0]
			m[fm.Name] = actualV.Field(i)
		}
	}

	return m
}
func genericSet(obj Entity, name string, value interface{}) {
	n2p := pointersOf(reflect.ValueOf(obj))
	var val interface{}
	for k, v := range n2p {
		if k == name {
			val = v
		}
	}
	val.(reflect.Value).Set(reflect.ValueOf(value))
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
	} else {
		panic("you need to have table name for getting schema.")
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

func (e *schema) getConnection() *Connection {
	if len(globalConnections) > 1 && (e.Connection == "" || e.Table == "") {
		panic("need table and Connection name when having more than 1 Connection registered")
	}
	if len(globalConnections) == 1 {
		for _, db := range globalConnections {
			return db
		}
	}
	if db, exists := globalConnections[fmt.Sprintf("%s", e.Connection)]; exists {
		return db
	}
	panic("no db found")
}
