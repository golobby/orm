package query

import (
	"fmt"
	"strings"
)

type Query struct {
	table     string
	projected *selectClause
	filters   []*whereClause
	orderBy   *orderbyClause
}

func (q *Query) Table(t string) *Query {
	q.table = t
	return q
}
func (q *Query) Select(columns ...string) *selectClause {
	q.projected = &selectClause{
		parent:   q,
		distinct: false,
		columns:  columns,
	}
	return q.projected
}

func (q *Query) Max(column string) *Query {
	q.projected.columns = append(q.projected.columns, fmt.Sprintf("MAX(%s)", column))
	return q
}

func (q *Query) Min(column string) *Query {
	q.projected.columns = append(q.projected.columns, fmt.Sprintf("MIN(%s)", column))
	return q
}

func (q *Query) Count(column string) *Query {
	q.projected.columns = append(q.projected.columns, fmt.Sprintf("COUNT(%s)", column))
	return q
}

func (q *Query) Avg(column string) *Query {
	q.projected.columns = append(q.projected.columns, fmt.Sprintf("AVG(%s)", column))
	return q
}

func (q *Query) Sum(column string) *Query {
	q.projected.columns = append(q.projected.columns, fmt.Sprintf("SUM(%s)", column))
	return q
}

func (q *Query) Where(parts ...string) *whereClause {
	w := &whereClause{conds: []string{strings.Join(parts, "")}, parent: q}
	q.filters = append(q.filters, w)
	return w
}

func (q *Query) WhereNot(parts ...string) *whereClause {
	w := &whereClause{conds: []string{"NOT " + strings.Join(parts, "")}, parent: q}
	q.filters = append(q.filters, w)
	return w
}

func (q *Query) WhereLike(col string, pattern string) *whereClause {
	w := &whereClause{conds: []string{col + " LIKE ", pattern}, parent: q}
	q.filters = append(q.filters, w)
	return w

}

func (q *Query) OrderBy(columns ...string) *orderbyClause {
	q.orderBy = &orderbyClause{
		parent:  q,
		columns: columns,
		desc:    false,
	}

	return q.orderBy
}

func (q *Query) SQL() (string, error) {
	sections := []string{}
	// handle select
	if q.projected == nil {
		q.projected = &selectClause{parent: q, distinct: false, columns: []string{"*"}}
	}

	sections = append(sections, "SELECT", q.projected.String())

	if q.table == "" {
		return "", fmt.Errorf("table name cannot be empty")
	}

	// Handle from TABLE-NAME
	sections = append(sections, "FROM "+q.table)

	// handle where
	if q.filters != nil {
		sections = append(sections, "WHERE")
		for _, f := range q.filters {
			sections = append(sections, f.String())
		}
	}

	if q.orderBy != nil {
		sections = append(sections, "ORDER BY", q.orderBy.String())
	}

	return strings.Join(sections, " "), nil
}

func New() *Query {
	return &Query{}
}
