package orm

import (
	"database/sql"
	"fmt"
)

type DB struct {
	name      string
	dialect   *dialect
	conn      *sql.DB
	metadatas map[string]*objectMetadata
}

var globalORM = map[string]*DB{}

type ConnectionConfig struct {
	Name     string
	Dialect  int
	Entities []IsEntity
}

func Initialize(confs ...ConnectionConfig) error {
	for _, conf := range confs {
		initialize(conf.Name, getDialect(conf.Dialect), conf.Entities)
	}
	return nil
}

func initialize(name string, dialect *dialect, entities []IsEntity) *DB {
	metadatas := map[string]*objectMetadata{}
	for _, entity := range entities {
		metadatas[fmt.Sprintf("%s", entity.EntityConfig().Table)] = objectMetadataFrom(entity, dialect)
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
