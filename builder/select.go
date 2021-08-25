package builder

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

type cursorClause struct {
	typ string
	n   int
}

func (c *cursorClause) String() string {
	return fmt.Sprintf("%s %d", c.typ, c.n)
}

type QueryStmt struct {
	table    string
	selected *selectClause
	where    string
	orderBy  *orderbyClause
	groupBy  *groupByClause
	joins    []*joinClause
	limit    *cursorClause
	offset   *cursorClause
	having   string
}

func (q *QueryStmt) Having(cond ...string) *QueryStmt {
	if q.having == "" {
		q.having = strings.Join(cond, " ")
		return q
	}
	q.having = fmt.Sprintf("%s AND %s", q.having, strings.Join(cond, " "))
	return q
}

func (q *QueryStmt) Limit(n int) *QueryStmt {
	q.limit = &cursorClause{typ: "LIMIT", n: n}
	return q
}

func (q *QueryStmt) Offset(n int) *QueryStmt {
	q.offset = &cursorClause{typ: "OFFSET", n: n}
	return q
}

func (q *QueryStmt) Skip(n int) *QueryStmt {
	return q.Offset(n)
}

func (q *QueryStmt) Take(n int) *QueryStmt {
	return q.Limit(n)
}

type joinClause struct {
	parent *QueryStmt
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

func (j *joinClause) Query() *QueryStmt {
	return j.parent
}

type orderbyClause struct {
	parent  *QueryStmt
	columns map[string]string
}

func (q *QueryStmt) OrderBy(column, order string) *QueryStmt {
	if q.orderBy == nil {
		q.orderBy = &orderbyClause{
			parent:  q,
			columns: map[string]string{column: order},
		}
		return q
	}
	q.orderBy.columns[column] = order
	return q

}

func (s *orderbyClause) String() string {
	pairs := []string{}
	for col, order := range s.columns {
		pairs = append(pairs, fmt.Sprintf("%s %s", col, order))
	}
	return strings.Join(pairs, ", ")
}

type selectClause struct {
	distinct bool
	columns  []string
}

func (q *QueryStmt) Select(columns ...string) *QueryStmt {
	if q.selected == nil {
		q.selected = &selectClause{
			distinct: false,
			columns:  columns,
		}
		return q
	}
	q.selected.columns = append(q.selected.columns, columns...)
	return q
}

func (s *QueryStmt) Distinct() *QueryStmt {
	s.selected.distinct = true
	return s
}

func makeFunctionFormatter(function string) func(string) string {
	return func(column string) string {
		return fmt.Sprintf("%s(%s)", function, column)
	}
}

type functionCall func(string) string

type aggregators struct {
	Min   functionCall
	Max   functionCall
	Count functionCall
	Avg   functionCall
	Sum   functionCall
}

var Aggregators = &aggregators{
	Min:   makeFunctionFormatter("MIN"),
	Max:   makeFunctionFormatter("MAX"),
	Count: makeFunctionFormatter("COUNT"),
	Avg:   makeFunctionFormatter("AVG"),
	Sum:   makeFunctionFormatter("SUM"),
}

func (s *selectClause) String() string {
	output := strings.Join(s.columns, ", ")
	if s.distinct {
		output = "DISTINCT " + output
	}
	return output
}

type groupByClause struct {
	columns []string
}

func (g *groupByClause) String() string {
	output := strings.Join(g.columns, ", ")
	return output
}

func (q *QueryStmt) Table(t string) *QueryStmt {
	q.table = t
	return q
}

func (q *QueryStmt) InnerJoin(table string) *joinClause {
	j := &joinClause{parent: q, table: table, joinType: "INNER"}
	q.joins = append(q.joins, j)
	return j
}

func (q *QueryStmt) RightJoin(table string) *joinClause {
	j := &joinClause{parent: q, table: table, joinType: "RIGHT"}
	q.joins = append(q.joins, j)
	return j

}
func (q *QueryStmt) LeftJoin(table string) *joinClause {
	j := &joinClause{parent: q, table: table, joinType: "LEFT"}
	q.joins = append(q.joins, j)
	return j

}
func (q *QueryStmt) FullOuterJoin(table string) *joinClause {
	j := &joinClause{parent: q, table: table, joinType: "FULL OUTER"}
	q.joins = append(q.joins, j)
	return j

}
func (q *QueryStmt) GroupBy(columns ...string) *QueryStmt {
	if q.groupBy == nil {
		q.groupBy = &groupByClause{
			columns: columns,
		}
		return q
	}
	q.groupBy.columns = append(q.groupBy.columns, columns...)

	return q
}

func (q *QueryStmt) Where(parts ...string) *QueryStmt {
	if q.where == "" {
		q.where = fmt.Sprintf("%s", strings.Join(parts, " "))
		return q
	}
	q.where = fmt.Sprintf("%s AND %s", q.where, strings.Join(parts, " "))
	return q
}

func (q *QueryStmt) OrWhere(parts ...string) *QueryStmt {
	q.where = fmt.Sprintf("%s OR %s", q.where, strings.Join(parts, " "))
	return q
}

func (q *QueryStmt) AndWhere(parts ...string) *QueryStmt {
	return q.Where(parts...)
}

func (q *QueryStmt) SQL() (string, error) {
	sections := []string{}
	// handle select
	if q.selected == nil {
		q.selected = &selectClause{distinct: false, columns: []string{"*"}}
	}

	sections = append(sections, "SELECT", q.selected.String())

	if q.table == "" {
		return "", fmt.Errorf("table name cannot be empty")
	}

	// Handle from TABLE-NAME
	sections = append(sections, "FROM "+q.table)

	// handle where
	if q.where != "" {
		sections = append(sections, "WHERE")
		sections = append(sections, q.where)
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

	if q.limit != nil {
		sections = append(sections, q.limit.String())
	}

	if q.offset != nil {
		sections = append(sections, q.offset.String())
	}

	if q.having != "" {
		sections = append(sections, "HAVING", q.having)
	}

	return strings.Join(sections, " "), nil
}

func (q *QueryStmt) Exec(db *sql.DB, args ...interface{}) (sql.Result, error) {
	query, err := q.SQL()
	if err != nil {
		return nil, err
	}
	return exec(context.Background(), db, query, args)

}

func (q *QueryStmt) ExecContext(ctx context.Context, db *sql.DB, args ...interface{}) (sql.Result, error) {
	s, err := q.SQL()
	if err != nil {
		return nil, err
	}
	return exec(context.Background(), db, s, args)
}

func (q *QueryStmt) Bind(db *sql.DB, v interface{}, args ...interface{}) error {
	s, err := q.SQL()
	if err != nil {
		return err
	}
	return _bind(context.Background(), db, v, s, args)
}

func (q *QueryStmt) BindContext(ctx context.Context, db *sql.DB, v interface{}, args ...interface{}) error {
	s, err := q.SQL()
	if err != nil {
		return err
	}
	return _bind(ctx, db, v, s, args)
}

func NewQuery() *QueryStmt {
	return &QueryStmt{}
}
