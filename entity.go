package orm

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"unsafe"
)

type entity struct {
	obj IsEntity
}

func Entity(obj IsEntity) *entity {
	return &entity{obj: obj}
}
func (e *entity) getMetadata() *objectMetadata {
	db := e.getDB()
	return db.metadatas[e.getTable()]
}

func (e *entity) getTable() string {
	return e.obj.EntityConfig().Table
}

func (e *entity) getConnection() *sql.DB {
	return e.getDB().conn
}

func (e *entity) getDialect() *dialect {
	return e.getMetadata().dialect
}

func (e *entity) getDB() *DB {
	if len(globalORM) > 1 && (e.obj.EntityConfig().Connection == "" || e.getTable() == "") {
		panic("need table and connection name when having more than 1 connection registered")
	}
	if len(globalORM) == 1 {
		for _, db := range globalORM {
			return db
		}
	}
	if db, exists := globalORM[fmt.Sprintf("%s", e.obj.EntityConfig().Connection)]; exists {
		return db
	}
	panic("no db found")

}

func (e *entity) getFields() []*Field {
	var fields []*Field
	if e.obj.EntityConfig().GetFields != nil {
		fields = e.obj.EntityConfig().GetFields()
	} else {
		fields = e.getMetadata().Fields
	}
	return fields
}

type IsEntity interface {
	EntityConfig() EntityConfig
}

func (e *entity) getValues(o interface{}, withPK bool) []interface{} {
	if e.obj.EntityConfig().Values != nil {
		return e.obj.EntityConfig().Values(o, withPK)
	}
	t := reflect.TypeOf(o)
	v := reflect.ValueOf(o)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}
	fields := e.getFields()
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

// Fill the struct
func (e *entity) Fill() error {
	var q string
	var args []interface{}
	var err error
	pkValue := e.getPkValue(e.obj)
	ph := e.getDialect().PlaceholderChar
	if e.getDialect().IncludeIndexInPlaceholder {
		ph = ph + "1"
	}
	builder := newSelect().
		Select(e.getMetadata().Columns(true)...).
		From(e.getMetadata().Table).
		Where(WhereHelpers.Equal(e.getMetadata().pkName(), ph)).
		WithArgs(pkValue)
	q, args = builder.
		Build()
	rows, err := e.getConnection().Query(q, args...)
	if err != nil {
		return err
	}
	return e.getMetadata().Bind(rows, e.obj)
}

func (e *entity) SelectBuilder() *selectStmt {
	return newSelect().From(e.getTable()).Select(e.getMetadata().Columns(true)...)
}

func (e *entity) InsertBuilder() *insertStmt {
	return newInsert().Table(e.getTable()).Into(e.getMetadata().Columns(true)...)
}

func (e *entity) UpdateBuilder() *updateStmt {
	return newUpdate().Table(e.getTable())
}

func (e *entity) DeleteBuilder() *deleteStmt {
	return newDelete().Table(e.getTable())
}

// Save given object
func (e *entity) Save() error {
	cols := e.getMetadata().Columns(false)
	values := e.getValues(e.obj, false)
	var phs []string
	if e.getDialect().PlaceholderChar == "$" {
		phs = PlaceHolderGenerators.Postgres(len(cols))
	} else {
		phs = PlaceHolderGenerators.MySQL(len(cols))
	}
	q, args := newInsert().
		Table(e.getTable()).
		Into(cols...).
		Values(phs...).
		WithArgs(values...).Build()

	res, err := e.getConnection().Exec(q, args...)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	e.setPkValue(e.obj, id)
	return nil
}

// Update object in database
func (e *entity) Update() error {
	ph := e.getDialect().PlaceholderChar
	if e.getDialect().IncludeIndexInPlaceholder {
		ph = ph + "1"
	}
	counter := 2
	kvs := e.toMap(e.obj)
	var kvsWithPh []keyValue
	var args []interface{}
	for _, kv := range kvs {
		thisPh := e.getDialect().PlaceholderChar
		if e.getDialect().IncludeIndexInPlaceholder {
			thisPh += fmt.Sprint(counter)
		}
		kvsWithPh = append(kvsWithPh, keyValue{Key: kv.Key, Value: thisPh})
		args = append(args, kv.Value)
		counter++
	}
	query := WhereHelpers.Equal(e.getMetadata().pkName(), ph)
	q, args := newUpdate().
		Table(e.getTable()).
		Where(query).WithArgs(e.getPkValue(e.obj)).
		Set(kvsWithPh...).WithArgs(args...).
		Build()
	_, err := e.getConnection().Exec(q, args...)
	return err
}

// Delete the object from database
func (e *entity) Delete() error {
	ph := e.getDialect().PlaceholderChar
	if e.getDialect().IncludeIndexInPlaceholder {
		ph = ph + "1"
	}
	query := WhereHelpers.Equal(e.getMetadata().pkName(), ph)
	q, args := newDelete().
		Table(e.getTable()).
		Where(query).
		WithArgs(e.getPkValue(e.obj)).
		Build()
	_, err := e.getConnection().Exec(q, args...)
	return err
}

func (e *entity) BindContext(ctx context.Context, out interface{}, q string, args ...interface{}) error {
	rows, err := e.getConnection().QueryContext(ctx, q, args...)
	if err != nil {
		return err
	}
	return e.getMetadata().Bind(rows, out)
}

type EntityConfig[T any, PK any] struct {
	Table      string
	Connection string
	SetPK      func(obj T, pk interface{})
	GetPK      func(obj interface{}) PK
	GetFields  func() []*Field
	Values     func(obj T, withPK bool) []interface{}
}

func (o *objectMetadata) pkName() string {
	for _, field := range o.Fields {
		if field.IsPK {
			return field.Name
		}
	}
	return ""
}

func (e *entity) getPkValue(v interface{}) interface{} {

	if e.obj.EntityConfig().GetPK != nil {
		c := e.obj.EntityConfig()
		c.GetPK(e.obj)
	}

	t := reflect.TypeOf(v)
	val := reflect.ValueOf(v)
	if t.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	fields := e.getFields()
	for i, field := range fields {
		if field.IsPK {
			return val.Field(i).Interface()
		}
	}
	return ""
}

type SetPKValue interface {
	SetPKValue(pk interface{})
}

func (e *entity) setPkValue(v interface{}, value interface{}) {
	if e.obj.EntityConfig().SetPK != nil {
		e.obj.EntityConfig().SetPK(v, value)
		return
	}
	t := reflect.TypeOf(v)
	val := reflect.ValueOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		val = val.Elem()
	}
	pkIdx := -1
	for i, field := range e.getFields() {
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

func (e *entity) toMap(obj interface{}) []keyValue {
	var kvs []keyValue
	vs := e.getValues(obj, true)
	cols := e.getMetadata().Columns(true)
	for i, col := range cols {
		kvs = append(kvs, keyValue{
			Key:   col,
			Value: vs[i],
		})
	}
	return kvs
}

type objectMetadata struct {
	// DriverName of the table that the object represents
	Table   string
	Type    reflect.Type
	dialect *dialect
	Fields  []*Field
}

func (o *objectMetadata) Columns(withPK bool) []string {
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

func objectMetadataFrom(v IsEntity, dialect *dialect) *objectMetadata {
	e := &entity{obj: v}
	if len(e.getTable()) == 0 {
		panic("You should specify table name.")
	}
	return &objectMetadata{
		Table:   e.getTable(),
		dialect: dialect,
		Type:    reflect.TypeOf(v),
		Fields:  fieldsOf(v, dialect),
	}
}
