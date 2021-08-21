package query

import "strings"

type groupByClause struct {
	parent  *Query
	columns []string
}

func (g *groupByClause) Query() *Query {
	return g.parent
}

func (g *groupByClause) String() string {
	output := strings.Join(g.columns, ", ")
	return output
}
