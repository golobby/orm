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
	cols                  []string
	args                  []interface{}
	placeholdersGenerator func(n int) string
}

func NewInsert() *InsertStmt {
	return &InsertStmt{}
}
func (i *InsertStmt) Repository(repository *Repository) *InsertStmt {
	i.repository = repository
	i.table = repository.metadata.Table
	return i
}
func (i *InsertStmt) Into(cols ...string) *InsertStmt {
	i.cols = append(i.cols, cols...)
	return i
}
func (i *InsertStmt) WithArgs(args ...interface{}) *InsertStmt {
	i.args = append(i.args, args...)
	return i
}
func (i *InsertStmt) PlaceHolderGenerator(generator func(n int) string) *InsertStmt {
	i.placeholdersGenerator = generator
	return i
}

func (i *InsertStmt) Exec() (sql.Result, error) {
	s, args, err := i.SQL()
	if err != nil {
		return nil, err
	}
	return i.repository.conn.Exec(s, args...)
}
func (i *InsertStmt) ExecContext(ctx context.Context) (sql.Result, error) {
	s, args, err := i.SQL()
	if err != nil {
		return nil, err
	}
	return i.repository.conn.ExecContext(ctx, s, args...)
}

//SQL returns a query, and list of arguments to query executor
func (i *InsertStmt) SQL() (string, []interface{}, error) {
	if i.placeholdersGenerator == nil {
		i.placeholdersGenerator = PlaceHolderGenerators.Postgres
	}

	return fmt.Sprintf(`INSERT INTO %s (%s) VALUES (%s)`, i.table, strings.Join(i.cols, ","), i.placeholdersGenerator(len(i.args))), i.args, nil
}
func (i *InsertStmt) Table(t string) *InsertStmt {
	i.table = t
	return i
}
