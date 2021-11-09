package qb

import (
	"fmt"
	"strings"
)

type M = map[string]interface{}

type UpdateStmt struct {
	table string
	where string
	set   M
	args  []interface{}
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

type KV map[string]interface{}

func (u *UpdateStmt) Set(values M) *UpdateStmt {
	u.set = values
	return u
}

func (u *UpdateStmt) Build() (string, []interface{}, error) {
	var pairs []string
	for k, v := range u.set {
		pairs = append(pairs, fmt.Sprintf("%s=%s", k, fmt.Sprint(v)))
	}
	return fmt.Sprintf("UPDATE %s WHERE %s SET %s", u.table, u.where, strings.Join(pairs, ",")), u.args, nil
}

func (u *UpdateStmt) Table(table string) *UpdateStmt {
	u.table = table
	return u
}
func NewUpdate() *UpdateStmt {
	return &UpdateStmt{}
}
