package builder

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

type query struct {
	table     string
	projected *selectClause
	filters   string
	orderBy   *orderbyClause
	groupBy   *groupByClause
	joins     []*joinClause
}

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

type selectClause struct {
	parent   *query
	distinct bool
	columns  []string
}

func makeFunctionFormatter(function string) func(string) string {
	return func(column string) string {
		return fmt.Sprintf("%s(%s)", function, column)
	}
}

type functionCall func(string) string

type selectHelpers struct {
	Min   functionCall
	Max   functionCall
	Count functionCall
	Avg   functionCall
	Sum   functionCall
}

var SelectHelpers = &selectHelpers{
	Min:   makeFunctionFormatter("MIN"),
	Max:   makeFunctionFormatter("MAX"),
	Count: makeFunctionFormatter("COUNT"),
	Avg:   makeFunctionFormatter("AVG"),
	Sum:   makeFunctionFormatter("SUM"),
}

func (s *selectClause) Distinct() *selectClause {
	s.distinct = true
	return s
}

func (s *selectClause) Query() *query {
	return s.parent
}

func (s *selectClause) String() string {
	output := strings.Join(s.columns, ", ")
	if s.distinct {
		output = "DISTINCT " + output
	}
	return output
}

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

func (q *query) Table(t string) *query {
	q.table = t
	return q
}
func (q *query) Select(columns ...string) *selectClause {
	q.projected = &selectClause{
		parent:   q,
		distinct: false,
		columns:  columns,
	}
	return q.projected
}
func (q *query) InnerJoin(table string) *joinClause {
	j := &joinClause{parent: q, table: table, joinType: "INNER"}
	q.joins = append(q.joins, j)
	return j
}

func (q *query) RightJoin(table string) *joinClause {
	j := &joinClause{parent: q, table: table, joinType: "RIGHT"}
	q.joins = append(q.joins, j)
	return j

}
func (q *query) LeftJoin(table string) *joinClause {
	j := &joinClause{parent: q, table: table, joinType: "LEFT"}
	q.joins = append(q.joins, j)
	return j

}
func (q *query) FullOuterJoin(table string) *joinClause {
	j := &joinClause{parent: q, table: table, joinType: "FULL OUTER"}
	q.joins = append(q.joins, j)
	return j

}
func (q *query) GroupBy(columns ...string) *groupByClause {
	q.groupBy = &groupByClause{
		parent:  q,
		columns: columns,
	}
	return q.groupBy
}

func (q *query) Where(parts ...string) *query {
	q.filters = strings.Join(parts, " ")
	return q
}

func (q *query) OrderBy(columns ...string) *orderbyClause {
	q.orderBy = &orderbyClause{
		parent:  q,
		columns: columns,
		desc:    false,
	}

	return q.orderBy
}

func (q *query) SQL() (string, error) {
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
	if q.filters != "" {
		sections = append(sections, "WHERE")
		sections = append(sections, q.filters)
	}

	if q.orderBy != nil {
		sections = append(sections, "ORDER BY", q.orderBy.String())
	}

	if q.groupBy != nil {
		sections = append(sections, "GROUP BY", q.groupBy.String())
	}

	if q.joins != nil {
		for _, join := range q.joins {
			sections = append(sections, join.String())
		}
	}

	return strings.Join(sections, " "), nil
}

func (q *query) Exec(db *sql.DB, args ...interface{}) (sql.Result, error) {
	query, err := q.SQL()
	if err != nil {
		return nil, err
	}
	return exec(context.Background(), db, query, args)

}

func (q *query) ExecContext(ctx context.Context, db *sql.DB, args ...interface{}) (sql.Result, error) {
	s, err := q.SQL()
	if err != nil {
		return nil, err
	}
	return exec(context.Background(), db, s, args)
}

func (q *query) Bind(db *sql.DB, v interface{}, args ...interface{}) error {
	s, err := q.SQL()
	if err != nil {
		return err
	}
	return _bind(context.Background(), db, v, s, args)
}

func (q *query) BindContext(ctx context.Context, db *sql.DB, v interface{}, args ...interface{}) error {
	s, err := q.SQL()
	if err != nil {
		return err
	}
	return _bind(ctx, db, v, s, args)
}

func NewQuery() *query {
	return &query{}
}
