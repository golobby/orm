package querybuilder

import (
	"fmt"
	"strings"
)

type Delete struct {
	table string
	where string
	args  []interface{}
}

func (q *Delete) WithArgs(args ...interface{}) *Delete {
	q.args = append(q.args, args...)
	return q
}
func (q *Delete) Where(parts ...string) *Delete {
	if q.where == "" {
		q.where = fmt.Sprintf("%s", strings.Join(parts, " "))
		return q
	}
	q.where = fmt.Sprintf("%s AND %s", q.where, strings.Join(parts, " "))
	return q
}

func (q *Delete) OrWhere(parts ...string) *Delete {
	q.where = fmt.Sprintf("%s OR %s", q.where, strings.Join(parts, " "))
	return q
}

func (q *Delete) AndWhere(parts ...string) *Delete {
	return q.Where(parts...)
}

func (d *Delete) Build() (string, []interface{}) {
	return fmt.Sprintf("DELETE FROM %s WHERE %s", d.table, d.where), d.args
}

func (d *Delete) Table(t string) *Delete {
	d.table = t
	return d
}
