package builder

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

type updateStmt struct {
	table string
	where string
	set   string
}

func (u *updateStmt) Where(parts ...string) *updateStmt {
	w := strings.Join(parts, " ")
	u.where = w
	return u
}

type KV map[string]string

func (u *updateStmt) Set(kv KV) *updateStmt {
	pairs := []string{}
	for k, v := range kv {
		pairs = append(pairs, fmt.Sprintf("%s = %s", k, v))
	}

	u.set = strings.Join(pairs, ", ")
	return u
}

func (u *updateStmt) SQL() (string, error) {
	return fmt.Sprintf("UPDATE %s WHERE %s SET %s", u.table, u.where, u.set), nil
}

func (d *updateStmt) ExecContext(ctx context.Context, db *sql.DB, args ...interface{}) (sql.Result, error) {
	s, err := d.SQL()
	if err != nil {
		return nil, err
	}
	return exec(context.Background(), db, s, args)
}
func (d *updateStmt) Exec(db *sql.DB, args ...interface{}) (sql.Result, error) {
	query, err := d.SQL()
	if err != nil {
		return nil, err
	}
	return exec(context.Background(), db, query, args)

}
func NewUpdate(table string) *updateStmt {
	return &updateStmt{table: table}
}
