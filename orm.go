package orm

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"time"

	"github.com/jedib0t/go-pretty/table"

	//Drivers
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

//Schematic prints all information ORM infered from your entities in startup, remember to pass
//your entities in Entities when you call Initialize if you want their data infered
//otherwise Schematic does not print correct data since GolobbyORM also
//incrementaly cache your entities metadata and schema.
func Schematic() {
	for conn, connObj := range globalORM {
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
		fmt.Printf("table: %s\n", t)
		w := table.NewWriter()
		w.AppendHeader(table.Row{"SQL Name", "Type", "Is Primary Key", "Is Virtual"})
		for _, field := range schema.fields {
			w.AppendRow(table.Row{field.Name, field.Type, field.IsPK, field.Virtual})
		}
		fmt.Println(w.Render())
		for table, rel := range schema.relations {
			switch rel.(type) {
			case HasOneConfig:
				fmt.Printf("%s 1-1 %s => %+v\n", t, table, rel)
			case HasManyConfig:
				fmt.Printf("%s 1-N %s => %+v\n", t, table, rel)

			case BelongsToConfig:
				fmt.Printf("%s N-1 %s => %+v\n", t, table, rel)

			case BelongsToManyConfig:
				fmt.Printf("%s N-N %s => %+v\n", t, table, rel)
			}
		}
		fmt.Println("")
	}
}

func (d *Connection) getSchema(t string) *schema {
	return d.Schemas[t]
}

func getDialectForConnection(connection string) *Dialect {
	return GetConnection(connection).Dialect
}

func (d *Connection) setSchema(e Entity, s *schema) {
	var configurator EntityConfigurator
	e.ConfigureEntity(&configurator)
	d.Schemas[configurator.table] = s
}

var globalORM = map[string]*Connection{}

func GetConnection(name string) *Connection {
	return globalORM[name]
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

func initTableName(e Entity) string {
	configurator := newEntityConfigurator()
	e.ConfigureEntity(configurator)

	if configurator.table == "" {
		panic("table name is mandatory for entities")
	}
	return configurator.table
}

//Initialize gets list of ConnectionConfig and builds up ORM for you.
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
	globalORM[fmt.Sprintf("%s", name)] = s
	return s
}

type Entity interface {
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

// insertStmt given Entity
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

// Save upserts given entity.
func Save(obj Entity) error {
	if isZero(getSchemaFor(obj).getPK(obj)) {
		return Insert(obj)
	} else {
		return Update(obj)
	}
}

// Find finds the Entity you want based on Entity generic type and primary key you passed.
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

// Update given Entity in database
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

type HasManyConfig struct {
	PropertyTable      string
	PropertyForeignKey string
}

func HasMany[OUT Entity](owner Entity) *QueryBuilder[OUT] {
	outSchema := getSchemaFor(*new(OUT))
	// getting config from our cache
	c, ok := getSchemaFor(owner).relations[outSchema.Table].(HasManyConfig)
	if !ok {
		panic("wrong config passed for HasMany")
	}

	s := getSchemaFor(owner)
	return NewQueryBuilder[OUT]().SetDialect(s.getDialect()).Table(c.PropertyTable).Select(outSchema.Columns(true)...).Where(c.PropertyForeignKey, genericGetPKValue(owner))
}

type HasOneConfig struct {
	PropertyTable      string
	PropertyForeignKey string
}

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

type BelongsToConfig struct {
	OwnerTable        string
	LocalForeignKey   string
	ForeignColumnName string
}

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

type BelongsToManyConfig struct {
	IntermediateTable      string
	IntermediatePropertyID string
	IntermediateOwnerID    string
	ForeignTable           string
	ForeignLookupColumn    string
}

//BelongsToMany
func BelongsToMany[OWNER Entity](property Entity) *QueryBuilder[OWNER] {
	out := new(OWNER)
	c, ok := getSchemaFor(property).relations[getSchemaFor(*out).Table].(BelongsToManyConfig)
	if !ok {
		panic("wrong config passed for HasMany")
	}
	return NewQueryBuilder[OWNER]().
		Select(getSchemaFor(*out).Columns(true)...).
		Table(getSchemaFor(*out).Table).
		WhereIn(c.ForeignLookupColumn, Raw(fmt.Sprintf(`SELECT %s FROM %s WHERE %s = ?`,
			c.IntermediateOwnerID,
			c.IntermediateTable, c.IntermediatePropertyID), genericGetPKValue(property)))
}

//Add adds `items` to `to` using relations defined between items and to in ConfigureRelations method of `to`.
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

func Query[E Entity]() *QueryBuilder[E] {
	q := NewQueryBuilder[E]()
	s := getSchemaFor(*new(E))
	q.SetDialect(s.getDialect()).Table(s.Table)
	return q
}

func Exec[E Entity](stmt *QueryBuilder[E]) (lastInsertedId int64, rowsAffected int64, err error) {
	e := new(E)
	s := getSchemaFor(*e)
	var lastInsertedID int64
	stmt.SetDialect(s.getDialect()).Table(s.Table)
	q, args, err := stmt.ToSql()
	if err != nil {
		return -1, -1, err
	}
	res, err := s.getSQLDB().Exec(q, args...)
	if err != nil {
		return 0, 0, err
	}

	if stmt.typ == queryType_UPDATE {
		lastInsertedID, err = res.LastInsertId()
		if err != nil {
			return 0, 0, err
		}
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return 0, 0, err
	}

	return lastInsertedID, affected, nil
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
	err = getSchemaFor(*o).bind(rows, &output)
	if err != nil {
		return nil, err
	}
	return output, nil
}
