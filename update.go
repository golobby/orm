package orm

import (
	"fmt"
	"strings"
)

type M = map[string]interface{}

type updateStmt struct {
	table string
	where string
	set   []keyValue
	args  []interface{}
}

func (q *updateStmt) WithArgs(args ...interface{}) *updateStmt {
	q.args = append(q.args, args...)
	return q
}
func (q *updateStmt) Where(parts ...string) *updateStmt {
	if q.where == "" {
		q.where = fmt.Sprintf("%s", strings.Join(parts, " "))
		return q
	}
	q.where = fmt.Sprintf("%s AND %s", q.where, strings.Join(parts, " "))
	return q
}

func (q *updateStmt) OrWhere(parts ...string) *updateStmt {
	q.where = fmt.Sprintf("%s OR %s", q.where, strings.Join(parts, " "))
	return q
}

func (q *updateStmt) AndWhere(parts ...string) *updateStmt {
	return q.Where(parts...)
}

func (u *updateStmt) Set(key string, value interface{}) *updateStmt {
	u.set = append(u.set, keyValue{
		Key:   key,
		Value: value,
	})
	return u
}

func (u *updateStmt) Build() (string, []interface{}) {
	var pairs []string
	for _, kv := range u.set {
		pairs = append(pairs, fmt.Sprintf("%s=%s", kv.Key, fmt.Sprint(kv.Value)))
	}
	return fmt.Sprintf("UPDATE %s SET %s WHERE %s", u.table, strings.Join(pairs, ","), u.where), u.args
}

func (u *updateStmt) Table(table string) *updateStmt {
	u.table = table
	return u
}
func Update() *updateStmt {
	return &updateStmt{}
}
