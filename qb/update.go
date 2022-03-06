package qb

import (
	"fmt"
	"strings"
)

type updateTuple struct {
	Key   string
	Value interface{}
}

type Update struct {
	PlaceHolderGenerator func(n int) []string
	Table                string
	Set                  [][2]interface{}
	Where                *Where
}

func pop(phs *[]string) string {
	top := (*phs)[len(*phs)-1]
	*phs = (*phs)[:len(*phs)-1]
	return top
}

func (u Update) kvString() string {
	phs := u.PlaceHolderGenerator(len(u.Set))
	var sets []string
	for _, pair := range u.Set {
		sets = append(sets, fmt.Sprintf("%s=%s", pair[0], pop(&phs)))
	}
	return strings.Join(sets, ",")
}

func (u Update) args() []interface{} {
	var values []interface{}
	for _, pair := range u.Set {
		values = append(values, pair[1])
	}
	return values
}

func (u Update) ToSql() (string, []interface{}) {
	base := fmt.Sprintf("UPDATE %s SET %s", u.Table, u.kvString())
	args := u.args()
	if u.Where != nil {
		u.Where.PlaceHolderGenerator = u.PlaceHolderGenerator
		where, whereArgs := u.Where.ToSql()
		args = append(args, whereArgs...)
		base += " WHERE " + where
	}
	return base, args
}
