package builder

import (
	"context"
	"database/sql"
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
	arg       []string
	delimiter string
}

func (c *Clause) String() string {
	if c.delimiter == "" {
		c.delimiter = " "
	}
	return fmt.Sprintf("%s %s", c.typ, strings.Join(c.arg, c.delimiter))
}

type SelectStmt struct {
	table    string
	selected *Clause
	where    *Clause
	orderBy  *Clause
	groupBy  *Clause
	joins    []*Clause
	limit    *Clause
	offset   *Clause
	having   *Clause
}

func (q *SelectStmt) Having(cond ...string) *SelectStmt {
	if q.having == nil {
		q.having = &Clause{
			typ: ClauseType_Having,
			arg: cond,
		}
		return q
	}

	q.having.arg = append(q.having.arg, "AND")
	q.having.arg = append(q.having.arg, cond...)
	return q
}

func (q *SelectStmt) Limit(n int) *SelectStmt {
	q.limit = &Clause{typ: ClauseType_Limit, arg: []string{fmt.Sprint(n)}}
	return q
}

func (q *SelectStmt) Offset(n int) *SelectStmt {
	q.offset = &Clause{typ: ClauseType_Offset, arg: []string{fmt.Sprint(n)}}
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
	j := &Clause{typ: ClauseType_InnerJoin, arg: arg}
	q.joins = append(q.joins, j)
	return q
}

func (q *SelectStmt) RightJoin(table string, conds ...string) *SelectStmt {
	arg := []string{table, "ON"}
	arg = append(arg, conds...)
	j := &Clause{typ: ClauseType_RightJoin, arg: arg}
	q.joins = append(q.joins, j)
	return q
}
func (q *SelectStmt) LeftJoin(table string, conds ...string) *SelectStmt {
	arg := []string{table, "ON"}
	arg = append(arg, conds...)
	j := &Clause{typ: ClauseType_LeftJoin, arg: arg}
	q.joins = append(q.joins, j)
	return q
}
func (q *SelectStmt) FullOuterJoin(table string, conds ...string) *SelectStmt {
	arg := []string{table, "ON"}
	arg = append(arg, conds...)
	j := &Clause{typ: ClauseType_FullOuterJoin, arg: arg}
	q.joins = append(q.joins, j)
	return q

}

func (q *SelectStmt) OrderBy(column, order string) *SelectStmt {
	if q.orderBy == nil {
		q.orderBy = &Clause{
			typ:       ClauseType_OrderBy,
			arg:       []string{fmt.Sprintf("%s %s", column, order)},
			delimiter: ", ",
		}
		return q
	}
	q.orderBy.arg = append(q.orderBy.arg, fmt.Sprintf("%s %s", column, order))
	return q
}

func (q *SelectStmt) Select(columns ...string) *SelectStmt {
	if q.selected == nil {
		q.selected = &Clause{
			typ: ClauseType_Select,
			arg: []string{strings.Join(columns, ", ")},
		}
		return q
	}
	q.selected.arg = append(q.selected.arg, columns...)
	return q
}

func (s *SelectStmt) Distinct() *SelectStmt {
	s.selected.arg = append([]string{"DISTINCT"}, s.selected.arg...)
	return s
}

func (q *SelectStmt) Table(t string) *SelectStmt {
	q.table = t
	return q
}

func (q *SelectStmt) GroupBy(columns ...string) *SelectStmt {
	if q.groupBy == nil {
		q.groupBy = &Clause{
			typ:       ClauseType_GroupBy,
			arg:       columns,
			delimiter: ", ",
		}
		return q
	}

	q.groupBy.arg = append(q.groupBy.arg, columns...)
	return q
}

func (q *SelectStmt) Where(parts ...string) *SelectStmt {
	if q.where == nil {
		q.where = &Clause{
			typ: ClauseType_Where,
			arg: parts,
		}
		return q
	}

	q.where.arg = append(q.where.arg, "AND")
	q.where.arg = append(q.where.arg, parts...)
	return q
}

func (q *SelectStmt) OrWhere(parts ...string) *SelectStmt {
	if q.where == nil {
		return q.Where(parts...)
	}

	q.where.arg = append(q.where.arg, "OR")
	q.where.arg = append(q.where.arg, parts...)
	return q
}

func (q *SelectStmt) AndWhere(parts ...string) *SelectStmt {
	return q.Where(parts...)
}

func (q *SelectStmt) SQL() (string, error) {
	sections := []string{}
	// handle select
	if q.selected == nil {
		q.selected = &Clause{typ: ClauseType_Select, arg: []string{"*"}}
	}

	sections = append(sections, q.selected.String())

	if q.table == "" {
		return "", fmt.Errorf("table name cannot be empty")
	}

	// Handle from TABLE-NAME
	sections = append(sections, "FROM "+q.table)

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

	if q.having != nil {
		sections = append(sections, q.having.String())
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
