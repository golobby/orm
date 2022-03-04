package qb

import (
	"fmt"
	"strings"
)

type M = map[string]interface{}

type Update struct {
	table string
	where string
	set   []keyValue
	args  []interface{}
}

func (q *Update) WithArgs(args ...interface{}) *Update {
	q.args = append(q.args, args...)
	return q
}
func (q *Update) Where(parts ...string) *Update {
	if q.where == "" {
		q.where = fmt.Sprintf("%s", strings.Join(parts, " "))
		return q
	}
	q.where = fmt.Sprintf("%s AND %s", q.where, strings.Join(parts, " "))
	return q
}

func (q *Update) OrWhere(parts ...string) *Update {
	q.where = fmt.Sprintf("%s OR %s", q.where, strings.Join(parts, " "))
	return q
}

func (q *Update) AndWhere(parts ...string) *Update {
	return q.Where(parts...)
}

func (u *Update) Set(key string, value interface{}) *Update {
	u.set = append(u.set, keyValue{
		Key:   key,
		Value: value,
	})
	return u
}

func (u *Update) Build() (string, []interface{}) {
	var pairs []string
	for _, kv := range u.set {
		pairs = append(pairs, fmt.Sprintf("%s=%s", kv.Key, fmt.Sprint(kv.Value)))
	}
	return fmt.Sprintf("UPDATE %s SET %s WHERE %s", u.table, strings.Join(pairs, ","), u.where), u.args
}

func (u *Update) Table(table string) *Update {
	u.table = table
	return u
}
func UpdateStmt() *Update {
	return &Update{}
}
