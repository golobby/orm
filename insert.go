package orm

import (
	"fmt"
	"strings"
)

type insertStmt struct {
	PlaceHolderGenerator func(n int) []string
	Table                string
	Columns              []string
	Values               [][]interface{}
}

func (i insertStmt) flatValues() []interface{} {
	var values []interface{}
	for _, row := range i.Values {
		values = append(values, row...)
	}
	return values
}

func (i insertStmt) getValuesStr() string {
	phs := i.PlaceHolderGenerator(len(i.Values) * len(i.Values[0]))

	var output []string
	for _, valueRow := range i.Values {
		output = append(output, fmt.Sprintf("(%s)", strings.Join(phs[:len(valueRow)], ",")))
		phs = phs[len(valueRow):]
	}
	return strings.Join(output, ",")
}

func (i insertStmt) ToSql() (string, []interface{}) {
	base := fmt.Sprintf("INSERT INTO %s (%s) VALUES %s",
		i.Table,
		strings.Join(i.Columns, ","),
		i.getValuesStr(),
	)
	return base, i.flatValues()
}
