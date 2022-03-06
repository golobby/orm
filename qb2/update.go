package qb2

import (
	"fmt"
	"strings"
)

type updateTuple struct {
	Key   string
	Value interface{}
}

type Update struct {
	Dialect *Dialect
	Table   string
	Set     []updateTuple
	Where   *Where
}

func pop(phs *[]string) string {
	top := (*phs)[len(*phs)-1]
	*phs = (*phs)[:len(*phs)-1]
	return top
}

func (u Update) kvString() string {
	phs := u.Dialect.PlaceHolderGenerator(len(u.Set))
	var sets []string
	for _, pair := range u.Set {
		sets = append(sets, fmt.Sprintf("%s=%s", pair.Key, pop(&phs)))
	}
	return strings.Join(sets, ",")
}

func (u Update) args() []interface{} {
	var values []interface{}
	for _, pair := range u.Set {
		values = append(values, pair.Value)
	}
	return values
}

func (u Update) ToSql() (string, []interface{}) {
	base := fmt.Sprintf("UPDATE %s SET %s", u.Table, u.kvString())
	args := u.args()
	if u.Where != nil {
		u.Where.Dialect = u.Dialect
		where, whereArgs := u.Where.ToSql()
		args = append(args, whereArgs...)
		base += " WHERE " + where
	}
	return base, args
}
