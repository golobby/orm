package orm

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"time"

	"github.com/jedib0t/go-pretty/table"

	// Drivers
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

var globalConnections = map[string]*Connection{}

// Schematic prints all information ORM inferred from your entities in startup, remember to pass
// your entities in Entities when you call Initialize if you want their data inferred
// otherwise Schematic does not print correct data since GoLobby ORM also
// incrementally cache your entities metadata and schema.
func Schematic() {
	for conn, connObj := range globalConnections {
		fmt.Printf("----------------%s---------------\n", conn)
		connObj.Schematic()
		fmt.Println("-----------------------------------")
	}
}

type Connection struct {
	Name       string
	Dialect    *Dialect
	Connection *sql.DB
	Schemas    map[string]*schema
}

func (c *Connection) Schematic() {
	fmt.Printf("SQL Dialect: %s\n", c.Dialect.DriverName)
	for t, schema := range c.Schemas {
		fmt.Printf("t: %s\n", t)
		w := table.NewWriter()
		w.AppendHeader(table.Row{"SQL Name", "Type", "Is Primary Key", "Is Virtual"})
		for _, field := range schema.fields {
			w.AppendRow(table.Row{field.Name, field.Type, field.IsPK, field.Virtual})
		}
		fmt.Println(w.Render())
		for t, rel := range schema.relations {
			switch rel.(type) {
			case HasOneConfig:
				fmt.Printf("%s 1-1 %s => %+v\n", t, t, rel)
			case HasManyConfig:
				fmt.Printf("%s 1-N %s => %+v\n", t, t, rel)

			case BelongsToConfig:
				fmt.Printf("%s N-1 %s => %+v\n", t, t, rel)

			case BelongsToManyConfig:
				fmt.Printf("%s N-N %s => %+v\n", t, t, rel)
			}
		}
		fmt.Println("")
	}
}

func (c *Connection) getSchema(t string) *schema {
	return c.Schemas[t]
}

func (c *Connection) setSchema(e Entity, s *schema) {
	var configurator EntityConfigurator
	e.ConfigureEntity(&configurator)
	c.Schemas[configurator.table] = s
}

func GetConnection(name string) *Connection {
	return globalConnections[name]
}

type ConnectionConfig struct {
	// Name of your database connection, it's up to you to name them anything
	// just remember that having a connection name is mandatory if
	// you have multiple connections
	Name string
	// Which driver to be used for your connection, you can name [mysql, sqlite3, postgres]
	Driver string
	// Which connection string to be passed as second argument to sql.Open(driver, connectionString)
	ConnectionString string
	// If you already have an active database connection configured pass it in this value and
	// do not pass Driver and ConnectionString fields.
	DB *sql.DB
	// Which dialect of sql to generate queries for, you don't need it most of the times when you are using
	// traditional databases such as mysql, sqlite3, postgres.
	Dialect *Dialect
	// List of entities that you want to use for this connection, remember that you can ignore this field
	// and GolobbyORM will build our metadata cache incrementally but you will lose schematic
	// information that we can provide you and also potentialy validations that we
	// can do with the database
	Entities []Entity
}

//Initialize gets list of ConnectionConfig and builds up ORM for you.
func Initialize(configs ...ConnectionConfig) error {
	for _, conf := range configs {
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
		initialize(conf.Name, dialect, db, conf.Entities)
	}
	return nil
}

func initialize(name string, dialect *Dialect, db *sql.DB, entities []Entity) *Connection {
	schemas := map[string]*schema{}
	if name == "" {
		name = "default"
	}
	for _, entity := range entities {
		s := schemaOf(entity)
		var configurator EntityConfigurator
		entity.ConfigureEntity(&configurator)
		schemas[configurator.table] = s
	}
	s := &Connection{
		Name:       name,
		Connection: db,
		Schemas:    schemas,
		Dialect:    dialect,
	}
	globalConnections[fmt.Sprintf("%s", name)] = s
	return s
}

//Entity defines the interface that each of your structs that
//you want to use as database entities should have,
//it's a simple one and its ConfigureEntity.
type Entity interface {
	//ConfigureEntity should be defined for all of your database entities
	//and it can define Table, Connection and also relations of your Entity.
	ConfigureEntity(e *EntityConfigurator)
}

func getDB(driver string, connectionString string) (*sql.DB, error) {
	return sql.Open(driver, connectionString)
}

func getDialect(driver string) (*Dialect, error) {
	switch driver {
	case "mysql":
		return Dialects.MySQL, nil
	case "sqlite", "sqlite3":
		return Dialects.SQLite3, nil
	case "postgres":
		return Dialects.PostgreSQL, nil
	default:
		return nil, fmt.Errorf("err no dialect matched with driver")
	}
}

// Insert given entities into database based on their ConfigureEntity
// we can find table and also Connection name.
func Insert(objs ...Entity) error {
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

	q, args := insertStmt{
		PlaceHolderGenerator: s.getDialect().PlaceHolderGenerator,
		Table:                s.getTable(),
		Columns:              cols,
		Values:               values,
	}.ToSql()

	res, err := s.getSQLDB().Exec(q, args...)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}

	getSchemaFor(objs[len(objs)-1]).setPK(objs[len(objs)-1], id)
	return nil
}

func isZero(val interface{}) bool {
	switch val.(type) {
	case int64:
		return val.(int64) == 0
	case int:
		return val.(int) == 0
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
	q, args, err := NewQueryBuilder[T]().SetDialect(md.getDialect()).Table(md.Table).Select(md.Columns(true)...).Where(md.pkName(), id).ToSql()
	if err != nil {
		return *out, err
	}
	err = bindContext[T](context.Background(), out, q, args)

	if err != nil {
		return *out, err
	}

	return *out, nil
}

func toTuples(obj Entity, withPK bool) [][2]interface{} {
	var tuples [][2]interface{}
	vs := genericValuesOf(obj, withPK)
	cols := getSchemaFor(obj).Columns(withPK)
	for i, col := range cols {
		tuples = append(tuples, [2]interface{}{
			col,
			vs[i],
		})
	}
	return tuples
}

// Update given Entity in database.
func Update(obj Entity) error {
	s := getSchemaFor(obj)
	q, args, err := NewQueryBuilder[Entity]().SetDialect(s.getDialect()).Sets(toTuples(obj, false)...).Where(s.pkName(), genericGetPKValue(obj)).Table(s.Table).ToSql()

	if err != nil {
		return err
	}
	_, err = s.getSQLDB().Exec(q, args...)
	return err
}

// Delete given Entity from database
func Delete(obj Entity) error {
	s := getSchemaFor(obj)
	genericSet(obj, "deleted_at", sql.NullTime{Time: time.Now(), Valid: true})
	query, args, err := NewQueryBuilder[Entity]().SetDialect(s.getDialect()).Table(s.Table).Where(s.pkName(), genericGetPKValue(obj)).SetDelete().ToSql()
	if err != nil {
		return err
	}
	_, err = getSchemaFor(obj).getSQLDB().Exec(query, args...)
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

//HasManyConfig contains all information we need for querying HasMany relationships.
//We can infer both fields if you have them in standard way but you
//can specify them if you want custom ones.
type HasManyConfig struct {
	// PropertyTable is table of the property of HasMany relationship,
	// consider `Comment` in Post and Comment relationship,
	// each Post HasMany Comment, so PropertyTable is
	// `comments`.
	PropertyTable string
	// PropertyForeignKey is the foreign key column name in the property table,
	// forexample in Post HasMany Comment, if comment has `post_id` column,
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
	// getting config from our cache
	c, ok := getSchemaFor(owner).relations[outSchema.Table].(HasManyConfig)
	if !ok {
		panic("wrong config passed for HasMany")
	}

	s := getSchemaFor(owner)
	return NewQueryBuilder[PROPERTY]().SetDialect(s.getDialect()).Table(c.PropertyTable).Select(outSchema.Columns(true)...).Where(c.PropertyForeignKey, genericGetPKValue(owner))
}

//HasOneConfig contains all information we need for a HasOne relationship,
//it's similar to HasManyConfig.
type HasOneConfig struct {
	// PropertyTable is table of the property of HasOne relationship,
	// consider `HeaderPicture` in Post and HeaderPicture relationship,
	// each Post HasOne HeaderPicture, so PropertyTable is
	// `header_pictures`.
	PropertyTable string
	// PropertyForeignKey is the foreign key column name in the property table,
	// forexample in Post HasOne HeaderPicture, if header_picture has `post_id` column,
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
	c, ok := getSchemaFor(owner).relations[property.Table].(HasOneConfig)
	if !ok {
		panic("wrong config passed for HasOne")
	}

	//settings default config Values
	return NewQueryBuilder[PROPERTY]().SetDialect(property.getDialect()).Table(c.PropertyTable).
		Select(property.Columns(true)...).Where(c.PropertyForeignKey, genericGetPKValue(owner))
}

//BelongsToConfig contains all information we need for a BelongsTo relationship
//BelongsTo is a relationship between a Comment and it's Post,
//A Comment BelongsTo Post.
type BelongsToConfig struct {
	//OwnerTable is the table that contains owner of a BelongsTo
	//relationship.
	OwnerTable string
	//LocalForeignKey is name of the column that links property
	//to its owner in BelongsTo relation. for example when
	//a Comment BelongsTo Post, LocalForeignKey is
	//post_id of Comment.
	LocalForeignKey string
	//ForeignColumnName is name of the column that LocalForeignKey
	//column value will point to it, for example when
	//a Comment BelongsTo Post, ForeignColumnName is
	//id of Post.
	ForeignColumnName string
}

//BelongsTo configures a QueryBuilder for a BelongsTo relationship between
//OWNER type parameter and property argument, so
//property BelongsTo OWNER.
func BelongsTo[OWNER Entity](property Entity) *QueryBuilder[OWNER] {
	owner := getSchemaFor(*new(OWNER))
	c, ok := getSchemaFor(property).relations[owner.Table].(BelongsToConfig)
	if !ok {
		panic("wrong config passed for BelongsTo")
	}

	ownerIDidx := 0
	for idx, field := range owner.fields {
		if field.Name == c.LocalForeignKey {
			ownerIDidx = idx
		}
	}

	ownerID := genericValuesOf(property, true)[ownerIDidx]

	return NewQueryBuilder[OWNER]().
		SetDialect(owner.getDialect()).
		Table(c.OwnerTable).Select(owner.Columns(true)...).
		Where(c.ForeignColumnName, ownerID)

}

//BelongsToManyConfig contains information that we
//need for creating many to many queries.
type BelongsToManyConfig struct {
	//IntermediateTable is the name of the middle table
	//in a BelongsToMany (Many to Many) relationship.
	//for example when we have Post BelongsToMany
	//Category, this table will be post_categories
	//table, remember that this field cannot be
	//inferred.
	IntermediateTable string
	//IntermediatePropertyID is the name of the column name
	//of property foreign key in intermediate table,
	//for example when we have Post BelongsToMany
	//Category, in post_categories table, it would
	//be post_id.
	IntermediatePropertyID string
	//IntermediateOwnerID is the name of the column name
	//of property foreign key in intermediate table,
	//for example when we have Post BelongsToMany
	//Category, in post_categories table, it would
	//be category_id.
	IntermediateOwnerID string
	//Table name of the owner in BelongsToMany relation,
	//for example in Post BelongsToMany Category
	//Owner table is name of Category table
	//for example `categories`.
	OwnerTable string
	//OwnerLookupColumn is name of the column in the owner
	//table that is used in query, for example in Post BelongsToMany Category
	//Owner lookup column would be Category primary key which is id.
	OwnerLookupColumn string
}

//BelongsToMany configures a QueryBuilder for a BelongsToMany relationship
func BelongsToMany[OWNER Entity](property Entity) *QueryBuilder[OWNER] {
	out := new(OWNER)
	c, ok := getSchemaFor(property).relations[getSchemaFor(*out).Table].(BelongsToManyConfig)
	if !ok {
		panic("wrong config passed for HasMany")
	}
	return NewQueryBuilder[OWNER]().
		Select(getSchemaFor(*out).Columns(true)...).
		Table(getSchemaFor(*out).Table).
		WhereIn(c.OwnerLookupColumn, Raw(fmt.Sprintf(`SELECT %s FROM %s WHERE %s = ?`,
			c.IntermediateOwnerID,
			c.IntermediateTable, c.IntermediatePropertyID), genericGetPKValue(property)))
}

//Add adds `items` to `to` using relations defined between items and to in ConfigureEntity method of `to`.
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
		panic("not implemented yet")
	default:
		return fmt.Errorf("cannot add for relation: %T", rels[getSchemaFor(items[0]).Table])
	}
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
				panic("owner pk idx is not correct")
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
				panic("owner pk idx is not correct")
			}
			vals = append(vals[:ownerPKIdx+1], vals[ownerPKIdx:]...)
			vals[ownerPKIdx] = ownerPK
			i.Values = append(i.Values, vals)
		}
	}

	q, args := i.ToSql()

	_, err := getConnectionFor(items[0]).Connection.Exec(q, args...)
	if err != nil {
		return err
	}

	return err

}

// addBelongsToMany(Post, Category)
func addBelongsToMany(to Entity, items ...Entity) error {
	return nil
}

//Query creates a new QueryBuilder for given type parameter, sets dialect and table as well.
func Query[E Entity]() *QueryBuilder[E] {
	q := NewQueryBuilder[E]()
	s := getSchemaFor(*new(E))
	q.SetDialect(s.getDialect()).Table(s.Table)
	return q
}

//ExecRaw executes given query string and arguments on given type parameter database connection.
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

//QueryRaw queries given query string and arguments on given type parameter database connection.
func QueryRaw[OUTPUT Entity](q string, args ...interface{}) ([]OUTPUT, error) {
	o := new(OUTPUT)
	rows, err := getSchemaFor(*o).getSQLDB().Query(q, args...)
	if err != nil {
		return nil, err
	}
	var output []OUTPUT
	err = getSchemaFor(*o).bind(rows, &output)
	if err != nil {
		return nil, err
	}
	return output, nil
}
