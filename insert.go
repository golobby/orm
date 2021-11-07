package orm

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

type InsertStmt struct {
	repository            *Repository
	table                 string
	columns               []string
	values                []interface{}
	placeholdersGenerator func(n int) string
}

func NewInsert() *InsertStmt {
	return &InsertStmt{}
}
func (i *InsertStmt) Repository(repository *Repository) *InsertStmt {
	i.repository = repository
	i.table = repository.metadata.Table
	i.columns = repository.metadata.Columns(repository.metadata.PrimaryKey)
	return i
}

func (i *InsertStmt) Into(columns ...string) *InsertStmt {
	i.columns = columns
	return i
}

func (i *InsertStmt) Values(values ...interface{}) *InsertStmt {
	i.values = values
	return i
}

func (i *InsertStmt) PlaceHolderGenerator(generator func(n int) string) *InsertStmt {
	i.placeholdersGenerator = generator
	return i
}

func (i *InsertStmt) Exec() (sql.Result, error) {
	s, err := i.SQL()
	if err != nil {
		return nil, err
	}
	return i.repository.conn.Exec(s, i.values...)
}
func (i *InsertStmt) ExecContext(ctx context.Context) (sql.Result, error) {
	s, err := i.SQL()
	if err != nil {
		return nil, err
	}
	return i.repository.conn.ExecContext(ctx, s, i.values...)

}
func (i *InsertStmt) SQL() (string, error) {
	if i.placeholdersGenerator == nil {
		i.placeholdersGenerator = PlaceHolderGenerators.Postgres
	}
	return fmt.Sprintf(`INSERT INTO %s (%s) VALUES (%s)`, i.table, strings.Join(i.columns, ", "), i.placeholdersGenerator(len(i.values))), nil
}
