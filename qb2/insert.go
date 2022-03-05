package qb2

import (
	"fmt"
	"strings"
)

type Insert struct {
	Into    string
	Columns []string
	Values  [][]string
}

func (i Insert) getValuesStr() string {
	var output []string
	for _, valueRow := range i.Values {
		output = append(output, fmt.Sprintf("(%s)", strings.Join(valueRow, ",")))
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
