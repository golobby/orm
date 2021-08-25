package builder

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

type UpdateStmt struct {
	table string
	where string
	set   string
}

func (u *UpdateStmt) Where(parts ...string) *UpdateStmt {
	w := strings.Join(parts, " ")
	u.where = w
	return u
}

type KV map[string]string

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

func (d *UpdateStmt) ExecContext(ctx context.Context, db *sql.DB, args ...interface{}) (sql.Result, error) {
	s, err := d.SQL()
	if err != nil {
		return nil, err
	}
	return exec(context.Background(), db, s, args)
}
func (d *UpdateStmt) Exec(db *sql.DB, args ...interface{}) (sql.Result, error) {
	query, err := d.SQL()
	if err != nil {
		return nil, err
	}
	return exec(context.Background(), db, query, args)

}
func NewUpdate(table string) *UpdateStmt {
	return &UpdateStmt{table: table}
}
