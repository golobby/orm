package orm

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

type UpdateStmt struct {
	repository *Repository
	table      string
	where      string
	set        string
}

func (q *UpdateStmt) Where(parts ...string) *UpdateStmt {
	if q.where == "" {
		q.where = fmt.Sprintf("%s", strings.Join(parts, " "))
		return q
	}
	q.where = fmt.Sprintf("%s AND %s", q.where, strings.Join(parts, " "))
	return q
}

func (q *UpdateStmt) OrWhere(parts ...string) *UpdateStmt {
	q.where = fmt.Sprintf("%s OR %s", q.where, strings.Join(parts, " "))
	return q
}

func (q *UpdateStmt) AndWhere(parts ...string) *UpdateStmt {
	return q.Where(parts...)
}

type KV map[string]interface{}

func (u *UpdateStmt) Set(kv KV) *UpdateStmt {
	pairs := []string{}
	for k, v := range kv {
		pairs = append(pairs, fmt.Sprintf("%s = %s", k, v))
	}

	u.set = strings.Join(pairs, ", ")
	return u
}

func (u *UpdateStmt) SQL() (string, error) {
	return fmt.Sprintf("UPDATE %s WHERE %s SET %s", u.table, u.where, u.set), nil
}

func (d *UpdateStmt) ExecContext(ctx context.Context, args ...interface{}) (sql.Result, error) {
	s, err := d.SQL()
	if err != nil {
		return nil, err
	}
	return exec(context.Background(), d.repository.conn, s, args)
}
func (d *UpdateStmt) Exec(args ...interface{}) (sql.Result, error) {
	query, err := d.SQL()
	if err != nil {
		return nil, err
	}
	return exec(context.Background(), d.repository.conn, query, args)

}

func (u *UpdateStmt) Repository(repository *Repository) *UpdateStmt {
	u.repository = repository
	u.table = repository.metadata.Table
	return u
}
func (u *UpdateStmt) Table(table string) *UpdateStmt {
	u.table = table
	return u
}
func NewUpdate() *UpdateStmt {
	return &UpdateStmt{}
}
