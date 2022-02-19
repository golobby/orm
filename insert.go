package orm

import (
	"fmt"
	"strings"
)

type insertStmt struct {
	table  string
	cols   []string
	values [][]string
	args   []interface{}
}

func Insert() *insertStmt {
	return &insertStmt{}
}

func (i *insertStmt) Into(cols ...string) *insertStmt {
	i.cols = append(i.cols, cols...)
	return i
}
func (i *insertStmt) Values(values ...string) *insertStmt {
	i.values = append(i.values, values)
	return i
}
func (i *insertStmt) WithArgs(args ...interface{}) *insertStmt {
	i.args = append(i.args, args...)
	return i
}

//SQL returns a query, and list of arguments to query executor
func (i *insertStmt) Build() (string, []interface{}) {
	var valuesJoined []string
	for _, v := range i.values {
		valuesJoined = append(valuesJoined, fmt.Sprintf("(%s)", strings.Join(v, ",")))
	}
	return fmt.Sprintf(`INSERT INTO %s (%s) VALUES %s`, i.table, strings.Join(i.cols, ","), strings.Join(valuesJoined, ",")), i.args
}
func (i *insertStmt) Table(t string) *insertStmt {
	i.table = t
	return i
}
