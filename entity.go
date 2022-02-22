package orm

import (
	"context"
	"fmt"
	"reflect"
	"unsafe"

	"github.com/gertd/go-pluralize"
)

type Entity struct {
	*BaseEntity
	obj IsEntity
}

func AsEntity(obj IsEntity) *Entity {
	return &Entity{BaseEntity: obj.E(), obj: obj}
}

type IsEntity interface {
	E() *BaseEntity
}

func (e *Entity) Load(rl func(e *Entity) Relation) error {
	r := rl(e)
	switch r.typ {
	case relationTypeHasOne:
		return e.hasOne(r.output.(IsEntity), r.c.(HasOneConfig))
	case relationTypeHasMany:
		return e.hasMany(r.output.([]IsEntity), r.c.(HasManyConfig))
	case relationTypeMany2Many:
		return e.manyToMany(r.output.([]IsEntity), r.c.(ManyToManyConfig))
	case relationTypeBelongsTo:
		return e.belongsTo(r.output.(IsEntity), r.c.(BelongsToConfig))
	default:
		return fmt.Errorf("type of relation not matched")
	}
}

func (e *Entity) LoadAll(rls ...func(e *Entity) Relation) error {
	for _, rl := range rls {
		if err := e.Load(rl); err != nil {
			return err
		}
	}
	return nil
}

func (e *Entity) getValues(o interface{}, withPK bool) []interface{} {
	if e.obj.E().values != nil {
		return e.obj.E().values(o, withPK)
	}
	t := reflect.TypeOf(o)
	v := reflect.ValueOf(o)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}
	fields := e.obj.E().getFields()
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
	pkValue := e.getPkValue(e.obj)
	ph := e.obj.E().getDialect().PlaceholderChar
	if e.obj.E().getDialect().IncludeIndexInPlaceholder {
		ph = ph + "1"
	}
	builder := Select().
		Select(e.obj.E().getMetadata().Columns(true)...).
		From(e.obj.E().getMetadata().Table).
		Where(WhereHelpers.Equal(e.obj.E().getMetadata().pkName(), ph)).
		WithArgs(pkValue)

	q, args = builder.
		Build()

	eq := &ExecutableQuery{
		q:      q,
		args:   args,
		bindTo: e.obj,
	}

	return eq.Exec()
}

// Save given object
func (e *Entity) Save() error {
	cols := e.obj.E().getMetadata().Columns(false)
	values := e.getValues(e.obj, false)
	var phs []string
	if e.obj.E().getDialect().PlaceholderChar == "$" {
		phs = PlaceHolderGenerators.Postgres(len(cols))
	} else {
		phs = PlaceHolderGenerators.MySQL(len(cols))
	}
	q, args := Insert().
		Table(e.obj.E().getTable()).
		Into(cols...).
		Values(phs...).
		WithArgs(values...).Build()

	res, err := e.obj.E().getConnection().Exec(q, args...)
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
	ph := e.obj.E().getDialect().PlaceholderChar
	if e.obj.E().getDialect().IncludeIndexInPlaceholder {
		ph = ph + "1"
	}
	counter := 2
	kvs := e.toMap(e.obj)
	var kvsWithPh []keyValue
	var args []interface{}
	whereClause := WhereHelpers.Equal(e.obj.E().getMetadata().pkName(), ph)
	query := Update().
		Table(e.obj.E().getTable()).
		Where(whereClause).WithArgs(e.getPkValue(e.obj))
	for _, kv := range kvs {
		thisPh := e.obj.E().getDialect().PlaceholderChar
		if e.obj.E().getDialect().IncludeIndexInPlaceholder {
			thisPh += fmt.Sprint(counter)
		}
		kvsWithPh = append(kvsWithPh, keyValue{Key: kv.Key, Value: thisPh})
		query.Set(kv.Key, thisPh)
		query.WithArgs(kv.Value)
		counter++
	}
	q, args := query.Build()
	_, err := e.obj.E().getConnection().Exec(q, args...)
	return err
}

// Delete the object from database
func (e *Entity) Delete() error {
	ph := e.obj.E().getDialect().PlaceholderChar
	if e.obj.E().getDialect().IncludeIndexInPlaceholder {
		ph = ph + "1"
	}
	query := WhereHelpers.Equal(e.obj.E().getMetadata().pkName(), ph)
	q, args := Delete().
		Table(e.obj.E().getTable()).
		Where(query).
		WithArgs(e.getPkValue(e.obj)).
		Build()
	_, err := e.obj.E().getConnection().Exec(q, args...)
	return err
}

func (e *Entity) BindContext(ctx context.Context, out interface{}, q string, args ...interface{}) error {
	rows, err := e.obj.E().getConnection().QueryContext(ctx, q, args...)
	if err != nil {
		return err
	}
	return e.obj.E().getMetadata().Bind(rows, out)
}

func (e *Entity) Query() *SelectStmt {
	return &SelectStmt{
		table:           e.obj.E().getTable(),
		referenceEntity: e.obj,
	}
}

const (
	relationTypeHasOne    = "hasone"
	relationTypeHasMany   = "hasmany"
	relationTypeMany2Many = "many2many"
	relationTypeBelongsTo = "belongsto"
)

type Relation struct {
	typ    string
	output interface{}
	c      interface{}
}
type HasManyConfig struct {
	PropertyTable      string
	PropertyForeignKey string
}

func (e *Entity) hasMany(out []IsEntity, c HasManyConfig) error {
	outEntity := &Entity{obj: reflect.New(reflect.TypeOf(out).Elem()).Interface().(IsEntity)}
	//settings default config values
	if c.PropertyTable == "" {
		c.PropertyTable = outEntity.obj.E().getTable()
	}
	if c.PropertyForeignKey == "" {
		c.PropertyForeignKey = pluralize.NewClient().Singular(e.obj.E().getTable()) + "_id"
	}

	ph := e.obj.E().getDialect().PlaceholderChar
	if e.obj.E().getDialect().IncludeIndexInPlaceholder {
		ph = ph + fmt.Sprint(1)
	}
	var q string
	var args []interface{}

	q, args = Select().
		From(c.PropertyTable).
		Where(WhereHelpers.Equal(c.PropertyForeignKey, ph)).
		WithArgs(e.getPkValue(e.obj)).
		Build()

	if q == "" {
		return fmt.Errorf("cannot build the query")
	}

	eq := &ExecutableQuery{
		q:      q,
		args:   args,
		bindTo: out,
	}

	return eq.Exec()

}

type HasOneConfig struct {
	PropertyTable      string
	PropertyForeignKey string
}

func (e *Entity) hasOne(out IsEntity, c HasOneConfig) error {
	outEntity := &Entity{obj: out}
	//settings default config values
	if c.PropertyTable == "" {
		c.PropertyTable = outEntity.obj.E().getTable()
	}
	if c.PropertyForeignKey == "" {
		c.PropertyForeignKey = pluralize.NewClient().Singular(e.obj.E().getTable()) + "_id"
	}

	ph := e.obj.E().getDialect().PlaceholderChar
	if e.obj.E().getDialect().IncludeIndexInPlaceholder {
		ph = ph + fmt.Sprint(1)
	}
	var q string
	var args []interface{}

	q, args = Select().
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

func (e *Entity) belongsTo(out IsEntity, c BelongsToConfig) error {
	outEntity := &Entity{obj: out}
	if c.OwnerTable == "" {
		c.OwnerTable = outEntity.obj.E().getTable()
	}
	if c.LocalForeignKey == "" {
		c.LocalForeignKey = pluralize.NewClient().Singular(outEntity.obj.E().getTable()) + "_id"
	}
	if c.ForeignColumnName == "" {
		c.ForeignColumnName = "id"
	}

	ph := e.obj.E().getDialect().PlaceholderChar
	if e.obj.E().getDialect().IncludeIndexInPlaceholder {
		ph = ph + "1"
	}
	ownerIDidx := 0
	for idx, field := range e.obj.E().getFields() {
		if field.Name == c.LocalForeignKey {
			ownerIDidx = idx
		}
	}

	ownerID := e.getValues(e.obj, true)[ownerIDidx]

	q, args := Select().
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

func (e *Entity) manyToMany(out []IsEntity, c ManyToManyConfig) error {
	outEntity := &Entity{obj: reflect.New(reflect.TypeOf(out).Elem()).Interface().(IsEntity)}
	if c.IntermediateTable == "" {
		return fmt.Errorf("no way to infer many to many intermediate table yet.")
	}
	if c.IntermediateLocalColumn == "" {
		table := e.obj.E().getTable()
		table = pluralize.NewClient().Singular(table)
		c.IntermediateLocalColumn = table + "_id"
	}
	if c.IntermediateForeignColumn == "" {
		table := outEntity.obj.E().getTable()
		c.IntermediateForeignColumn = pluralize.NewClient().Singular(table) + "_id"
	}
	if c.ForeignTable == "" {
		c.IntermediateForeignColumn = outEntity.obj.E().getTable()
	}
	//TODO: this logic is wrong
	sub, _ := Select().From(c.IntermediateTable).Where(c.IntermediateLocalColumn, "=", fmt.Sprint(e.getPkValue(e.obj))).Build()
	q, args := Select().
		From(c.ForeignTable).
		Where(c.ForeignLookupColumn, "in", sub).
		Build()

	return e.
		BindContext(context.Background(), out, q, args...)

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

	if e.obj.E().getPK != nil {
		c := e.obj.E()
		return c.getPK(e.obj)
	}

	t := reflect.TypeOf(v)
	val := reflect.ValueOf(v)
	if t.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	fields := e.obj.E().getFields()
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
	if e.obj.E().setPK != nil {
		e.obj.E().setPK(v, value)
		return
	}
	t := reflect.TypeOf(v)
	val := reflect.ValueOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		val = val.Elem()
	}
	pkIdx := -1
	for i, field := range e.obj.E().getFields() {
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
	cols := e.obj.E().getMetadata().Columns(true)
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
