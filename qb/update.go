package qb

import (
	"fmt"
	"github.com/golobby/orm/ds"
	"strings"
)

type M = map[string]interface{}

type UpdateStmt struct {
	table string
	where string
	set   []ds.KV
	args []interface{}
}

func (q *UpdateStmt) WithArgs(args ...interface{}) *UpdateStmt {
	q.args = append(q.args, args...)
	return q
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

func (u *UpdateStmt) Set(kvs ...ds.KV) *UpdateStmt {
	u.set = append(u.set, kvs...)
	return u
}

func (u *UpdateStmt) Build() (string, []interface{}, error) {
	var pairs []string
	for _, kv := range u.set {
		pairs = append(pairs, fmt.Sprintf("%s=%s", kv.Key, fmt.Sprint(kv.Value)))
	}
	return fmt.Sprintf("UPDATE %s SET %s WHERE %s", u.table, strings.Join(pairs, ","), u.where), u.args, nil
}

func (u *UpdateStmt) Table(table string) *UpdateStmt {
	u.table = table
	return u
}
func NewUpdate() *UpdateStmt {
	return &UpdateStmt{}
}
