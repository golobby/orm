package orm

import (
	"database/sql"
	"fmt"
	"reflect"
	"time"

	// Drivers
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

var globalConnections = map[string]*connection{}

// Schematic prints all information ORM inferred from your entities in startup, remember to pass
// your entities in Entities when you call SetupConnections if you want their data inferred
// otherwise Schematic does not print correct data since GoLobby ORM also
// incrementally cache your entities metadata and schema.
func Schematic() {
	for conn, connObj := range globalConnections {
		fmt.Printf("----------------%s---------------\n", conn)
		connObj.Schematic()
		fmt.Println("-----------------------------------")
	}
}

type ConnectionConfig struct {
	// Name of your database connection, it's up to you to name them anything
	// just remember that having a connection name is mandatory if
	// you have multiple connections
	Name string
	// If you already have an active database connection configured pass it in this value and
	// do not pass Driver and DSN fields.
	DB *sql.DB
	// Which dialect of sql to generate queries for, you don't need it most of the times when you are using
	// traditional databases such as mysql, sqlite3, postgres.
	Dialect *Dialect
	// List of entities that you want to use for this connection, remember that you can ignore this field
	// and GoLobby ORM will build our metadata cache incrementally but you will lose schematic
	// information that we can provide you and also potentialy validations that we
	// can do with the database
	Entities []Entity
	// Database validations, check if all tables exists and also table schemas contains all necessary columns.
	// Check if all infered tables exist in your database
	DatabaseValidations bool
}

// SetupConnections declares a new connections for ORM.
func SetupConnections(configs ...ConnectionConfig) error {

	for _, c := range configs {
		if err := setupConnection(c); err != nil {
			return err
		}
	}
	for _, conn := range globalConnections {
		if !conn.DatabaseValidations {
			continue
		}

		tables, err := getListOfTables(conn.Dialect.QueryListTables)(conn.DB)
		if err != nil {
			return err
		}

		for _, table := range tables {
			if conn.DatabaseValidations {
				spec, err := getTableSchema(conn.Dialect.QueryTableSchema)(conn.DB, table)
				if err != nil {
					return err
				}
				conn.DBSchema[table] = spec
			} else {
				conn.DBSchema[table] = nil
			}
		}

		// check tables existence
		if conn.DatabaseValidations {
			err := conn.validateAllTablesArePresent()
			if err != nil {
				return err
			}
		}

		if conn.DatabaseValidations {
			err = conn.validateTablesSchemas()
			if err != nil {
				return err
			}
		}

	}

	return nil
}

func setupConnection(config ConnectionConfig) error {
	schemas := map[string]*schema{}
	if config.Name == "" {
		config.Name = "default"
	}

	for _, entity := range config.Entities {
		s := schemaOfHeavyReflectionStuff(entity)
		var configurator EntityConfigurator
		entity.ConfigureEntity(&configurator)
		schemas[configurator.table] = s
	}

	s := &connection{
		Name:                config.Name,
		DB:                  config.DB,
		Dialect:             config.Dialect,
		Schemas:             schemas,
		DBSchema:            make(map[string][]columnSpec),
		DatabaseValidations: config.DatabaseValidations,
	}

	globalConnections[fmt.Sprintf("%s", config.Name)] = s

	return nil
}

// Entity defines the interface that each of your structs that
// you want to use as database entities should have,
// it's a simple one and its ConfigureEntity.
type Entity interface {
	// ConfigureEntity should be defined for all of your database entities
	// and it can define Table, DB and also relations of your Entity.
	ConfigureEntity(e *EntityConfigurator)
}

// InsertAll given entities into database based on their ConfigureEntity
// we can find table and also DB name.
func InsertAll(objs ...Entity) error {
	if len(objs) == 0 {
		return nil
	}
	s := getSchemaFor(objs[0])
	cols := s.Columns(false)
	var values [][]interface{}
	for _, obj := range objs {
		createdAtF := s.createdAt()
		if createdAtF != nil {
			genericSet(obj, createdAtF.Name, sql.NullTime{Time: time.Now(), Valid: true})
		}
		updatedAtF := s.updatedAt()
		if updatedAtF != nil {
			genericSet(obj, updatedAtF.Name, sql.NullTime{Time: time.Now(), Valid: true})
		}
		values = append(values, genericValuesOf(obj, false))
	}

	is := insertStmt{
		PlaceHolderGenerator: s.getDialect().PlaceHolderGenerator,
		Table:                s.getTable(),
		Columns:              cols,
		Values:               values,
	}

	q, args := is.ToSql()

	_, err := s.getConnection().exec(q, args...)
	if err != nil {
		return err
	}
	return nil
}

// Insert given entity into database based on their ConfigureEntity
// we can find table and also DB name.
func Insert(o Entity) error {
	s := getSchemaFor(o)
	cols := s.Columns(false)
	var values [][]interface{}
	createdAtF := s.createdAt()
	if createdAtF != nil {
		genericSet(o, createdAtF.Name, sql.NullTime{Time: time.Now(), Valid: true})
	}
	updatedAtF := s.updatedAt()
	if updatedAtF != nil {
		genericSet(o, updatedAtF.Name, sql.NullTime{Time: time.Now(), Valid: true})
	}
	values = append(values, genericValuesOf(o, false))

	is := insertStmt{
		PlaceHolderGenerator: s.getDialect().PlaceHolderGenerator,
		Table:                s.getTable(),
		Columns:              cols,
		Values:               values,
	}

	if s.getDialect().DriverName == "postgres" {
		is.Returning = s.pkName()
	}
	q, args := is.ToSql()

	res, err := s.getConnection().exec(q, args...)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}

	if s.pkName() != "" {
		// intermediate tables usually have no single pk column.
		s.setPK(o, id)
	}
	return nil
}

func isZero(val interface{}) bool {
	switch val.(type) {
	case int64:
		return val.(int64) == 0
	case int:
		return val.(int) == 0
	case string:
		return val.(string) == ""
	default:
		return reflect.ValueOf(val).Elem().IsZero()
	}
}

// Save saves given entity, if primary key is set
// we will make an update query and if
// primary key is zero value we will
// insert it.
func Save(obj Entity) error {
	if isZero(getSchemaFor(obj).getPK(obj)) {
		return Insert(obj)
	} else {
		return Update(obj)
	}
}

// Find finds the Entity you want based on generic type and primary key you passed.
func Find[T Entity](id interface{}) (T, error) {
	var q string
	out := new(T)
	md := getSchemaFor(*out)
	q, args, err := NewQueryBuilder[T](md).
		SetDialect(md.getDialect()).
		Table(md.Table).
		Select(md.Columns(true)...).
		Where(md.pkName(), id).
		ToSql()
	if err != nil {
		return *out, err
	}
	err = bind[T](out, q, args)

	if err != nil {
		return *out, err
	}

	return *out, nil
}

func toKeyValues(obj Entity, withPK bool) []any {
	var tuples []any
	vs := genericValuesOf(obj, withPK)
	cols := getSchemaFor(obj).Columns(withPK)
	for i, col := range cols {
		tuples = append(tuples, col, vs[i])
	}
	return tuples
}

// Update given Entity in database.
func Update(obj Entity) error {
	s := getSchemaFor(obj)
	q, args, err := NewQueryBuilder[Entity](s).
		SetDialect(s.getDialect()).
		Set(toKeyValues(obj, false)...).
		Where(s.pkName(), genericGetPKValue(obj)).Table(s.Table).ToSql()

	if err != nil {
		return err
	}
	_, err = s.getConnection().exec(q, args...)
	return err
}

// Delete given Entity from database
func Delete(obj Entity) error {
	s := getSchemaFor(obj)
	genericSet(obj, "deleted_at", sql.NullTime{Time: time.Now(), Valid: true})
	query, args, err := NewQueryBuilder[Entity](s).SetDialect(s.getDialect()).Table(s.Table).Where(s.pkName(), genericGetPKValue(obj)).SetDelete().ToSql()
	if err != nil {
		return err
	}
	_, err = s.getConnection().exec(query, args...)
	return err
}

func bind[T Entity](output interface{}, q string, args []interface{}) error {
	outputMD := getSchemaFor(*new(T))
	rows, err := outputMD.getConnection().query(q, args...)
	if err != nil {
		return err
	}
	return newBinder(outputMD).bind(rows, output)
}

// HasManyConfig contains all information we need for querying HasMany relationships.
// We can infer both fields if you have them in standard way but you
// can specify them if you want custom ones.
type HasManyConfig struct {
	// PropertyTable is table of the property of HasMany relationship,
	// consider `Comment` in Post and Comment relationship,
	// each Post HasMany Comment, so PropertyTable is
	// `comments`.
	PropertyTable string
	// PropertyForeignKey is the foreign key field name in the property table,
	// for example in Post HasMany Comment, if comment has `post_id` field,
	// it's the PropertyForeignKey field.
	PropertyForeignKey string
}

// HasMany configures a QueryBuilder for a HasMany relationship
// this relationship will be defined for owner argument
// that has many of PROPERTY generic type for example
// HasMany[Comment](&Post{})
// is for Post HasMany Comment relationship.
func HasMany[PROPERTY Entity](owner Entity) *QueryBuilder[PROPERTY] {
	outSchema := getSchemaFor(*new(PROPERTY))

	q := NewQueryBuilder[PROPERTY](outSchema)
	// getting config from our cache
	c, ok := getSchemaFor(owner).relations[outSchema.Table].(HasManyConfig)
	if !ok {
		q.err = fmt.Errorf("wrong config passed for HasMany")
	}

	s := getSchemaFor(owner)
	return q.
		SetDialect(s.getDialect()).
		Table(c.PropertyTable).
		Select(outSchema.Columns(true)...).
		Where(c.PropertyForeignKey, genericGetPKValue(owner))
}

// HasOneConfig contains all information we need for a HasOne relationship,
// it's similar to HasManyConfig.
type HasOneConfig struct {
	// PropertyTable is table of the property of HasOne relationship,
	// consider `HeaderPicture` in Post and HeaderPicture relationship,
	// each Post HasOne HeaderPicture, so PropertyTable is
	// `header_pictures`.
	PropertyTable string
	// PropertyForeignKey is the foreign key field name in the property table,
	// forexample in Post HasOne HeaderPicture, if header_picture has `post_id` field,
	// it's the PropertyForeignKey field.
	PropertyForeignKey string
}

// HasOne configures a QueryBuilder for a HasOne relationship
// this relationship will be defined for owner argument
// that has one of PROPERTY generic type for example
// HasOne[HeaderPicture](&Post{})
// is for Post HasOne HeaderPicture relationship.
func HasOne[PROPERTY Entity](owner Entity) *QueryBuilder[PROPERTY] {
	property := getSchemaFor(*new(PROPERTY))
	q := NewQueryBuilder[PROPERTY](property)
	c, ok := getSchemaFor(owner).relations[property.Table].(HasOneConfig)
	if !ok {
		q.err = fmt.Errorf("wrong config passed for HasOne")
	}

	// settings default config Values
	return q.
		SetDialect(property.getDialect()).
		Table(c.PropertyTable).
		Select(property.Columns(true)...).
		Where(c.PropertyForeignKey, genericGetPKValue(owner))
}

// BelongsToConfig contains all information we need for a BelongsTo relationship
// BelongsTo is a relationship between a Comment and it's Post,
// A Comment BelongsTo Post.
type BelongsToConfig struct {
	// OwnerTable is the table that contains owner of a BelongsTo
	// relationship.
	OwnerTable string
	// LocalForeignKey is name of the field that links property
	// to its owner in BelongsTo relation. for example when
	// a Comment BelongsTo Post, LocalForeignKey is
	// post_id of Comment.
	LocalForeignKey string
	// ForeignColumnName is name of the field that LocalForeignKey
	// field value will point to it, for example when
	// a Comment BelongsTo Post, ForeignColumnName is
	// id of Post.
	ForeignColumnName string
}

// BelongsTo configures a QueryBuilder for a BelongsTo relationship between
// OWNER type parameter and property argument, so
// property BelongsTo OWNER.
func BelongsTo[OWNER Entity](property Entity) *QueryBuilder[OWNER] {
	owner := getSchemaFor(*new(OWNER))
	q := NewQueryBuilder[OWNER](owner)
	c, ok := getSchemaFor(property).relations[owner.Table].(BelongsToConfig)
	if !ok {
		q.err = fmt.Errorf("wrong config passed for BelongsTo")
	}

	ownerIDidx := 0
	for idx, field := range owner.fields {
		if field.Name == c.LocalForeignKey {
			ownerIDidx = idx
		}
	}

	ownerID := genericValuesOf(property, true)[ownerIDidx]

	return q.
		SetDialect(owner.getDialect()).
		Table(c.OwnerTable).Select(owner.Columns(true)...).
		Where(c.ForeignColumnName, ownerID)

}

// BelongsToManyConfig contains information that we
// need for creating many to many queries.
type BelongsToManyConfig struct {
	// IntermediateTable is the name of the middle table
	// in a BelongsToMany (Many to Many) relationship.
	// for example when we have Post BelongsToMany
	// Category, this table will be post_categories
	// table, remember that this field cannot be
	// inferred.
	IntermediateTable string
	// IntermediatePropertyID is the name of the field name
	// of property foreign key in intermediate table,
	// for example when we have Post BelongsToMany
	// Category, in post_categories table, it would
	// be post_id.
	IntermediatePropertyID string
	// IntermediateOwnerID is the name of the field name
	// of property foreign key in intermediate table,
	// for example when we have Post BelongsToMany
	// Category, in post_categories table, it would
	// be category_id.
	IntermediateOwnerID string
	// Table name of the owner in BelongsToMany relation,
	// for example in Post BelongsToMany Category
	// Owner table is name of Category table
	// for example `categories`.
	OwnerTable string
	// OwnerLookupColumn is name of the field in the owner
	// table that is used in query, for example in Post BelongsToMany Category
	// Owner lookup field would be Category primary key which is id.
	OwnerLookupColumn string
}

// BelongsToMany configures a QueryBuilder for a BelongsToMany relationship
func BelongsToMany[OWNER Entity](property Entity) *QueryBuilder[OWNER] {
	out := *new(OWNER)
	outSchema := getSchemaFor(out)
	q := NewQueryBuilder[OWNER](outSchema)
	c, ok := getSchemaFor(property).relations[outSchema.Table].(BelongsToManyConfig)
	if !ok {
		q.err = fmt.Errorf("wrong config passed for HasMany")
	}
	return q.
		Select(outSchema.Columns(true)...).
		Table(outSchema.Table).
		WhereIn(c.OwnerLookupColumn, Raw(fmt.Sprintf(`SELECT %s FROM %s WHERE %s = ?`,
			c.IntermediatePropertyID,
			c.IntermediateTable, c.IntermediateOwnerID), genericGetPKValue(property)))
}

// Add adds `items` to `to` using relations defined between items and to in ConfigureEntity method of `to`.
func Add(to Entity, items ...Entity) error {
	if len(items) == 0 {
		return nil
	}
	rels := getSchemaFor(to).relations
	tname := getSchemaFor(items[0]).Table
	c, ok := rels[tname]
	if !ok {
		return fmt.Errorf("no config found for given to and item...")
	}
	switch c.(type) {
	case HasManyConfig:
		return addProperty(to, items...)
	case HasOneConfig:
		return addProperty(to, items[0])
	case BelongsToManyConfig:
		return addM2M(to, items...)
	default:
		return fmt.Errorf("cannot add for relation: %T", rels[getSchemaFor(items[0]).Table])
	}
}

func addM2M(to Entity, items ...Entity) error {
	//TODO: Optimize this
	rels := getSchemaFor(to).relations
	tname := getSchemaFor(items[0]).Table
	c := rels[tname].(BelongsToManyConfig)
	var values [][]interface{}
	ownerPk := genericGetPKValue(to)
	for _, item := range items {
		pk := genericGetPKValue(item)
		if isZero(pk) {
			err := Insert(item)
			if err != nil {
				return err
			}
			pk = genericGetPKValue(item)
		}
		values = append(values, []interface{}{ownerPk, pk})
	}
	i := insertStmt{
		PlaceHolderGenerator: getSchemaFor(to).getDialect().PlaceHolderGenerator,
		Table:                c.IntermediateTable,
		Columns:              []string{c.IntermediateOwnerID, c.IntermediatePropertyID},
		Values:               values,
	}

	q, args := i.ToSql()

	_, err := getConnectionFor(items[0]).DB.Exec(q, args...)
	if err != nil {
		return err
	}

	return err
}

// addHasMany(Post, comments)
func addProperty(to Entity, items ...Entity) error {
	var lastTable string
	for _, obj := range items {
		s := getSchemaFor(obj)
		if lastTable == "" {
			lastTable = s.Table
		} else {
			if lastTable != s.Table {
				return fmt.Errorf("cannot batch insert for two different tables: %s and %s", s.Table, lastTable)
			}
		}
	}
	i := insertStmt{
		PlaceHolderGenerator: getSchemaFor(to).getDialect().PlaceHolderGenerator,
		Table:                getSchemaFor(items[0]).getTable(),
	}
	ownerPKIdx := -1
	ownerPKName := getSchemaFor(items[0]).relations[getSchemaFor(to).Table].(BelongsToConfig).LocalForeignKey
	for idx, col := range getSchemaFor(items[0]).Columns(false) {
		if col == ownerPKName {
			ownerPKIdx = idx
		}
	}

	ownerPK := genericGetPKValue(to)
	if ownerPKIdx != -1 {
		cols := getSchemaFor(items[0]).Columns(false)
		i.Columns = append(i.Columns, cols...)
		// Owner PK is present in the items struct
		for _, item := range items {
			vals := genericValuesOf(item, false)
			if cols[ownerPKIdx] != getSchemaFor(items[0]).relations[getSchemaFor(to).Table].(BelongsToConfig).LocalForeignKey {
				return fmt.Errorf("owner pk idx is not correct")
			}
			vals[ownerPKIdx] = ownerPK
			i.Values = append(i.Values, vals)
		}
	} else {
		ownerPKIdx = 0
		cols := getSchemaFor(items[0]).Columns(false)
		cols = append(cols[:ownerPKIdx+1], cols[ownerPKIdx:]...)
		cols[ownerPKIdx] = getSchemaFor(items[0]).relations[getSchemaFor(to).Table].(BelongsToConfig).LocalForeignKey
		i.Columns = append(i.Columns, cols...)
		for _, item := range items {
			vals := genericValuesOf(item, false)
			if cols[ownerPKIdx] != getSchemaFor(items[0]).relations[getSchemaFor(to).Table].(BelongsToConfig).LocalForeignKey {
				return fmt.Errorf("owner pk idx is not correct")
			}
			vals = append(vals[:ownerPKIdx+1], vals[ownerPKIdx:]...)
			vals[ownerPKIdx] = ownerPK
			i.Values = append(i.Values, vals)
		}
	}

	q, args := i.ToSql()

	_, err := getConnectionFor(items[0]).DB.Exec(q, args...)
	if err != nil {
		return err
	}

	return err

}

// Query creates a new QueryBuilder for given type parameter, sets dialect and table as well.
func Query[E Entity]() *QueryBuilder[E] {
	s := getSchemaFor(*new(E))
	q := NewQueryBuilder[E](s)
	q.SetDialect(s.getDialect()).Table(s.Table)
	return q
}

// ExecRaw executes given query string and arguments on given type parameter database connection.
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

// QueryRaw queries given query string and arguments on given type parameter database connection.
func QueryRaw[OUTPUT Entity](q string, args ...interface{}) ([]OUTPUT, error) {
	o := new(OUTPUT)
	rows, err := getSchemaFor(*o).getSQLDB().Query(q, args...)
	if err != nil {
		return nil, err
	}
	var output []OUTPUT
	err = newBinder(getSchemaFor(*o)).bind(rows, &output)
	if err != nil {
		return nil, err
	}
	return output, nil
}
