package orm

import (
	"context"
	"database/sql"
	"fmt"
)

// BaseMetadata contains common behaviours for entity structs.
type BaseMetadata struct {
	table      string
	connection string
	getPK      func(obj interface{}) interface{}
	setPK      func(obj interface{}, value interface{})
	fields     func() []*Field
	values     func(obj interface{}, withPK bool) []interface{}
}

func (e *BaseMetadata) getMetadata() *EntityMetadata {
	db := e.getDB()
	return db.metadatas[e.getTable()]
}

func (e *BaseMetadata) getTable() string {
	return e.table
}

func (e *BaseMetadata) getConnection() *sql.DB {
	return e.getDB().conn
}

func (e *BaseMetadata) getDialect() *Dialect {
	return e.getMetadata().dialect
}

func (e *BaseMetadata) getDB() *DB {
	if len(globalORM) > 1 && (e.connection == "" || e.getTable() == "") {
		panic("need table and connection name when having more than 1 connection registered")
	}
	if len(globalORM) == 1 {
		for _, db := range globalORM {
			return db
		}
	}
	if db, exists := globalORM[fmt.Sprintf("%s", e.connection)]; exists {
		return db
	}
	panic("no db found")

}

func (e *BaseMetadata) getFields() []*Field {
	var fields []*Field
	if e.fields != nil {
		fields = e.fields()
	} else {
		fields = e.getMetadata().Fields
	}
	return fields
}
func (e *BaseMetadata) QueryContext(ctx context.Context, q string, args ...interface{}) (*sql.Rows, error) {
	return e.getConnection().QueryContext(ctx, q, args...)
}

func (e *BaseMetadata) BindContext(ctx context.Context, out interface{}, q string, args ...interface{}) error {
	rows, err := e.QueryContext(ctx, q, args...)
	if err != nil {
		return err
	}
	return e.getMetadata().Bind(rows, out)
}
