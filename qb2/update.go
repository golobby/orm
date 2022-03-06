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
	Table string
	Set   []updateTuple
	Where *Where
}

func (u Update) kvString() string {
	var sets []string
	for _, pair := range u.Set {
		switch pair.Value.(type) {
		case string:
			sets = append(sets, fmt.Sprintf("%s='%s'", pair.Key, pair.Value))
		default:
			sets = append(sets, fmt.Sprintf("%s=%s", pair.Key, fmt.Sprint(pair.Value)))

		}
	}
	return strings.Join(sets, ",")
}

func (u Update) String() string {
	base := fmt.Sprintf("UPDATE %s SET %s", u.Table, u.kvString())
	if u.Where != nil {
		base += " WHERE " + u.Where.String()
	}
	return base
}
