package orm

import (
	"fmt"
	"strings"
)

type deleteStmt struct {
	table string
	where string
	args  []interface{}
}

func (q *deleteStmt) WithArgs(args ...interface{}) *deleteStmt {
	q.args = append(q.args, args...)
	return q
}
func (q *deleteStmt) Where(parts ...string) *deleteStmt {
	if q.where == "" {
		q.where = fmt.Sprintf("%s", strings.Join(parts, " "))
		return q
	}
	q.where = fmt.Sprintf("%s AND %s", q.where, strings.Join(parts, " "))
	return q
}

func (q *deleteStmt) OrWhere(parts ...string) *deleteStmt {
	q.where = fmt.Sprintf("%s OR %s", q.where, strings.Join(parts, " "))
	return q
}

func (q *deleteStmt) AndWhere(parts ...string) *deleteStmt {
	return q.Where(parts...)
}

func (d *deleteStmt) Build() (string, []interface{}) {
	return fmt.Sprintf("DELETE FROM %s WHERE %s", d.table, d.where), d.args
}

func (d *deleteStmt) Table(t string) *deleteStmt {
	d.table = t
	return d
}

func DeleteStmt() *deleteStmt {
	return &deleteStmt{}
}
