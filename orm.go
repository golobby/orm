package orm

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/gertd/go-pluralize"
	"github.com/golobby/orm/querybuilder"
	"reflect"

	//Drivers
	_ "github.com/mattn/go-sqlite3"
)

type Connection struct {
	Name       string
	Dialect    *querybuilder.Dialect
	Connection *sql.DB
	Schemas    map[string]*Schema
}

func (d *Connection) getSchema(t string) *Schema {
	return d.Schemas[t]
}

var globalORM = map[string]*Connection{}

func GetConnection(name string) *Connection {
	return globalORM[name]
}

type ConnectionConfig struct {
	Name             string
	Driver           string
	ConnectionString string
	DB               *sql.DB
	Dialect          *querybuilder.Dialect
	Entities         []Entity
}

func initTableName(e Entity) string {
	if e.Schema().Table == "" {
		panic("Table name is mandatory for entities")
	}
	return e.Schema().Table
}

func Initialize(confs ...ConnectionConfig) error {
	for _, conf := range confs {
		var dialect *querybuilder.Dialect
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
		initialize(conf.Name, dialect, db, conf.Entities)
	}
	return nil
}

func initialize(name string, dialect *querybuilder.Dialect, db *sql.DB, entities []Entity) *Connection {
	schemas := map[string]*Schema{}
	for _, entity := range entities {
		md := schemaOf(entity)
		if md.Dialect == nil {
			md.Dialect = dialect
		}
		schemas[fmt.Sprintf("%s", initTableName(entity))] = md
	}
	s := &Connection{
		Name:       name,
		Connection: db,
		Schemas:    schemas,
		Dialect:    dialect,
	}
	globalORM[fmt.Sprintf("%s", name)] = s
	return s
}

type Entity interface {
	Schema() *Schema
}

func getDB(driver string, connectionString string) (*sql.DB, error) {
	return sql.Open(driver, connectionString)
}

func getDialect(driver string) (*querybuilder.Dialect, error) {
	switch driver {
	case "mysql":
		return querybuilder.Dialects.MySQL, nil
	case "sqlite", "sqlite3":
		return querybuilder.Dialects.SQLite3, nil
	case "postgres":
		return querybuilder.Dialects.PostgreSQL, nil
	default:
		return nil, fmt.Errorf("err no Dialect matched with driver")
	}
}

// Insert given Entity
func Insert(obj Entity) error {
	cols := obj.Schema().Get().Columns(false)
	values := obj.Schema().Get().Values(obj, false)
	var phs []string
	if obj.Schema().Get().getDialect().PlaceholderChar == "$" {
		phs = PlaceHolderGenerators.Postgres(len(cols))
	} else {
		phs = PlaceHolderGenerators.MySQL(len(cols))
	}
	qb := &querybuilder.Insert{}
	q, args := qb.
		Table(obj.Schema().Get().getTable()).
		Into(cols...).
		Values(phs...).
		WithArgs(values...).Build()

	res, err := obj.Schema().Get().getSQLDB().Exec(q, args...)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	obj.Schema().Get().SetPK(obj, id)
	return nil
}
func InsertAll(objs ...Entity) error {
	obj := objs[0]
	cols := objs[0].Schema().Get().Columns(false)
	qb := &querybuilder.Insert{}
	qb = qb.
		Table(obj.Schema().Get().getTable()).
		Into(cols...)
	for _, obj := range objs {
		var ph []string
		if obj.Schema().Get().getDialect().PlaceholderChar == "$" {
			ph = PlaceHolderGenerators.Postgres(len(cols))
		} else {
			ph = PlaceHolderGenerators.MySQL(len(cols))
		}
		qb.Values(ph...)
		qb.WithArgs(obj.Schema().Get().Values(obj, false))
	}

	res, err := obj.Schema().Get().getSQLDB().Exec(q, args...)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	obj.Schema().Get().SetPK(obj, id)
	return nil
}

// Save upserts given entity.
func Save(obj Entity) error {
	if reflect.ValueOf(obj.Schema().Get().GetPK(obj)).IsZero() {
		return Insert(obj)
	} else {
		return Update(obj)
	}
}

// SaveAll saves all given entities in one query, they all should be same type of entity ( same table ).
func SaveAll(objs ...Entity) error {
	// TODO
	return nil
}

// Find finds the Entity you want based on Entity generic type and primary key you passed.
func Find[T Entity](id interface{}) (T, error) {
	var q string
	out := new(T)
	md := (*out).Schema().Get()
	var args []interface{}
	ph := md.Dialect.PlaceholderChar
	if md.Dialect.IncludeIndexInPlaceholder {
		ph = ph + "1"
	}
	qb := &querybuilder.Select{}
	builder := qb.
		Select(md.Columns(true)...).
		From(md.Table).
		Where(querybuilder.WhereHelpers.Equal(md.pkName(), ph)).
		WithArgs(id)

	q, args = builder.
		Build()

	err := bindContext[T](context.Background(), out, q, args)

	if err != nil {
		return *out, err
	}

	return *out, nil
}

func toMap(obj Entity, withPK bool) []keyValue {
	var kvs []keyValue
	vs := obj.Schema().Get().Values(obj, withPK)
	cols := obj.Schema().Get().Columns(withPK)
	for i, col := range cols {
		kvs = append(kvs, keyValue{
			Key:   col,
			Value: vs[i],
		})
	}
	return kvs
}

// Update given Entity in database
func Update(obj Entity) error {
	ph := obj.Schema().Get().getDialect().PlaceholderChar
	if obj.Schema().Get().getDialect().IncludeIndexInPlaceholder {
		ph = ph + "1"
	}
	counter := 2
	kvs := toMap(obj, false)
	var kvsWithPh []keyValue
	var args []interface{}
	whereClause := querybuilder.WhereHelpers.Equal(obj.Schema().Get().pkName(), ph)
	query := querybuilder.UpdateStmt().
		Table(obj.Schema().Get().getTable()).
		Where(whereClause)
	for _, kv := range kvs {
		thisPh := obj.Schema().Get().getDialect().PlaceholderChar
		if obj.Schema().Get().getDialect().IncludeIndexInPlaceholder {
			thisPh += fmt.Sprint(counter)
		}
		kvsWithPh = append(kvsWithPh, keyValue{Key: kv.Key, Value: thisPh})
		query.Set(kv.Key, thisPh)
		query.WithArgs(kv.Value)
		counter++
	}
	query.WithArgs(obj.Schema().Get().GetPK(obj))
	q, args := query.Build()
	_, err := obj.Schema().Get().getSQLDB().Exec(q, args...)
	return err
}

// Delete given Entity from database
func Delete(obj Entity) error {
	ph := obj.Schema().Get().getDialect().PlaceholderChar
	if obj.Schema().Get().getDialect().IncludeIndexInPlaceholder {
		ph = ph + "1"
	}
	query := querybuilder.WhereHelpers.Equal(obj.Schema().Get().pkName(), ph)
	qb := &querybuilder.Delete{}
	q, args := qb.
		Table(obj.Schema().Get().getTable()).
		Where(query).
		WithArgs(obj.Schema().Get().GetPK(obj)).
		Build()
	_, err := obj.Schema().Get().getSQLDB().Exec(q, args...)
	return err
}

func bindContext[T Entity](ctx context.Context, output interface{}, q string, args []interface{}) error {
	outputMD := GetSchema[T]()
	rows, err := outputMD.getConnection().Connection.QueryContext(ctx, q, args...)
	if err != nil {
		return err
	}
	return outputMD.bind(rows, output)
}

type HasManyConfig struct {
	PropertyTable      string
	PropertyForeignKey string
}

func HasMany[OUT Entity](owner Entity, c HasManyConfig) ([]OUT, error) {
	property := schemaOf(*(new(OUT)))
	var out []OUT
	//settings default config Values
	if c.PropertyTable == "" {
		c.PropertyTable = property.Table
	}
	if c.PropertyForeignKey == "" {
		c.PropertyForeignKey = pluralize.NewClient().Singular(owner.Schema().getTable()) + "_id"
	}

	ph := owner.Schema().getDialect().PlaceholderChar
	if owner.Schema().getDialect().IncludeIndexInPlaceholder {
		ph = ph + fmt.Sprint(1)
	}
	var q string
	var args []interface{}
	qb := &querybuilder.Select{}
	q, args = qb.
		From(c.PropertyTable).
		Where(querybuilder.WhereHelpers.Equal(c.PropertyForeignKey, ph)).
		WithArgs(owner.Schema().Get().GetPK(owner)).
		Build()

	if q == "" {
		return nil, fmt.Errorf("cannot build the query")
	}

	err := bindContext[OUT](context.Background(), out, q, args)

	if err != nil {
		return nil, err
	}

	return out, nil
}

type HasOneConfig struct {
	PropertyTable      string
	PropertyForeignKey string
}

func HasOne[PROPERTY Entity](owner Entity, c HasOneConfig) (PROPERTY, error) {
	out := new(PROPERTY)
	property := GetSchema[PROPERTY]()
	//settings default config Values
	if c.PropertyTable == "" {
		c.PropertyTable = property.Table
	}
	if c.PropertyForeignKey == "" {
		c.PropertyForeignKey = pluralize.NewClient().Singular(property.Table) + "_id"
	}

	ph := property.Dialect.PlaceholderChar
	if property.Dialect.IncludeIndexInPlaceholder {
		ph = ph + fmt.Sprint(1)
	}
	var q string
	var args []interface{}
	qb := &querybuilder.Select{}
	q, args = qb.
		From(c.PropertyTable).
		Where(querybuilder.WhereHelpers.Equal(c.PropertyForeignKey, ph)).
		WithArgs(owner.Schema().Get().GetPK(owner)).
		Build()

	if q == "" {
		return *out, fmt.Errorf("cannot build the query")
	}

	err := bindContext[PROPERTY](context.Background(), out, q, args)

	return *out, err
}

type BelongsToConfig struct {
	OwnerTable        string
	LocalForeignKey   string
	ForeignColumnName string
}

func BelongsTo[OWNER Entity](property Entity, c BelongsToConfig) (OWNER, error) {
	out := new(OWNER)
	owner := GetSchema[OWNER]()
	if c.OwnerTable == "" {
		c.OwnerTable = owner.Table
	}
	if c.LocalForeignKey == "" {
		c.LocalForeignKey = pluralize.NewClient().Singular(owner.Table) + "_id"
	}
	if c.ForeignColumnName == "" {
		c.ForeignColumnName = "id"
	}

	ph := owner.Dialect.PlaceholderChar
	if owner.Dialect.IncludeIndexInPlaceholder {
		ph = ph + "1"
	}
	ownerIDidx := 0
	for idx, field := range owner.Fields {
		if field.Name == c.LocalForeignKey {
			ownerIDidx = idx
		}
	}

	ownerID := genericValuesOf(property, true)[ownerIDidx]
	qb := &querybuilder.Select{}
	q, args := qb.
		From(c.OwnerTable).
		Where(querybuilder.WhereHelpers.Equal(c.ForeignColumnName, ph)).
		WithArgs(ownerID).Build()

	err := bindContext[OWNER](context.Background(), out, q, args)
	return *out, err
}

type BelongsToManyConfig struct {
	IntermediateTable         string
	IntermediateLocalColumn   string
	IntermediateForeignColumn string
	ForeignTable              string
	ForeignLookupColumn       string
}

func BelongsToMany[TARGET any](obj Entity, c BelongsToManyConfig) ([]TARGET, error) {
	// TODO: Impl me
	return nil, nil
}

type RelationType int

const (
	RelationType_HasMany RelationType = iota + 1
	RelationType_HasOne
	RelationType_ManyToMany
	RelationType_BelongsTo
)

// Add is a relation function, inserts `items` into database and also creates necessary wiring of relationships based on `relationType`.
// RelationType is from perspective of `to`, so for post and comment example if you want to add comment to a post relationtype is hasMany.
func Add[T Entity](to Entity, relationType RelationType, items ...T) error {
	//TODO
	return nil
}

func Query[OUTPUT Entity](stmt *querybuilder.Select) ([]OUTPUT, error) {
	//TODO
	return nil, nil
}

func Exec(stmt querybuilder.SQL) (int, int, error) {
	//TODO
	return 0, 0, nil
}

func ExecRaw(q string, args ...interface{}) (int, int, error) {
	//TODO
	return 0, 0, nil
}

func RawQuery[OUTPUT Entity](q string, args ...interface{}) ([]OUTPUT, error) {
	//TODO
	return nil, nil
}
