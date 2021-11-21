package orm

import (
	"fmt"
	"strings"
)

type InsertStmt struct {
	table  string
	cols   []string
	values [][]string
	args   []interface{}
}

func newInsert() *InsertStmt {
	return &InsertStmt{}
}

func (i *InsertStmt) Into(cols ...string) *InsertStmt {
	i.cols = append(i.cols, cols...)
	return i
}
func (i *InsertStmt) Values(values ...string) *InsertStmt {
	i.values = append(i.values, values)
	return i
}
func (i *InsertStmt) WithArgs(args ...interface{}) *InsertStmt {
	i.args = append(i.args, args...)
	return i
}

//SQL returns a query, and list of arguments to query executor
func (i *InsertStmt) Build() (string, []interface{}) {
	var valuesJoined []string
	for _, v := range i.values {
		valuesJoined = append(valuesJoined, fmt.Sprintf("(%s)", strings.Join(v, ",")))
	}
	return fmt.Sprintf(`INSERT INTO %s (%s) VALUES %s`, i.table, strings.Join(i.cols, ","), strings.Join(valuesJoined, ",")), i.args
}
func (i *InsertStmt) Table(t string) *InsertStmt {
	i.table = t
	return i
}
