package qb2

import (
	"fmt"
	"strings"
)

type Insert struct {
	Into    string
	Columns []string
	Values  [][]interface{}
}

func (i Insert) getValuesStr() string {
	var output []string
	for _, valueRow := range i.Values {
		var row []string
		for _, v := range valueRow {
			switch v.(type) {
			case string:
				row = append(row, fmt.Sprintf(`'%s'`, v.(string)))
			default:
				row = append(row, fmt.Sprint(v))
			}
		}
		output = append(output, fmt.Sprintf("(%s)", strings.Join(row, ",")))
	}
	return strings.Join(output, ",")
}

func (i Insert) String() string {
	base := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s",
		i.Into,
		strings.Join(i.Columns, ","),
		i.getValuesStr(),
	)
	return base
}
