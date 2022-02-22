package orm

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/gertd/go-pluralize"
	"github.com/iancoleman/strcase"
	"reflect"
	"strings"
	"unsafe"
)

type DB struct {
	name      string
	dialect   *Dialect
	conn      *sql.DB
	metadatas map[string]*EntityMetadata
}

var globalORM = map[string]*DB{}

type ConnectionConfig struct {
	Name             string
	Driver           string
	ConnectionString string
	DB               *sql.DB
	Dialect          *Dialect
	Entities         []Entity
	EntityMDs        []*EntityMetadata
}

func initTableName(e Entity) string {
	if e.MD().table != "" {
		return e.MD().table
	}
	t := reflect.TypeOf(e)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() == reflect.Slice {
		t = t.Elem()
	}
	name := t.Name()
	if name == "" {
		name = t.String()
	}
	parts := strings.Split(name, ".")
	name = parts[len(parts)-1]
	return strcase.ToSnake(pluralize.NewClient().Plural(name))
}

func Initialize(confs ...ConnectionConfig) error {
	for _, conf := range confs {
		var dialect *Dialect
		var db *sql.DB
		var err error
		if conf.DB != nil && conf.Dialect != nil {
			dialect = conf.Dialect
			db = conf.DB
		} else {
			dialect, err = getDialect(conf.Driver)
			if err != nil {
				return err
			}
			db, err = getDB(conf.Driver, conf.ConnectionString)
			if err != nil {
				return err
			}
		}
		initialize(conf.Name, dialect, db, conf.Entities, conf.EntityMDs)
	}
	return nil
}

func initialize(name string, dialect *Dialect, db *sql.DB, entities []Entity, entityMDs []*EntityMetadata) *DB {
	metadatas := map[string]*EntityMetadata{}
	for idx, entity := range entities {
		md := EntityMetadataFor(entity, dialect)
		metadatas[fmt.Sprintf("%s", initTableName(entity))] = md
		entityMDs[idx] = md
	}
	s := &DB{
		name:      name,
		conn:      db,
		metadatas: metadatas,
		dialect:   dialect,
	}
	globalORM[fmt.Sprintf("%s", name)] = s
	return s
}

type Entity interface {
	MD() *BaseMetadata
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

type HasFields interface {
	Fields() []*Field
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

func fieldsOf(obj interface{}, dialect *Dialect) []*Field {
	hasFields, is := obj.(HasFields)
	if is {
		return hasFields.Fields()
	}
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

type EntityMetadata struct {
	// DriverName of the table that the object represents
	Connection string
	Table      string
	Type       reflect.Type
	dialect    *Dialect
	Fields     []*Field
}

func (o *EntityMetadata) Columns(withPK bool) []string {
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

func (e *EntityMetadata) pkName() string {
	for _, field := range e.Fields {
		if field.IsPK {
			return field.Name
		}
	}
	return ""
}

func (e *EntityMetadata) getDB() *DB {
	if len(globalORM) > 1 && (e.Connection == "" || e.Table == "") {
		panic("need table and connection name when having more than 1 connection registered")
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

func EntityMetadataFor(v Entity, dialect *Dialect) *EntityMetadata {
	return &EntityMetadata{
		Connection: v.MD().connection,
		Table:      initTableName(v),
		dialect:    dialect,
		Type:       reflect.TypeOf(v),
		Fields:     fieldsOf(v, dialect),
	}
}

func getDB(driver string, connectionString string) (*sql.DB, error) {
	return sql.Open(driver, connectionString)
}

func getDialect(driver string) (*Dialect, error) {
	switch driver {
	case "mysql":
		return Dialects.MySQL, nil
	case "sqlite":
		return Dialects.SQLite3, nil
	case "postgres":
		return Dialects.PostgreSQL, nil
	default:
		return nil, fmt.Errorf("err no dialect matched with driver")
	}
}

func getMetadataFor(obj Entity) *EntityMetadata {
	e := obj.MD()
	db := e.getDB()
	return db.metadatas[e.getTable()]
}

func getTableFor(obj Entity) string {
	e := obj.MD()
	return e.table
}

func getConnectionFor(obj Entity) *sql.DB {
	e := obj.MD()
	return e.getDB().conn
}

func getDialectFor(obj Entity) *Dialect {
	e := obj.MD()
	return e.getMetadata().dialect
}

func getDBFor(obj Entity) *DB {
	e := obj.MD()
	if len(globalORM) > 1 && (e.connection == "" || e.getTable() == "") {
		panic("need table and connection name when having more than 1 connection registered")
	}
	if len(globalORM) == 1 {
		for _, db := range globalORM {
			return db
		}
	}
	if db, exists := globalORM[fmt.Sprintf("%s", e.connection)]; exists {
		return db
	}
	panic("no db found")
}

func getValuesOf(o Entity, withPK bool) []interface{} {

	if o.MD().values != nil {
		return o.MD().values(o, withPK)
	}
	t := reflect.TypeOf(o)
	v := reflect.ValueOf(o)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}
	fields := o.MD().getFields()
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

func setPkValueFor(obj Entity, value interface{}) {
	if obj.MD().setPK != nil {
		obj.MD().setPK(obj, value)
		return
	}
	t := reflect.TypeOf(obj)
	val := reflect.ValueOf(obj)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		val = val.Elem()
	}
	pkIdx := -1
	for i, field := range obj.MD().getFields() {
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
	if obj.MD().getPK != nil {
		c := obj.MD()
		return c.getPK(obj)
	}

	t := reflect.TypeOf(obj)
	val := reflect.ValueOf(obj)
	if t.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	fields := obj.MD().getFields()
	for i, field := range fields {
		if field.IsPK {
			return val.Field(i).Interface()
		}
	}
	return ""
}

// Save given Entity
func Save(obj Entity) error {
	cols := obj.MD().getMetadata().Columns(false)
	values := getValuesOf(obj, false)
	var phs []string
	if obj.MD().getDialect().PlaceholderChar == "$" {
		phs = PlaceHolderGenerators.Postgres(len(cols))
	} else {
		phs = PlaceHolderGenerators.MySQL(len(cols))
	}
	q, args := Insert().
		Table(obj.MD().getTable()).
		Into(cols...).
		Values(phs...).
		WithArgs(values...).Build()

	res, err := obj.MD().getConnection().Exec(q, args...)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	setPkValueFor(obj, id)
	return nil
}

func Find[T Entity](md *EntityMetadata, id interface{}) (T, error) {
	var q string
	out := new(T)
	var args []interface{}
	ph := md.dialect.PlaceholderChar
	if md.dialect.IncludeIndexInPlaceholder {
		ph = ph + "1"
	}
	builder := Select().
		Select(md.Columns(true)...).
		From(md.Table).
		Where(WhereHelpers.Equal(md.pkName(), ph)).
		WithArgs(id)

	q, args = builder.
		Build()

	err := BindContext(context.Background(), md, out, q, args)

	if err != nil {
		return *out, err
	}

	return *out, nil
}

// Fill given Entity
func Fill(obj Entity) error {
	var q string
	var args []interface{}
	pkValue := getPkValue(obj)
	ph := obj.MD().getDialect().PlaceholderChar
	if obj.MD().getDialect().IncludeIndexInPlaceholder {
		ph = ph + "1"
	}
	builder := Select().
		Select(obj.MD().getMetadata().Columns(true)...).
		From(obj.MD().getMetadata().Table).
		Where(WhereHelpers.Equal(obj.MD().getMetadata().pkName(), ph)).
		WithArgs(pkValue)

	q, args = builder.
		Build()

	return BindContext(context.Background(), obj.MD().getMetadata(), obj, q, args)
}

func toMap(obj Entity) []keyValue {
	var kvs []keyValue
	vs := getValuesOf(obj, true)
	cols := obj.MD().getMetadata().Columns(true)
	for i, col := range cols {
		kvs = append(kvs, keyValue{
			Key:   col,
			Value: vs[i],
		})
	}
	return kvs
}

// Update Entity in database
func Update(obj Entity) error {
	ph := obj.MD().getDialect().PlaceholderChar
	if obj.MD().getDialect().IncludeIndexInPlaceholder {
		ph = ph + "1"
	}
	counter := 2
	kvs := toMap(obj)
	var kvsWithPh []keyValue
	var args []interface{}
	whereClause := WhereHelpers.Equal(obj.MD().getMetadata().pkName(), ph)
	query := UpdateStmt().
		Table(obj.MD().getTable()).
		Where(whereClause).WithArgs(getPkValue(obj))
	for _, kv := range kvs {
		thisPh := obj.MD().getDialect().PlaceholderChar
		if obj.MD().getDialect().IncludeIndexInPlaceholder {
			thisPh += fmt.Sprint(counter)
		}
		kvsWithPh = append(kvsWithPh, keyValue{Key: kv.Key, Value: thisPh})
		query.Set(kv.Key, thisPh)
		query.WithArgs(kv.Value)
		counter++
	}
	q, args := query.Build()
	_, err := obj.MD().getConnection().Exec(q, args...)
	return err
}

// Delete the object from database
func Delete(obj Entity) error {
	ph := obj.MD().getDialect().PlaceholderChar
	if obj.MD().getDialect().IncludeIndexInPlaceholder {
		ph = ph + "1"
	}
	query := WhereHelpers.Equal(obj.MD().getMetadata().pkName(), ph)
	q, args := DeleteStmt().
		Table(obj.MD().getTable()).
		Where(query).
		WithArgs(getPkValue(obj)).
		Build()
	_, err := obj.MD().getConnection().Exec(q, args...)
	return err
}

func BindContext(ctx context.Context, outputMD *EntityMetadata, output interface{}, q string, args []interface{}) error {
	rows, err := outputMD.getDB().conn.QueryContext(ctx, q, args...)
	if err != nil {
		return err
	}
	return outputMD.Bind(rows, output)
}

type HasManyConfig struct {
	PropertyTable      string
	PropertyForeignKey string
}

func HasMany[OUT any](owner Entity, property *EntityMetadata, c HasManyConfig) ([]OUT, error) {
	var out []OUT
	//settings default config values
	if c.PropertyTable == "" {
		c.PropertyTable = property.Table
	}
	if c.PropertyForeignKey == "" {
		c.PropertyForeignKey = pluralize.NewClient().Singular(owner.MD().getTable()) + "_id"
	}

	ph := owner.MD().getDialect().PlaceholderChar
	if owner.MD().getDialect().IncludeIndexInPlaceholder {
		ph = ph + fmt.Sprint(1)
	}
	var q string
	var args []interface{}

	q, args = Select().
		From(c.PropertyTable).
		Where(WhereHelpers.Equal(c.PropertyForeignKey, ph)).
		WithArgs(getPkValue(owner)).
		Build()

	if q == "" {
		return nil, fmt.Errorf("cannot build the query")
	}

	err := BindContext(context.Background(), property, out, q, args)

	if err != nil {
		return nil, err
	}

	return out, nil
}

type HasOneConfig struct {
	PropertyTable      string
	PropertyForeignKey string
}

func HasOne[PROPERTY any](owner Entity, property *EntityMetadata, c HasOneConfig) (PROPERTY, error) {
	out := new(PROPERTY)
	//settings default config values
	if c.PropertyTable == "" {
		c.PropertyTable = property.Table
	}
	if c.PropertyForeignKey == "" {
		c.PropertyForeignKey = pluralize.NewClient().Singular(property.Table) + "_id"
	}

	ph := property.dialect.PlaceholderChar
	if property.dialect.IncludeIndexInPlaceholder {
		ph = ph + fmt.Sprint(1)
	}
	var q string
	var args []interface{}

	q, args = Select().
		From(c.PropertyTable).
		Where(WhereHelpers.Equal(c.PropertyForeignKey, ph)).
		WithArgs(getPkValue(owner)).
		Build()

	if q == "" {
		return *out, fmt.Errorf("cannot build the query")
	}

	err := BindContext(context.Background(), property, out, q, args)

	return *out, err
}

type BelongsToConfig struct {
	OwnerTable        string
	LocalForeignKey   string
	ForeignColumnName string
}

func BelongsTo[OWNER any](property Entity, owner *EntityMetadata, c BelongsToConfig) (OWNER, error) {
	out := new(OWNER)
	if c.OwnerTable == "" {
		c.OwnerTable = owner.Table
	}
	if c.LocalForeignKey == "" {
		c.LocalForeignKey = pluralize.NewClient().Singular(owner.Table) + "_id"
	}
	if c.ForeignColumnName == "" {
		c.ForeignColumnName = "id"
	}

	ph := owner.dialect.PlaceholderChar
	if owner.dialect.IncludeIndexInPlaceholder {
		ph = ph + "1"
	}
	ownerIDidx := 0
	for idx, field := range owner.Fields {
		if field.Name == c.LocalForeignKey {
			ownerIDidx = idx
		}
	}

	ownerID := getValuesOf(property, true)[ownerIDidx]

	q, args := Select().
		From(c.OwnerTable).
		Where(WhereHelpers.Equal(c.ForeignColumnName, ph)).
		WithArgs(ownerID).Build()

	err := BindContext(context.Background(), owner, out, q, args)
	return *out, err
}

type ManyToManyConfig struct {
	IntermediateTable         string
	IntermediateLocalColumn   string
	IntermediateForeignColumn string
	ForeignTable              string
	ForeignLookupColumn       string
}

func ManyToMany[TARGET any](obj Entity, target *EntityMetadata, c ManyToManyConfig) ([]TARGET, error) {
	// TODO: Impl me
	return nil, nil
}
