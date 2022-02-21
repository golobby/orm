package orm

import (
	"database/sql"
	"fmt"
	"github.com/gertd/go-pluralize"
	"github.com/iancoleman/strcase"
	"reflect"
	"strings"
)

type DB struct {
	name      string
	dialect   *Dialect
	conn      *sql.DB
	metadatas map[string]*objectMetadata
}

var globalORM = map[string]*DB{}

type ConnectionConfig struct {
	Name             string
	Driver           string
	ConnectionString string
	DB               *sql.DB
	Dialect          *Dialect
	Entities         []IsEntity
}

func initTableName(e IsEntity) string {
	if e.E().table != "" {
		return e.E().table
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
		initialize(conf.Name, dialect, db, conf.Entities)
	}
	return nil
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

func initialize(name string, dialect *Dialect, db *sql.DB, entities []IsEntity) *DB {
	metadatas := map[string]*objectMetadata{}
	for _, entity := range entities {
		metadatas[fmt.Sprintf("%s", initTableName(entity))] = objectMetadataFrom(entity, dialect)
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
