package qb2

import "fmt"

type Delete struct {
	From  string
	Where *Where
}

func (d Delete) String() string {
	base := fmt.Sprintf("DELETE FROM %s", d.From)
	if d.Where != nil {
		base += " WHERE " + d.Where.String()
	}
	return base
}
