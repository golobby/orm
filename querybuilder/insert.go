package querybuilder

import (
	"fmt"
	"strings"
)

type Insert struct {
	table  string
	cols   []string
	values [][]string
	args   []interface{}
}

func (i *Insert) Into(cols ...string) *Insert {
	i.cols = append(i.cols, cols...)
	return i
}
func (i *Insert) Values(values ...string) *Insert {
	i.values = append(i.values, values)
	return i
}
func (i *Insert) WithArgs(args ...interface{}) *Insert {
	i.args = append(i.args, args...)
	return i
}

//SQL returns a query, and list of arguments to query executor
func (i *Insert) Build() (string, []interface{}) {
	var valuesJoined []string
	for _, v := range i.values {
		valuesJoined = append(valuesJoined, fmt.Sprintf("(%s)", strings.Join(v, ",")))
	}
	return fmt.Sprintf(`INSERT INTO %s (%s) VALUES %s`, i.table, strings.Join(i.cols, ","), strings.Join(valuesJoined, ",")), i.args
}
func (i *Insert) Table(t string) *Insert {
	i.table = t
	return i
}
