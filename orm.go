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
var globalLogger Logger

// Schematic prints all information ORM inferred from your entities in startup, remember to pass
// your entities in Entities when you call SetupConnection if you want their data inferred
// otherwise Schematic does not print correct data since GoLobby ORM also
// incrementally cache your entities metadata and schema.
func Schematic() {
	for conn, connObj := range globalConnections {
		fmt.Printf("----------------%s---------------\n", conn)
		connObj.Schematic()
		fmt.Println("-----------------------------------")
	}
}

type Config struct {
	// LogLevel
	LogLevel LogLevel
}

type ConnectionConfig struct {
	// Name of your database connection, it's up to you to name them anything
	// just remember that having a connection name is mandatory if
	// you have multiple connections
	Name string
	// Which driver to be used for your connection, you can name [mysql, sqlite3, postgres]
	Driver string
	// Which connection string to be passed as second argument to sql.Open(driver, DSN)
	DSN string
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
}

// SetupConnection declares a new connection for ORM.
func SetupConnection(conf ConnectionConfig) error {
	// configure logger
	// TODO: remove logger
	var err error
	globalLogger, err = newZapLogger(LogLevelDev)
	if err != nil {
		return err
	}

	var dialect *Dialect
	var db *sql.DB
	if conf.DB != nil && conf.Dialect != nil {
		globalLogger.Infof("Configuring an open connection")
		dialect = conf.Dialect
		db = conf.DB
	} else {
		globalLogger.Infof("Opening and configuring a connection using %s", conf.Driver)
		dialect, err = getDialect(conf.Driver)
		if err != nil {
			return err
		}
		db, err = sql.Open(conf.Driver, conf.DSN)
		if err != nil {
			return err
		}
	}
	conf.DB = db
	conf.Dialect = dialect

	_, err = initialize(conf)
	if err != nil {
		return err
	}

	return nil
}

func initialize(config ConnectionConfig) (*connection, error) {
	schemas := map[string]*schema{}
	if config.Name == "" {
		config.Name = "default"
	}
	globalLogger.Infof("Generating schema definitions for connection %s entities", config.Name)
	globalLogger.Infof("Entities are: %v", entitiesAsList(config.Entities))
	for _, entity := range config.Entities {
		s := schemaOfHeavyReflectionStuff(entity)
		var configurator EntityConfigurator
		entity.ConfigureEntity(&configurator)
		schemas[configurator.table] = s
	}
	s := &connection{
		Name:       config.Name,
		Connection: config.DB,
		Schemas:    schemas,
		Dialect:    config.Dialect,
	}
	globalConnections[fmt.Sprintf("%s", config.Name)] = s
	globalLogger.Infof("%s registered successfully.", config.Name)
	return s, nil
}

// Entity defines the interface that each of your structs that
// you want to use as database entities should have,
// it's a simple one and its ConfigureEntity.
type Entity interface {
	// ConfigureEntity should be defined for all of your database entities
	// and it can define Table, Connection and also relations of your Entity.
	ConfigureEntity(e *EntityConfigurator)
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
	globalLogger.Debugf("Going to insert %d objects", len(objs))
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

	res, err := s.getConnection().exec(q, args...)
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
		globalLogger.Debugf("Given object has no primary key set, going to insert it.")
		return Insert(obj)
	} else {
		globalLogger.Debugf("Given object has primary key set, going for update.")
		return Update(obj)
	}
}

// Find finds the Entity you want based on generic type and primary key you passed.
func Find[T Entity](id interface{}) (T, error) {
	var q string
	out := new(T)
	md := getSchemaFor(*out)
	q, args, err := NewQueryBuilder[T]().
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
	_, err = s.getConnection().exec(q, args...)
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
	_, err = s.getConnection().exec(query, args...)
	return err
}

func bind[T Entity](output interface{}, q string, args []interface{}) error {
	outputMD := getSchemaFor(*new(T))
	rows, err := outputMD.getConnection().query(q, args...)
	if err != nil {
		return err
	}
	return newBinder[T](outputMD).bind(rows, output)
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
	// forexample in Post HasMany Comment, if comment has `post_id` field,
	// it's the PropertyForeignKey field.
	PropertyForeignKey string
}

// HasMany configures a QueryBuilder for a HasMany relationship
// this relationship will be defined for owner argument
// that has many of PROPERTY generic type for example
// HasMany[Comment](&Post{})
// is for Post HasMany Comment relationship.
func HasMany[PROPERTY Entity](owner Entity) *QueryBuilder[PROPERTY] {
	q := NewQueryBuilder[PROPERTY]()
	outSchema := getSchemaFor(*new(PROPERTY))
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
	q := NewQueryBuilder[PROPERTY]()
	property := getSchemaFor(*new(PROPERTY))
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
	q := NewQueryBuilder[OWNER]()
	owner := getSchemaFor(*new(OWNER))
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
	q := NewQueryBuilder[OWNER]()
	out := new(OWNER)
	c, ok := getSchemaFor(property).relations[getSchemaFor(*out).Table].(BelongsToManyConfig)
	if !ok {
		q.err = fmt.Errorf("wrong config passed for HasMany")
	}
	return q.
		Select(getSchemaFor(*out).Columns(true)...).
		Table(getSchemaFor(*out).Table).
		WhereIn(c.OwnerLookupColumn, Raw(fmt.Sprintf(`SELECT %s FROM %s WHERE %s = ?`,
			c.IntermediateOwnerID,
			c.IntermediateTable, c.IntermediatePropertyID), genericGetPKValue(property)))
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
		return fmt.Errorf("adding to a belongs to many relation is not implemented yet")
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

	_, err := getConnectionFor(items[0]).Connection.Exec(q, args...)
	if err != nil {
		return err
	}

	return err

}

// Query creates a new QueryBuilder for given type parameter, sets dialect and table as well.
func Query[E Entity]() *QueryBuilder[E] {
	q := NewQueryBuilder[E]()
	s := getSchemaFor(*new(E))
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
	err = newBinder[OUTPUT](getSchemaFor(*o)).bind(rows, &output)
	if err != nil {
		return nil, err
	}
	return output, nil
}
