package qb

import (
	"fmt"
	"strings"
)

type ClauseType string

const (
	ClauseType_Where         = "WHERE"
	ClauseType_Limit         = "LIMIT"
	ClauseType_Offset        = "OFFSET"
	ClauseType_OrderBy       = "ORDER BY"
	ClauseType_GroupBy       = "GROUP BY"
	ClauseType_InnerJoin     = "INNER JOIN"
	ClauseType_LeftJoin      = "LEFT JOIN"
	ClauseType_RightJoin     = "RIGHT JOIN"
	ClauseType_FullOuterJoin = "FULL OUTER JOIN"
	ClauseType_Select        = "SELECT"
	ClauseType_Having        = "HAVING"
)

type Clause struct {
	typ       ClauseType
	parts     []string
	delimiter string
}

func (c *Clause) String() string {
	if c.delimiter == "" {
		c.delimiter = " "
	}
	return fmt.Sprintf("%s %s", c.typ, strings.Join(c.parts, c.delimiter))
}

type SelectClause struct {
	*Clause
	distinct bool
}
func (s *SelectClause) String() string {
	if s.distinct {
		s.Clause.typ += " DISTINCT"
	}
	return s.Clause.String()
}

type SelectStmt struct {
	table    string
	subQuery *SelectStmt
	selected *SelectClause
	where    *Clause
	orderBy  *Clause
	groupBy  *Clause
	joins    []*Clause
	limit    *Clause
	offset   *Clause
	having   *Clause
	args     []interface{}
}

func (s *SelectStmt) WithArgs(args ...interface{}) *SelectStmt {
	s.args = append(s.args, args...)
	return s
}

func (q *SelectStmt) Having(cond ...string) *SelectStmt {
	if q.having == nil {
		q.having = &Clause{
			typ:   ClauseType_Having,
			parts: cond,
		}
		return q
	}

	q.having.parts = append(q.having.parts, "AND")
	q.having.parts = append(q.having.parts, cond...)
	return q
}

func (q *SelectStmt) Limit(n int) *SelectStmt {
	q.limit = &Clause{typ: ClauseType_Limit, parts: []string{fmt.Sprint(n)}}
	return q
}

func (q *SelectStmt) Offset(n int) *SelectStmt {
	q.offset = &Clause{typ: ClauseType_Offset, parts: []string{fmt.Sprint(n)}}
	return q
}

func (q *SelectStmt) Skip(n int) *SelectStmt {
	return q.Offset(n)
}

func (q *SelectStmt) Take(n int) *SelectStmt {
	return q.Limit(n)
}

func (q *SelectStmt) InnerJoin(table string, conds ...string) *SelectStmt {
	arg := []string{table, "ON"}
	arg = append(arg, conds...)
	j := &Clause{typ: ClauseType_InnerJoin, parts: arg}
	q.joins = append(q.joins, j)
	return q
}

func (q *SelectStmt) RightJoin(table string, conds ...string) *SelectStmt {
	arg := []string{table, "ON"}
	arg = append(arg, conds...)
	j := &Clause{typ: ClauseType_RightJoin, parts: arg}
	q.joins = append(q.joins, j)
	return q
}
func (q *SelectStmt) LeftJoin(table string, conds ...string) *SelectStmt {
	arg := []string{table, "ON"}
	arg = append(arg, conds...)
	j := &Clause{typ: ClauseType_LeftJoin, parts: arg}
	q.joins = append(q.joins, j)
	return q
}
func (q *SelectStmt) FullOuterJoin(table string, conds ...string) *SelectStmt {
	arg := []string{table, "ON"}
	arg = append(arg, conds...)
	j := &Clause{typ: ClauseType_FullOuterJoin, parts: arg}
	q.joins = append(q.joins, j)
	return q

}

func (q *SelectStmt) OrderBy(column, order string) *SelectStmt {
	if q.orderBy == nil {
		q.orderBy = &Clause{
			typ:       ClauseType_OrderBy,
			parts:     []string{fmt.Sprintf("%s %s", column, order)},
			delimiter: ", ",
		}
		return q
	}
	q.orderBy.parts = append(q.orderBy.parts, fmt.Sprintf("%s %s", column, order))
	return q
}

func (q *SelectStmt) Select(columns ...string) *SelectStmt {
	if q.selected == nil {
		q.selected = &SelectClause{
			Clause: &Clause{
				typ:       ClauseType_Select,
				parts:     columns,
				delimiter: ", ",
			}}
		return q
	}
	q.selected.parts = append(q.selected.parts, columns...)
	return q
}

func (s *SelectStmt) Distinct() *SelectStmt {
	s.selected.distinct = true
	return s
}

func (q *SelectStmt) From(t string) *SelectStmt {
	q.table = t
	return q
}

func (q *SelectStmt) FromQuery(sub *SelectStmt) *SelectStmt {
	q.subQuery = sub
	return q
}

func (q *SelectStmt) GroupBy(columns ...string) *SelectStmt {
	if q.groupBy == nil {
		q.groupBy = &Clause{
			typ:       ClauseType_GroupBy,
			parts:     columns,
			delimiter: ", ",
		}
		return q
	}

	q.groupBy.parts = append(q.groupBy.parts, columns...)
	return q
}

func (q *SelectStmt) Where(parts ...string) *SelectStmt {
	if q.where == nil {
		q.where = &Clause{
			typ:   ClauseType_Where,
			parts: parts,
		}
		return q
	}

	q.where.parts = append(q.where.parts, "AND")
	q.where.parts = append(q.where.parts, parts...)
	return q
}

func (q *SelectStmt) OrWhere(parts ...string) *SelectStmt {
	if q.where == nil {
		return q.Where(parts...)
	}

	q.where.parts = append(q.where.parts, "OR")
	q.where.parts = append(q.where.parts, parts...)
	return q
}

func (q *SelectStmt) AndWhere(parts ...string) *SelectStmt {
	return q.Where(parts...)
}

func (q *SelectStmt) Build() (string, []interface{}, error) {
	sections := []string{}
	// handle select
	if q.selected == nil {
		q.selected = &SelectClause{Clause: &Clause{typ: ClauseType_Select, parts: []string{"*"}}}
	}

	sections = append(sections, q.selected.String())

	if q.table == "" && q.subQuery == nil {
		return "", nil, fmt.Errorf("table name cannot be empty")
	}

	// Handle from TABLE-NAME
	if q.subQuery == nil {
		sections = append(sections, "FROM "+q.table)
	} else {
		subquery, args, err := q.subQuery.Build()
		if err != nil {
			return "", nil, err
		}
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

	return strings.Join(sections, " "), q.args, nil
}

func NewSelect(opts ...func(stmt *SelectStmt)) *SelectStmt {
	s := &SelectStmt{}
	for _, opt := range opts {
		opt(s)
	}
	return s
}
