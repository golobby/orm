package orm

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"

	"github.com/gertd/go-pluralize"
	"github.com/golobby/orm/querybuilder"

	//Drivers
	_ "github.com/mattn/go-sqlite3"
)

type Connection struct {
	Name       string
	Dialect    *querybuilder.Dialect
	Connection *sql.DB
	Schemas    map[string]*schema
}

func (d *Connection) getSchema(t string) *schema {
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
	configurator := &EntityConfigurator{}
	e.ConfigureEntity(configurator)

	if configurator.table == "" {
		panic("Table name is mandatory for entities")
	}
	return configurator.table
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
	schemas := map[string]*schema{}
	for _, entity := range entities {
		md := schemaOf(entity)
		if md.dialect == nil {
			md.dialect = dialect
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
	ConfigureEntity(e *EntityConfigurator)
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
		return nil, fmt.Errorf("err no dialect matched with driver")
	}
}

// Insert given Entity
func Insert(obj Entity) error {
	cols := getSchemaFor(obj).Columns(false)
	values := genericValuesOf(obj, false)
	var phs []string
	if getSchemaFor(obj).getDialect().PlaceholderChar == "$" {
		phs = PlaceHolderGenerators.Postgres(len(cols))
	} else {
		phs = PlaceHolderGenerators.MySQL(len(cols))
	}
	qb := &querybuilder.Insert{}
	q, args := qb.
		Table(getSchemaFor(obj).getTable()).
		Into(cols...).
		Values(phs...).
		WithArgs(values...).Build()

	res, err := getSchemaFor(obj).getSQLDB().Exec(q, args...)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	genericSetPkValue(obj, id)
	return nil
}

//func InsertAll(objs ...Entity) error {
//	obj := objs[0]
//	cols := objs[0].schema().Get().Columns(false)
//	qb := &querybuilder.Insert{}
//	qb = qb.
//		Table(obj.schema().Get().getTable()).
//		Into(cols...)
//	for _, obj := range objs {
//		var ph []string
//		if obj.schema().Get().getDialect().PlaceholderChar == "$" {
//			ph = PlaceHolderGenerators.Postgres(len(cols))
//		} else {
//			ph = PlaceHolderGenerators.MySQL(len(cols))
//		}
//		qb.Values(ph...)
//		qb.WithArgs(obj.schema().Get().Values(obj, false))
//	}
//
//	res, err := obj.schema().Get().getSQLDB().Exec(q, args...)
//	if err != nil {
//		return err
//	}
//	id, err := res.LastInsertId()
//	if err != nil {
//		return err
//	}
//	obj.schema().Get().SetPK(obj, id)
//	return nil
//}

// Save upserts given entity.
func Save(obj Entity) error {
	if reflect.ValueOf(genericGetPKValue(obj)).IsZero() {
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
	md := getSchemaFor(*out)
	var args []interface{}
	ph := md.dialect.PlaceholderChar
	if md.dialect.IncludeIndexInPlaceholder {
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
	vs := genericValuesOf(obj, withPK)
	cols := getSchemaFor(obj).Columns(withPK)
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
	ph := getSchemaFor(obj).getDialect().PlaceholderChar
	if getSchemaFor(obj).getDialect().IncludeIndexInPlaceholder {
		ph = ph + "1"
	}
	counter := 2
	kvs := toMap(obj, false)
	var kvsWithPh []keyValue
	var args []interface{}
	whereClause := querybuilder.WhereHelpers.Equal(getSchemaFor(obj).pkName(), ph)
	query := querybuilder.UpdateStmt().
		Table(getSchemaFor(obj).getTable()).
		Where(whereClause)
	for _, kv := range kvs {
		thisPh := getSchemaFor(obj).getDialect().PlaceholderChar
		if getSchemaFor(obj).getDialect().IncludeIndexInPlaceholder {
			thisPh += fmt.Sprint(counter)
		}
		kvsWithPh = append(kvsWithPh, keyValue{Key: kv.Key, Value: thisPh})
		query.Set(kv.Key, thisPh)
		query.WithArgs(kv.Value)
		counter++
	}
	query.WithArgs(genericGetPKValue(obj))
	q, args := query.Build()
	_, err := getSchemaFor(obj).getSQLDB().Exec(q, args...)
	return err
}

// Delete given Entity from database
func Delete(obj Entity) error {
	ph := getSchemaFor(obj).getDialect().PlaceholderChar
	if getSchemaFor(obj).getDialect().IncludeIndexInPlaceholder {
		ph = ph + "1"
	}
	query := querybuilder.WhereHelpers.Equal(getSchemaFor(obj).pkName(), ph)
	qb := &querybuilder.Delete{}
	q, args := qb.
		Table(getSchemaFor(obj).getTable()).
		Where(query).
		WithArgs(genericGetPKValue(obj)).
		Build()
	_, err := getSchemaFor(obj).getSQLDB().Exec(q, args...)
	return err
}

func bindContext[T Entity](ctx context.Context, output interface{}, q string, args []interface{}) error {
	outputMD := getSchemaFor(*new(T))
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
		c.PropertyForeignKey = pluralize.NewClient().Singular(getSchemaFor(owner).getTable()) + "_id"
	}

	ph := getSchemaFor(owner).getDialect().PlaceholderChar
	if getSchemaFor(owner).getDialect().IncludeIndexInPlaceholder {
		ph = ph + fmt.Sprint(1)
	}
	var q string
	var args []interface{}
	qb := &querybuilder.Select{}
	q, args = qb.
		From(c.PropertyTable).
		Where(querybuilder.WhereHelpers.Equal(c.PropertyForeignKey, ph)).
		WithArgs(genericGetPKValue(owner)).
		Build()

	if q == "" {
		return nil, fmt.Errorf("cannot build the query")
	}

	err := bindContext[OUT](context.Background(), &out, q, args)

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
	property := getSchemaFor(*new(PROPERTY))
	//settings default config Values
	if c.PropertyTable == "" {
		c.PropertyTable = property.Table
	}
	if c.PropertyForeignKey == "" {
		c.PropertyForeignKey = pluralize.NewClient().Singular(getSchemaFor(owner).Table) + "_id"
	}

	ph := property.dialect.PlaceholderChar
	if property.dialect.IncludeIndexInPlaceholder {
		ph = ph + fmt.Sprint(1)
	}
	var q string
	var args []interface{}
	qb := &querybuilder.Select{}
	q, args = qb.
		From(c.PropertyTable).
		Where(querybuilder.WhereHelpers.Equal(c.PropertyForeignKey, ph)).
		WithArgs(genericGetPKValue(owner)).
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
	owner := getSchemaFor(*new(OWNER))
	if c.OwnerTable == "" {
		c.OwnerTable = owner.Table
	}
	if c.LocalForeignKey == "" {
		c.LocalForeignKey = pluralize.NewClient().Singular(owner.Table) + "_id"
	}
	if c.ForeignColumnName == "" {
		c.ForeignColumnName = "id"
	}

	ph := owner.getDialect().PlaceholderChar
	if owner.getDialect().IncludeIndexInPlaceholder {
		ph = ph + fmt.Sprint(1)
	}
	ownerIDidx := 0
	for idx, field := range owner.fields {
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
	IntermediateTable      string
	IntermediatePropertyID string
	IntermediateOwnerID    string
	ForeignTable           string
	ForeignLookupColumn    string
}

//BelongsToMany
func BelongsToMany[OWNER Entity](property Entity, c BelongsToManyConfig) ([]OWNER, error) {
	out := new(OWNER)

	if c.ForeignLookupColumn == "" {
		c.ForeignLookupColumn = getSchemaFor(*new(OWNER)).pkName()
	}
	if c.ForeignTable == "" {
		c.ForeignTable = getSchemaFor(*new(OWNER)).Table
	}
	if c.IntermediateTable == "" {
		return nil, fmt.Errorf("cannot infer intermediate table yet")
	}
	if c.IntermediatePropertyID == "" {
		c.IntermediatePropertyID = pluralize.NewClient().Singular(getSchemaFor(property).Table) + "_id"
	}
	if c.IntermediateOwnerID == "" {
		c.IntermediateOwnerID = pluralize.NewClient().Singular(getSchemaFor(*out).Table) + "_id"
	}

	q := fmt.Sprintf(`select %s from %s where %s IN (select %s from %s where %s = ?)`,
		strings.Join(getSchemaFor(*out).Columns(true), ","),
		getSchemaFor(*out).Table,
		c.ForeignLookupColumn,
		c.IntermediateOwnerID,
		c.IntermediateTable,
		c.IntermediatePropertyID,
	)

	args := []interface{}{genericGetPKValue(property)}

	rows, err := getSchemaFor(*out).getSQLDB().Query(q, args...)

	if err != nil {
		return nil, err
	}
	var output []OWNER
	err = getSchemaFor(*out).bind(rows, &output)
	if err != nil {
		return nil, err
	}

	return output, nil
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
	switch relationType {
	}
	// TODO: impl me
	return nil
}

func Query[OUTPUT Entity](stmt *querybuilder.Select) ([]OUTPUT, error) {
	o := new(OUTPUT)
	rows, err := getSchemaFor(*o).getSQLDB().Query(stmt.Build())
	if err != nil {
		return nil, err
	}
	var output []OUTPUT
	err = getSchemaFor(*o).bind(rows, output)
	if err != nil {
		return nil, err
	}
	return output, nil
}

func Exec[E Entity](stmt querybuilder.SQL) (int64, int64, error) {
	e := new(E)

	res, err := getSchemaFor(*e).getSQLDB().Exec(stmt.Build())
	if err != nil {
		return 0, 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, 0, err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return 0, 0, err
	}

	return id, affected, nil
}

func ExecRaw[E Entity](q string, args ...interface{}) (int64, int64, error) {
	e := new(E)

	res, err := getSchemaFor(*e).getSQLDB().Exec(q, args...)
	if err != nil {
		return 0, 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, 0, err
	}

	affected, err := res.RowsAffected()
	if err != nil {
		return 0, 0, err
	}

	return id, affected, nil
}

func QueryRaw[OUTPUT Entity](q string, args ...interface{}) ([]OUTPUT, error) {
	o := new(OUTPUT)
	rows, err := getSchemaFor(*o).getSQLDB().Query(q, args...)
	if err != nil {
		return nil, err
	}
	var output []OUTPUT
	err = getSchemaFor(*o).bind(rows, output)
	if err != nil {
		return nil, err
	}
	return output, nil
}
