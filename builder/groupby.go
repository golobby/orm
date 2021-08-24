package builder

import "strings"

type groupByClause struct {
	parent  *query
	columns []string
}

func (g *groupByClause) Query() *query {
	return g.parent
}

func (g *groupByClause) String() string {
	output := strings.Join(g.columns, ", ")
	return output
}
