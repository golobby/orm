package orm

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"reflect"
)

func getConnectionFor(e Entity) *connection {
	configurator := newEntityConfigurator()
	e.ConfigureEntity(configurator)

	if len(globalConnections) > 1 && (configurator.connection == "" || configurator.table == "") {
		panic("need table and DB name when having more than 1 DB registered")
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
		s = schemaOfHeavyReflectionStuff(e)
		c.setSchema(e, s)
	}
	return s
}

type schema struct {
	Connection        string
	Table             string
	fields            []*field
	relations         map[string]interface{}
	setPK             func(o Entity, value interface{})
	getPK             func(o Entity) interface{}
	columnConstraints []*FieldConfigurator
}

func (s *schema) getField(sf reflect.StructField) *field {
	for _, f := range s.fields {
		if sf.Name == f.Name {
			return f
		}
	}
	return nil
}

func (s *schema) getDialect() *Dialect {
	return GetConnection(s.Connection).Dialect
}
func (s *schema) Columns(withPK bool) []string {
	var cols []string
	for _, field := range s.fields {
		if field.Virtual {
			continue
		}
		if !withPK && field.IsPK {
			continue
		}
		if s.getDialect().AddTableNameInSelectColumns {
			cols = append(cols, s.Table+"."+field.Name)
		} else {
			cols = append(cols, field.Name)
		}
	}
	return cols
}

func (s *schema) pkName() string {
	for _, field := range s.fields {
		if field.IsPK {
			return field.Name
		}
	}
	return ""
}

func genericFieldsOf(obj Entity) []*field {
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
	var ec EntityConfigurator
	obj.ConfigureEntity(&ec)

	var fms []*field
	for i := 0; i < t.NumField(); i++ {
		ft := t.Field(i)
		fm := fieldMetadata(ft, ec.columnConstraints)
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
			// go into
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
			fm := fieldMetadata(actualV.Type().Field(i), nil)[0]
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
func schemaOfHeavyReflectionStuff(v Entity) *schema {
	userEntityConfigurator := newEntityConfigurator()
	v.ConfigureEntity(userEntityConfigurator)
	for _, relation := range userEntityConfigurator.resolveRelations {
		relation()
	}
	schema := &schema{}
	if userEntityConfigurator.connection != "" {
		schema.Connection = userEntityConfigurator.connection
	}
	if userEntityConfigurator.table != "" {
		schema.Table = userEntityConfigurator.table
	} else {
		panic("you need to have table name for getting schema.")
	}

	schema.columnConstraints = userEntityConfigurator.columnConstraints
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

	schema.relations = userEntityConfigurator.relations

	return schema
}

func (s *schema) getTable() string {
	return s.Table
}

func (s *schema) getSQLDB() *sql.DB {
	return s.getConnection().DB
}

func (s *schema) getConnection() *connection {
	if len(globalConnections) > 1 && (s.Connection == "" || s.Table == "") {
		panic("need table and DB name when having more than 1 DB registered")
	}
	if len(globalConnections) == 1 {
		for _, db := range globalConnections {
			return db
		}
	}
	if db, exists := globalConnections[fmt.Sprintf("%s", s.Connection)]; exists {
		return db
	}
	panic("no db found")
}
