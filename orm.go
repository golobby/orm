package orm

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"

	"github.com/golobby/orm/qb"
	"github.com/jedib0t/go-pretty/table"

	//Drivers
	_ "github.com/mattn/go-sqlite3"
)

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
		fmt.Printf("Table: %s\n", t)
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

var globalORM = map[string]*Connection{}

func GetConnection(name string) *Connection {
	return globalORM[name]
}

type ConnectionConfig struct {
	Name             string
	Driver           string
	ConnectionString string
	DB               *sql.DB
	Dialect          *Dialect
	Entities         []Entity
}

func initTableName(e Entity) string {
	configurator := newEntityConfigurator()
	e.ConfigureEntity(configurator)

	if configurator.table == "" {
		panic("Table name is mandatory for entities")
	}
	return configurator.table
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
		initialize(conf.Name, dialect, db, conf.Entities)
	}
	return nil
}

func initialize(name string, dialect *Dialect, db *sql.DB, entities []Entity) *Connection {
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
	ConfigureRelations(r *RelationConfigurator)
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

// Insert given Entity
func Insert(objs ...Entity) error {
	if len(objs) == 0 {
		return nil
	}
	s := getSchemaFor(objs[0])
	cols := s.Columns(false)
	var values [][]interface{}
	for _, obj := range objs {
		values = append(values, genericValuesOf(obj, false))
	}

	q, args := qb.Insert{
		PlaceholderGenerator: s.dialect.PlaceHolderGenerator,
		Into:                 s.getTable(),
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

	q, args, err := qb.Select{
		PlaceholderGenerator: md.dialect.PlaceHolderGenerator,
		Table:                md.Table,
		Selected:             &qb.Selected{Columns: md.Columns(true)},
		Where: &qb.Where{BinaryOp: qb.BinaryOp{
			Lhs: md.pkName(),
			Op:  qb.Eq,
			Rhs: id,
		}},
	}.ToSql()
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
	ph := getSchemaFor(obj).getDialect().PlaceholderChar
	s := getSchemaFor(obj)
	if getSchemaFor(obj).getDialect().IncludeIndexInPlaceholder {
		ph = ph + "1"
	}
	q, args := qb.Update{
		PlaceHolderGenerator: s.dialect.PlaceHolderGenerator,
		Table:                s.Table,
		Set:                  toTuples(obj, false),
		Where: &qb.Where{
			BinaryOp: qb.BinaryOp{
				Lhs: s.pkName(),
				Op:  qb.Eq,
				Rhs: genericGetPKValue(obj),
			},
		},
	}.ToSql()

	_, err := getSchemaFor(obj).getSQLDB().Exec(q, args...)
	return err
}

// Delete given Entity from database
func Delete(obj Entity) error {
	s := getSchemaFor(obj)

	q, args := qb.Delete{
		PlaceHolderGenerator: s.dialect.PlaceHolderGenerator,
		From:                 s.Table,
		Where: &qb.Where{BinaryOp: qb.BinaryOp{
			Lhs: s.pkName(),
			Op:  qb.Eq,
			Rhs: genericGetPKValue(obj),
		}},
	}.ToSql()

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

func HasMany[OUT Entity](owner Entity) ([]OUT, error) {
	outSchema := getSchemaFor(*new(OUT))
	// getting config from our cache
	c, ok := getSchemaFor(owner).relations[outSchema.Table].(HasManyConfig)
	if !ok {
		return nil, fmt.Errorf("wrong config passed for HasMany")
	}

	var out []OUT
	s := getSchemaFor(owner)
	var q string
	var args []interface{}
	q, args, err := qb.Select{
		PlaceholderGenerator: s.dialect.PlaceHolderGenerator,
		Table:                c.PropertyTable,
		Selected:             &qb.Selected{Columns: outSchema.Columns(true)},
		Where: &qb.Where{BinaryOp: qb.BinaryOp{
			Lhs: c.PropertyForeignKey,
			Op:  qb.Eq,
			Rhs: genericGetPKValue(owner),
		}},
	}.ToSql()

	if err != nil {
		return nil, err
	}

	err = bindContext[OUT](context.Background(), &out, q, args)

	if err != nil {
		return nil, err
	}

	return out, nil
}

type HasOneConfig struct {
	PropertyTable      string
	PropertyForeignKey string
}

func HasOne[PROPERTY Entity](owner Entity) (PROPERTY, error) {
	out := new(PROPERTY)
	property := getSchemaFor(*new(PROPERTY))
	c, ok := getSchemaFor(owner).relations[property.Table].(HasOneConfig)
	if !ok {
		return *new(PROPERTY), fmt.Errorf("wrong config passed for HasOne")
	}
	//settings default config Values

	var q string
	var args []interface{}
	q, args, err := qb.Select{
		PlaceholderGenerator: property.dialect.PlaceHolderGenerator,
		Table:                c.PropertyTable,
		Selected:             &qb.Selected{Columns: property.Columns(true)},
		Where: &qb.Where{BinaryOp: qb.BinaryOp{
			Lhs: c.PropertyForeignKey,
			Op:  qb.Eq,
			Rhs: genericGetPKValue(owner),
		}},
	}.ToSql()

	if err != nil {
		return *out, err
	}
	err = bindContext[PROPERTY](context.Background(), out, q, args)
	return *out, err
}

type BelongsToConfig struct {
	OwnerTable        string
	LocalForeignKey   string
	ForeignColumnName string
}

func BelongsTo[OWNER Entity](property Entity) (OWNER, error) {
	out := new(OWNER)
	owner := getSchemaFor(*new(OWNER))
	c, ok := getSchemaFor(property).relations[owner.Table].(BelongsToConfig)
	if !ok {
		return *new(OWNER), fmt.Errorf("wrong config passed for BelongsTo")
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
	q, args, err := qb.Select{
		PlaceholderGenerator: owner.dialect.PlaceHolderGenerator,
		Table:                c.OwnerTable,
		Selected:             &qb.Selected{Columns: owner.Columns(true)},
		Where: &qb.Where{BinaryOp: qb.BinaryOp{
			Lhs: c.ForeignColumnName,
			Op:  qb.Eq,
			Rhs: ownerID,
		}},
	}.ToSql()

	if err != nil {
		return *out, err
	}
	err = bindContext[OWNER](context.Background(), out, q, args)
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
func BelongsToMany[OWNER Entity](property Entity) ([]OWNER, error) {
	out := new(OWNER)
	c, ok := getSchemaFor(property).relations[getSchemaFor(*out).Table].(BelongsToManyConfig)
	if !ok {
		return nil, fmt.Errorf("wrong config passed for HasMany")
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
	i := qb.Insert{
		PlaceholderGenerator: getSchemaFor(to).dialect.PlaceHolderGenerator,
		Into:                 getSchemaFor(items[0]).getTable(),
	}
	var ownerPKIdx int

	for idx, col := range getSchemaFor(items[0]).Columns(false) {
		if col == getSchemaFor(items[0]).relations[getSchemaFor(to).Table].(BelongsToConfig).LocalForeignKey {
			ownerPKIdx = idx
		}
	}

	ownerPK := genericGetPKValue(to)
	if ownerPKIdx != 0 {
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
		cols := getSchemaFor(items[0]).Columns(false)
		cols2 := append(cols[:ownerPKIdx], getSchemaFor(items[0]).relations[getSchemaFor(to).Table].(BelongsToConfig).LocalForeignKey)
		cols2 = append(cols2, cols[ownerPKIdx+1:]...)
		cols = cols2
		i.Columns = append(i.Columns, cols...)
		for _, item := range items {
			vals := genericValuesOf(item, false)
			if cols[ownerPKIdx] != getSchemaFor(items[0]).relations[getSchemaFor(to).Table].(BelongsToConfig).LocalForeignKey {
				panic("owner pk idx is not correct")
			}
			vals2 := append(vals[:ownerPKIdx], ownerPK)
			vals2 = append(vals2, vals[ownerPKIdx+1:]...)
			vals = vals2
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
func makeEntity(obj Entity) *EntityConfigurator {
	var configurator EntityConfigurator
	obj.ConfigureEntity(&configurator)
	return &configurator
}

//func Query[OUTPUT Entity](mods ...qm.QM) ([]OUTPUT, error) {
//	o := new(OUTPUT)
//	s := qb.()
//
//	s.Table(makeEntity(*o).table)
//
//	for _, mod := range mods {
//		mod.Modify(s)
//	}
//
//	q, args := s.Build()
//	rows, err := getSchemaFor(*o).getSQLDB().Query(q, args...)
//	if err != nil {
//		return nil, err
//	}
//	var output []OUTPUT
//	err = getSchemaFor(*o).bind(rows, output)
//	if err != nil {
//		return nil, err
//	}
//	return output, nil
//
//}

func Exec[E Entity](stmt qb.ToSql) (int64, int64, error) {
	e := new(E)

	res, err := getSchemaFor(*e).getSQLDB().Exec(stmt.ToSql())
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
