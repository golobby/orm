package orm

import (
	"context"
	"database/sql"
	"fmt"
)

// BaseEntity contains common behaviours for entity structs.
type BaseEntity struct {
	table      string
	connection string
	getPK      func(obj interface{}) interface{}
	setPK      func(obj interface{}, value interface{})
	fields     func() []*Field
	values     func(obj interface{}, withPK bool) []interface{}
}

func NewBaseEntity() {

}

func (b *BaseEntity) HasOne(output interface{}, config HasOneConfig) Relation {
	return Relation{
		output: output,
		c:      config,
	}
}

func (e *BaseEntity) getMetadata() *objectMetadata {
	db := e.getDB()
	return db.metadatas[e.getTable()]
}

func (e *BaseEntity) getTable() string {
	return e.table
}

func (e *BaseEntity) getConnection() *sql.DB {
	return e.getDB().conn
}

func (e *BaseEntity) getDialect() *Dialect {
	return e.getMetadata().dialect
}

func (e *BaseEntity) getDB() *DB {
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

func (b *BaseEntity) HasMany(out interface{}, config HasManyConfig) Relation {
	return Relation{
		output: out,
		c:      config,
	}
}

func (e *BaseEntity) getFields() []*Field {
	var fields []*Field
	if e.fields != nil {
		fields = e.fields()
	} else {
		fields = e.getMetadata().Fields
	}
	return fields
}
func (e *BaseEntity) QueryContext(ctx context.Context, q string, args ...interface{}) (*sql.Rows, error) {
	return e.getConnection().QueryContext(ctx, q, args...)
}

func (b *BaseEntity) BelongsTo(output IsEntity, config BelongsToConfig) Relation {
	return Relation{
		typ:    relationTypeBelongsTo,
		output: output,
		c:      config,
	}
}

func (b *BaseEntity) ManyToMany(output []IsEntity, config ManyToManyConfig) Relation {
	return Relation{
		typ:    relationTypeMany2Many,
		output: output,
		c:      config,
	}
}
func (e *BaseEntity) BindContext(ctx context.Context, out interface{}, q string, args ...interface{}) error {
	rows, err := e.QueryContext(ctx, q, args...)
	if err != nil {
		return err
	}
	return e.getMetadata().Bind(rows, out)
}
func (e *BaseEntity) ExecContext() {}
