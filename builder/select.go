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

type SelectStmt struct {
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

func (q *SelectStmt) Having(cond ...string) *SelectStmt {
	if q.having == "" {
		q.having = strings.Join(cond, " ")
		return q
	}
	q.having = fmt.Sprintf("%s AND %s", q.having, strings.Join(cond, " "))
	return q
}

func (q *SelectStmt) Limit(n int) *SelectStmt {
	q.limit = &cursorClause{typ: "LIMIT", n: n}
	return q
}

func (q *SelectStmt) Offset(n int) *SelectStmt {
	q.offset = &cursorClause{typ: "OFFSET", n: n}
	return q
}

func (q *SelectStmt) Skip(n int) *SelectStmt {
	return q.Offset(n)
}

func (q *SelectStmt) Take(n int) *SelectStmt {
	return q.Limit(n)
}

type joinClause struct {
	// INNER LEFT RIGHT FULL
	joinType string
	conds    string
	table    string
}

func (q *SelectStmt) InnerJoin(table string, conds ...string) *SelectStmt {
	j := &joinClause{table: table, joinType: "INNER", conds: strings.Join(conds, " ")}
	q.joins = append(q.joins, j)
	return q
}

func (q *SelectStmt) RightJoin(table string, conds ...string) *SelectStmt {
	j := &joinClause{table: table, joinType: "RIGHT", conds: strings.Join(conds, " ")}
	q.joins = append(q.joins, j)
	return q

}
func (q *SelectStmt) LeftJoin(table string, conds ...string) *SelectStmt {
	j := &joinClause{table: table, joinType: "LEFT", conds: strings.Join(conds, " ")}
	q.joins = append(q.joins, j)
	return q

}
func (q *SelectStmt) FullOuterJoin(table string, conds ...string) *SelectStmt {
	j := &joinClause{table: table, joinType: "FULL OUTER", conds: strings.Join(conds, " ")}
	q.joins = append(q.joins, j)
	return q

}
func (j *joinClause) String() string {
	return fmt.Sprintf("%s JOIN %s ON %s", j.joinType, j.table, j.conds)
}

type orderbyClause struct {
	parent  *SelectStmt
	columns map[string]string
}

func (q *SelectStmt) OrderBy(column, order string) *SelectStmt {
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

func (q *SelectStmt) Select(columns ...string) *SelectStmt {
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

func (s *SelectStmt) Distinct() *SelectStmt {
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

func (q *SelectStmt) Table(t string) *SelectStmt {
	q.table = t
	return q
}

func (q *SelectStmt) GroupBy(columns ...string) *SelectStmt {
	if q.groupBy == nil {
		q.groupBy = &groupByClause{
			columns: columns,
		}
		return q
	}
	q.groupBy.columns = append(q.groupBy.columns, columns...)

	return q
}

func (q *SelectStmt) Where(parts ...string) *SelectStmt {
	if q.where == "" {
		q.where = fmt.Sprintf("%s", strings.Join(parts, " "))
		return q
	}
	q.where = fmt.Sprintf("%s AND %s", q.where, strings.Join(parts, " "))
	return q
}

func (q *SelectStmt) OrWhere(parts ...string) *SelectStmt {
	q.where = fmt.Sprintf("%s OR %s", q.where, strings.Join(parts, " "))
	return q
}

func (q *SelectStmt) AndWhere(parts ...string) *SelectStmt {
	return q.Where(parts...)
}

func (q *SelectStmt) SQL() (string, error) {
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

func (q *SelectStmt) Exec(db *sql.DB, args ...interface{}) (sql.Result, error) {
	query, err := q.SQL()
	if err != nil {
		return nil, err
	}
	return exec(context.Background(), db, query, args)

}

func (q *SelectStmt) ExecContext(ctx context.Context, db *sql.DB, args ...interface{}) (sql.Result, error) {
	s, err := q.SQL()
	if err != nil {
		return nil, err
	}
	return exec(context.Background(), db, s, args)
}

func (q *SelectStmt) Bind(db *sql.DB, v interface{}, args ...interface{}) error {
	s, err := q.SQL()
	if err != nil {
		return err
	}
	return _bind(context.Background(), db, v, s, args)
}

func (q *SelectStmt) BindContext(ctx context.Context, db *sql.DB, v interface{}, args ...interface{}) error {
	s, err := q.SQL()
	if err != nil {
		return err
	}
	return _bind(ctx, db, v, s, args)
}

func NewQuery() *SelectStmt {
	return &SelectStmt{}
}
