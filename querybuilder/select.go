package querybuilder

import (
	"fmt"
	"strings"
)

type clauseType string

const (
	_ClauseType_Where         = "WHERE"
	_ClauseType_Limit         = "LIMIT"
	_ClauseType_Offset        = "OFFSET"
	_ClauseType_OrderBy       = "ORDER BY"
	_ClauseType_GroupBy       = "GROUP BY"
	_ClauseType_InnerJoin     = "INNER JOIN"
	_ClauseType_LeftJoin      = "LEFT JOIN"
	_ClauseType_RightJoin     = "RIGHT JOIN"
	_ClauseType_FullOuterJoin = "FULL OUTER JOIN"
	_ClauseType_Select        = "SELECT"
	_ClauseType_Having        = "HAVING"
)

type clause struct {
	typ       clauseType
	parts     []string
	delimiter string
}

func (c *clause) String() string {
	if c.delimiter == "" {
		c.delimiter = " "
	}
	return fmt.Sprintf("%s %s", c.typ, strings.Join(c.parts, c.delimiter))
}

type selectClause struct {
	*clause
	distinct bool
}

func (s *selectClause) String() string {
	if s.distinct {
		s.clause.typ += " DISTINCT"
	}
	return s.clause.String()
}

type Select struct {
	table    string
	subQuery *Select
	selected *selectClause
	where    *clause
	orderBy  *clause
	groupBy  *clause
	joins    []*clause
	limit    *clause
	offset   *clause
	having   *clause
	args     []interface{}
}

func (s *Select) WithArgs(args ...interface{}) *Select {
	s.args = append(s.args, args...)
	return s
}

func (q *Select) Having(cond ...string) *Select {
	if q.having == nil {
		q.having = &clause{
			typ:   _ClauseType_Having,
			parts: cond,
		}
		return q
	}

	q.having.parts = append(q.having.parts, "AND")
	q.having.parts = append(q.having.parts, cond...)
	return q
}

func (q *Select) Limit(n int) *Select {
	q.limit = &clause{typ: _ClauseType_Limit, parts: []string{fmt.Sprint(n)}}
	return q
}

func (q *Select) Offset(n int) *Select {
	q.offset = &clause{typ: _ClauseType_Offset, parts: []string{fmt.Sprint(n)}}
	return q
}

func (q *Select) Skip(n int) *Select {
	return q.Offset(n)
}

func (q *Select) Take(n int) *Select {
	return q.Limit(n)
}

func (q *Select) InnerJoin(table string, conds ...string) *Select {
	arg := []string{table, "ON"}
	arg = append(arg, conds...)
	j := &clause{typ: _ClauseType_InnerJoin, parts: arg}
	q.joins = append(q.joins, j)
	return q
}

func (q *Select) RightJoin(table string, conds ...string) *Select {
	arg := []string{table, "ON"}
	arg = append(arg, conds...)
	j := &clause{typ: _ClauseType_RightJoin, parts: arg}
	q.joins = append(q.joins, j)
	return q
}
func (q *Select) LeftJoin(table string, conds ...string) *Select {
	arg := []string{table, "ON"}
	arg = append(arg, conds...)
	j := &clause{typ: _ClauseType_LeftJoin, parts: arg}
	q.joins = append(q.joins, j)
	return q
}
func (q *Select) FullOuterJoin(table string, conds ...string) *Select {
	arg := []string{table, "ON"}
	arg = append(arg, conds...)
	j := &clause{typ: _ClauseType_FullOuterJoin, parts: arg}
	q.joins = append(q.joins, j)
	return q

}

func (q *Select) OrderBy(column, order string) *Select {
	if q.orderBy == nil {
		q.orderBy = &clause{
			typ:       _ClauseType_OrderBy,
			parts:     []string{fmt.Sprintf("%s %s", column, order)},
			delimiter: ", ",
		}
		return q
	}
	q.orderBy.parts = append(q.orderBy.parts, fmt.Sprintf("%s %s", column, order))
	return q
}

func (q *Select) Select(columns ...string) *Select {
	if q.selected == nil {
		q.selected = &selectClause{
			clause: &clause{
				typ:       _ClauseType_Select,
				parts:     columns,
				delimiter: ", ",
			}}
		return q
	}
	q.selected.parts = append(q.selected.parts, columns...)
	return q
}

func (s *Select) Distinct() *Select {
	s.selected.distinct = true
	return s
}

func (q *Select) From(t string) *Select {
	q.table = t
	return q
}

func (q *Select) FromQuery(sub *Select) *Select {
	q.subQuery = sub
	return q
}

func (q *Select) GroupBy(columns ...string) *Select {
	if q.groupBy == nil {
		q.groupBy = &clause{
			typ:       _ClauseType_GroupBy,
			parts:     columns,
			delimiter: ", ",
		}
		return q
	}

	q.groupBy.parts = append(q.groupBy.parts, columns...)
	return q
}
func interfaceSliceToStringAndArgs(is []interface{}) []string {
	var parts []string
	for _, i := range is {
		parts = append(parts, fmt.Sprint(i))
	}
	return parts
}
func (q *Select) Where(parts ...interface{}) *Select {
	if q.where == nil {
		q.where = &clause{
			typ:   _ClauseType_Where,
			parts: interfaceSliceToStringAndArgs(parts),
		}
		return q
	}

	q.where.parts = append(q.where.parts, "AND")
	q.where.parts = append(q.where.parts, interfaceSliceToStringAndArgs(parts)...)
	return q
}

func (q *Select) OrWhere(parts ...interface{}) *Select {
	if q.where == nil {
		return q.Where(parts...)
	}

	q.where.parts = append(q.where.parts, "OR")
	q.where.parts = append(q.where.parts, interfaceSliceToStringAndArgs(parts)...)
	return q
}

func (q *Select) AndWhere(parts ...interface{}) *Select {
	return q.Where(parts...)
}

func (q *Select) Build() (string, []interface{}) {
	sections := []string{}
	// handle select
	if q.selected == nil {
		q.selected = &selectClause{clause: &clause{typ: _ClauseType_Select, parts: []string{"*"}}}
	}

	sections = append(sections, q.selected.String())

	// Handle from TABLE-NAME
	if q.subQuery == nil {
		sections = append(sections, "FROM "+q.table)
	} else {
		subquery, args := q.subQuery.Build()
		q.args = append(args, q.args...)
		sections = append(sections, fmt.Sprintf("FROM (%s)", subquery))
	}
	if q.joins != nil {
		for _, join := range q.joins {
			sections = append(sections, join.String())
		}
	}
	// handle where
	if q.where != nil {
		sections = append(sections, q.where.String())
	}

	if q.orderBy != nil {
		sections = append(sections, q.orderBy.String())
	}

	if q.groupBy != nil {
		sections = append(sections, q.groupBy.String())
	}

	if q.limit != nil {
		sections = append(sections, q.limit.String())
	}

	if q.offset != nil {
		sections = append(sections, q.offset.String())
	}

	if q.having != nil {
		sections = append(sections, q.having.String())
	}

	return strings.Join(sections, " "), q.args
}

type SQL interface {
	Build() (string, []interface{})
}
