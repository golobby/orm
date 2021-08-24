package builder

import (
	"fmt"
	"strings"
)

type joinClause struct {
	parent *query
	// INNER LEFT RIGHT FULL
	joinType string

	conds string

	table string
}

func (j *joinClause) On(parts ...string) *joinClause {
	j.conds = strings.Join(parts, " ")
	return j
}

func (j *joinClause) String() string {
	return fmt.Sprintf("%s JOIN %s ON %s", j.joinType, j.table, j.conds)
}

func (j *joinClause) Query() *query {
	return j.parent
}
