package query

import (
	"fmt"
	"strings"
)

type whereClause struct {
	parent *Query
	conds  []string
}

func Like(column string, pattern string) string {
	return fmt.Sprintf("%s LIKE %s", column, pattern)
}

func In(column string, values ...string) string {
	return fmt.Sprintf("%s IN (%s)", column, strings.Join(values, ", "))
}

func Between(column string, lower string, higher string) string {
	return fmt.Sprintf("%s BETWEEN %s AND %s", column, lower, higher)
}

func Not(cond ...string) string {
	return fmt.Sprintf("NOT %s", strings.Join(cond, " "))
}

func (w *whereClause) And(parts ...string) *whereClause {
	w.conds = append(w.conds, "AND "+strings.Join(parts, " "))
	return w
}
func (w *whereClause) Query() *Query {
	return w.parent
}
func (w *whereClause) Or(parts ...string) *whereClause {
	w.conds = append(w.conds, "OR "+strings.Join(parts, " "))
	return w
}
func (w *whereClause) Not(parts ...string) *whereClause {
	w.conds = append(w.conds, "NOT "+strings.Join(parts, " "))
	return w
}

func (w *whereClause) String() string {
	return strings.Join(w.conds, " ")
}
