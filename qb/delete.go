package qb

import "fmt"

type Delete struct {
	PlaceHolderGenerator func(n int) []string
	From                 string
	Where                *Where
}

func (d Delete) ToSql() (string, []interface{}) {
	base := fmt.Sprintf("DELETE FROM %s", d.From)
	var args []interface{}
	if d.Where != nil {
		d.Where.PlaceHolderGenerator = d.PlaceHolderGenerator
		where, whereArgs := d.Where.ToSql()
		base += " WHERE " + where
		args = append(args, whereArgs...)
	}
	return base, args
}
