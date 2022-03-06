package qb2

import "fmt"

type Delete struct {
	Dialect *Dialect
	From    string
	Where   *Where
}

func (d Delete) ToSql() (string, []interface{}) {
	base := fmt.Sprintf("DELETE FROM %s", d.From)
	var args []interface{}
	if d.Where != nil {
		d.Where.Dialect = d.Dialect
		where, whereArgs := d.Where.ToSql()
		base += " WHERE " + where
		args = append(args, whereArgs...)
	}
	return base, args
}
