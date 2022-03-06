package qb2

import (
	"fmt"
	"strings"
)

type Insert struct {
	dialect *Dialect
	Into    string
	Columns []string
	Values  [][]interface{}
}

func (i Insert) flatValues() []interface{} {
	var values []interface{}
	for _, row := range i.Values {
		values = append(values, row...)
	}
	return values
}

func (i Insert) getValuesStr() string {
	phs := i.dialect.PlaceHolderGenerator(len(i.Values) * len(i.Values[0]))

	var output []string
	for _, valueRow := range i.Values {
		output = append(output, fmt.Sprintf("(%s)", strings.Join(phs[:len(valueRow)], ",")))
		phs = phs[len(valueRow):]
	}
	return strings.Join(output, ",")
}

func (i Insert) ToSql() (string, []interface{}) {
	base := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s",
		i.Into,
		strings.Join(i.Columns, ","),
		i.getValuesStr(),
	)
	return base, i.flatValues()
}
