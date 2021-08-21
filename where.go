package query

import "strings"

type whereClause struct {
	parent *Query
	conds  []string
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

func (w *whereClause) SQL() string {
	return "WHERE " + strings.Join(w.conds, " ")
}
