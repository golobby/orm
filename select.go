package query

import (
	"fmt"
	"strings"
)

type selectClause struct {
	parent   *Query
	distinct bool
	columns  []string
}

func makeFormatter(function string) func(string) string {
	return func(column string) string {
		return fmt.Sprintf("%s(%s)", function, column)
	}
}

var Min = makeFormatter("MIN")
var Max = makeFormatter("MAX")
var Count = makeFormatter("COUNT")
var Avg = makeFormatter("AVG")
var Sum = makeFormatter("SUM")

func (s *selectClause) Distinct() *selectClause {
	s.distinct = true
	return s
}

func (s *selectClause) Query() *Query {
	return s.parent
}

func (s *selectClause) String() string {
	output := strings.Join(s.columns, ", ")
	if s.distinct {
		output = "DISTINCT " + output
	}
	return output
}
