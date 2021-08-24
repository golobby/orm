package builder

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

type insertStmt struct {
	table                 string
	columns               []string
	values                []interface{}
	placeholdersGenerator func(n int) string
}

func NewInsert(table string) *insertStmt {
	return &insertStmt{table: table}
}

func (i *insertStmt) Into(columns ...string) *insertStmt {
	i.columns = columns
	return i
}

func (i *insertStmt) Values(values ...interface{}) *insertStmt {
	i.values = values
	return i
}

func (i *insertStmt) PlaceHolderGenerator(generator func(n int) string) *insertStmt {
	i.placeholdersGenerator = generator
	return i
}

func (i *insertStmt) Exec(db *sql.DB) (sql.Result, error) {
	s, err := i.SQL()
	if err != nil {
		return nil, err
	}
	return db.Exec(s, i.values...)
}
func (i *insertStmt) ExecContext(ctx context.Context, db *sql.DB) (sql.Result, error) {
	s, err := i.SQL()
	if err != nil {
		return nil, err
	}
	return db.ExecContext(ctx, s, i.values...)

}

func (i *insertStmt) SQL() (string, error) {
	if i.placeholdersGenerator == nil {
		i.placeholdersGenerator = PlaceHolderGenerators.Postgres
	}
	return fmt.Sprintf(`INSERT INTO %s (%s) VALUES (%s)`, i.table, strings.Join(i.columns, ", "), i.placeholdersGenerator(len(i.values))), nil
}
