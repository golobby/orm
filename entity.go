package orm

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/gertd/go-pluralize"
	"reflect"
	"unsafe"
)

type Entity[T any] struct {
	obj IsEntity
}

func AsEntity(obj IsEntity) *Entity {
	return &Entity{obj: obj}
}
func (e *Entity) getMetadata() *objectMetadata {
	db := e.getDB()
	return db.metadatas[e.getTable()]
}

func (e *Entity) getTable() string {
	return e.obj.EntityConfig().Table
}

func (e *Entity) getConnection() *sql.DB {
	return e.getDB().conn
}

func (e *Entity) getDialect() *Dialect {
	return e.getMetadata().dialect
}

func (e *Entity) getDB() *DB {
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

func (e *Entity) getFields() []*Field {
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

func (e *Entity) getValues(o interface{}, withPK bool) []interface{} {
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
func (e *Entity) Fill() error {
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

func (e *Entity) SelectBuilder() *selectStmt {
	return newSelect().From(e.getTable()).Select(e.getMetadata().Columns(true)...)
}

func (e *Entity) InsertBuilder() *insertStmt {
	return newInsert().Table(e.getTable()).Into(e.getMetadata().Columns(true)...)
}

func (e *Entity) UpdateBuilder() *updateStmt {
	return newUpdate().Table(e.getTable())
}

func (e *Entity) DeleteBuilder() *deleteStmt {
	return newDelete().Table(e.getTable())
}

// Save given object
func (e *Entity) Save() error {
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
func (e *Entity) Update() error {
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
func (e *Entity) Delete() error {
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

func (e *Entity) BindContext(ctx context.Context, out interface{}, q string, args ...interface{}) error {
	rows, err := e.getConnection().QueryContext(ctx, q, args...)
	if err != nil {
		return err
	}
	return e.getMetadata().Bind(rows, out)
}

// BaseEntity contains common behaviours for entity structs.
type BaseEntity[T any] struct{}

type Relation func(e *Entity) error

func (e *Entity) Load(r Relation) error {
	return r(e)
}

type HasManyConfig struct {
	PropertyTable      string
	PropertyForeignKey string
}

func (b BaseEntity) HasMany(out []IsEntity, config HasManyConfig) func(e *Entity) error {
	return func(e *Entity) error {
		return e.HasMany(out, config)
	}
}
func (e *Entity) HasMany(out []IsEntity, c HasManyConfig) error {
	outEntity := &Entity{obj: reflect.New(reflect.TypeOf(out).Elem()).Interface().(IsEntity)}
	//settings default config values
	if c.PropertyTable == "" {
		c.PropertyTable = outEntity.getTable()
	}
	if c.PropertyForeignKey == "" {
		c.PropertyForeignKey = pluralize.NewClient().Singular(e.getTable()) + "_id"
	}

	ph := e.getDialect().PlaceholderChar
	if e.getDialect().IncludeIndexInPlaceholder {
		ph = ph + fmt.Sprint(1)
	}
	var q string
	var args []interface{}

	q, args = newSelect().
		From(c.PropertyTable).
		Where(WhereHelpers.Equal(c.PropertyForeignKey, ph)).
		WithArgs(e.getPkValue(e.obj)).
		Build()

	if q == "" {
		return fmt.Errorf("cannot build the query")
	}

	return outEntity.BindContext(context.Background(), out, q, args...)
}

type HasOneConfig struct {
	PropertyTable      string
	PropertyForeignKey string
}

func (b BaseEntity) HasOne(output IsEntity, config HasOneConfig) func(e *Entity) error {
	return func(e *Entity) error {
		return e.HasOne(output, config)
	}
}

func (e *Entity) HasOne(out IsEntity, c HasOneConfig) error {
	outEntity := &Entity{obj: out}
	//settings default config values
	if c.PropertyTable == "" {
		c.PropertyTable = outEntity.getTable()
	}
	if c.PropertyForeignKey == "" {
		c.PropertyForeignKey = pluralize.NewClient().Singular(e.getTable()) + "_id"
	}

	ph := e.getDialect().PlaceholderChar
	if e.getDialect().IncludeIndexInPlaceholder {
		ph = ph + fmt.Sprint(1)
	}
	var q string
	var args []interface{}

	q, args = newSelect().
		From(c.PropertyTable).
		Where(WhereHelpers.Equal(c.PropertyForeignKey, ph)).
		WithArgs(e.getPkValue(e.obj)).
		Build()

	if q == "" {
		return fmt.Errorf("cannot build the query")
	}
	return e.BindContext(context.Background(), out, q, args...)
}

type BelongsToConfig struct {
	OwnerTable        string
	LocalForeignKey   string
	ForeignColumnName string
}

func (b BaseEntity) BelongsTo(output IsEntity, config BelongsToConfig) func(e *Entity) error {
	return func(e *Entity) error {
		return e.BelongsTo(output, config)
	}
}
func (e *Entity) BelongsTo(out IsEntity, c BelongsToConfig) error {
	outEntity := &Entity{obj: out}
	if c.OwnerTable == "" {
		c.OwnerTable = outEntity.getTable()
	}
	if c.LocalForeignKey == "" {
		c.LocalForeignKey = pluralize.NewClient().Singular(outEntity.getTable()) + "_id"
	}
	if c.ForeignColumnName == "" {
		c.ForeignColumnName = "id"
	}

	ph := e.getDialect().PlaceholderChar
	if e.getDialect().IncludeIndexInPlaceholder {
		ph = ph + "1"
	}
	ownerIDidx := 0
	for idx, field := range e.getFields() {
		if field.Name == c.LocalForeignKey {
			ownerIDidx = idx
		}
	}

	ownerID := e.getValues(e.obj, true)[ownerIDidx]

	q, args := newSelect().
		From(c.OwnerTable).
		Where(WhereHelpers.Equal(c.ForeignColumnName, ph)).
		WithArgs(ownerID).Build()

	return e.BindContext(context.Background(), out, q, args...)
}

type ManyToManyConfig struct {
	IntermediateTable         string
	IntermediateLocalColumn   string
	IntermediateForeignColumn string
	ForeignTable              string
	ForeignLookupColumn       string
}

func (b BaseEntity) ManyToMany(output []IsEntity, config ManyToManyConfig) func(e *Entity) error {
	return func(e *Entity) error {
		return e.ManyToMany(output, config)
	}
}

func (e *Entity) ManyToMany(out []IsEntity, c ManyToManyConfig) error {
	outEntity := &Entity{obj: reflect.New(reflect.TypeOf(out).Elem()).Interface().(IsEntity)}
	if c.IntermediateTable == "" {
		return fmt.Errorf("no way to infer many to many intermediate table yet.")
	}
	if c.IntermediateLocalColumn == "" {
		table := e.getTable()
		table = pluralize.NewClient().Singular(table)
		c.IntermediateLocalColumn = table + "_id"
	}
	if c.IntermediateForeignColumn == "" {
		table := outEntity.getTable()
		c.IntermediateForeignColumn = pluralize.NewClient().Singular(table) + "_id"
	}
	if c.ForeignTable == "" {
		c.IntermediateForeignColumn = outEntity.getTable()
	}
	//TODO: this logic is wrong
	sub, _ := newSelect().From(c.IntermediateTable).Where(c.IntermediateLocalColumn, "=", fmt.Sprint(e.getPkValue(e.obj))).Build()
	q, args := newSelect().
		From(c.ForeignTable).
		Where(c.ForeignLookupColumn, "in", sub).
		Build()

	return e.
		BindContext(context.Background(), out, q, args...)

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

func (e *Entity) getPkValue(v interface{}) interface{} {

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

func (e *Entity) setPkValue(v interface{}, value interface{}) {
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

func (e *Entity) toMap(obj interface{}) []keyValue {
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
	dialect *Dialect
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

func objectMetadataFrom(v IsEntity, dialect *Dialect) *objectMetadata {
	return &objectMetadata{
		Table:   initTableName(v),
		dialect: dialect,
		Type:    reflect.TypeOf(v),
		Fields:  fieldsOf(v, dialect),
	}
}
