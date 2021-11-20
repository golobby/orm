package qb

import (
	"fmt"
	"strings"
)

type DeleteStmt struct {
	table string
	where string
	args  []interface{}
}

func (q *DeleteStmt) WithArgs(args ...interface{}) *DeleteStmt {
	q.args = append(q.args, args...)
	return q
}
func (q *DeleteStmt) Where(parts ...string) *DeleteStmt {
	if q.where == "" {
		q.where = fmt.Sprintf("%s", strings.Join(parts, " "))
		return q
	}
	q.where = fmt.Sprintf("%s AND %s", q.where, strings.Join(parts, " "))
	return q
}

func (q *DeleteStmt) OrWhere(parts ...string) *DeleteStmt {
	q.where = fmt.Sprintf("%s OR %s", q.where, strings.Join(parts, " "))
	return q
}

func (q *DeleteStmt) AndWhere(parts ...string) *DeleteStmt {
	return q.Where(parts...)
}

func (d *DeleteStmt) Build() (string, []interface{}) {
	return fmt.Sprintf("DELETE FROM %s WHERE %s", d.table, d.where), d.args
}

func (d *DeleteStmt) Table(t string) *DeleteStmt {
	d.table = t
	return d
}

func NewDelete() *DeleteStmt {
	return &DeleteStmt{}
}
