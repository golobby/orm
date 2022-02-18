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
	Name     string
	Dialect  *Dialect
	Entities []IsEntity
}

func initTableName(e IsEntity) string {
	if e.EntityConfig().Table != "" {
		return e.EntityConfig().Table
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
		initialize(conf.Name, conf.Dialect, conf.Entities)
	}
	return nil
}

func initialize(name string, dialect *Dialect, entities []IsEntity) *DB {
	metadatas := map[string]*objectMetadata{}
	for _, entity := range entities {
		metadatas[fmt.Sprintf("%s", initTableName(entity))] = objectMetadataFrom(entity, dialect)
	}
	s := &DB{
		name:      name,
		conn:      &sql.DB{}, //TODO: Fix me
		metadatas: metadatas,
		dialect:   dialect,
	}
	globalORM[fmt.Sprintf("%s", name)] = s
	return s
}
