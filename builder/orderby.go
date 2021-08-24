package builder

import "strings"

type orderbyClause struct {
	parent  *query
	columns []string
	desc    bool
}

func (s *orderbyClause) Desc() *orderbyClause {
	s.desc = true
	return s
}

func (s *orderbyClause) Query() *query {
	return s.parent
}

func (s *orderbyClause) String() string {
	output := strings.Join(s.columns, ", ")
	if s.desc {
		output = output + " DESC"
	}
	return output
}
