package query

import "strings"

type selectClause struct {
	parent   *Query
	distinct bool
	columns  []string
}

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
