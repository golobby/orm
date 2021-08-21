package query

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/golobby/sql/bind"
)

type Query struct {
	table     string
	projected *selectClause
	filters   []*whereClause
	orderBy   *orderbyClause
	groupBy   *groupByClause
	joins     []*joinClause
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
func (q *Query) InnerJoin(table string) *joinClause {
	j := &joinClause{parent: q, table: table, joinType: "INNER"}
	q.joins = append(q.joins, j)
	return j
}

func (q *Query) RightJoin(table string) *joinClause {
	j := &joinClause{parent: q, table: table, joinType: "RIGHT"}
	q.joins = append(q.joins, j)
	return j

}
func (q *Query) LeftJoin(table string) *joinClause {
	j := &joinClause{parent: q, table: table, joinType: "LEFT"}
	q.joins = append(q.joins, j)
	return j

}
func (q *Query) FullOuterJoin(table string) *joinClause {
	j := &joinClause{parent: q, table: table, joinType: "FULL OUTER"}
	q.joins = append(q.joins, j)
	return j

}
func (q *Query) GroupBy(columns ...string) *groupByClause {
	q.groupBy = &groupByClause{
		parent:  q,
		columns: columns,
	}
	return q.groupBy
}

func (q *Query) Where(parts ...string) *whereClause {
	w := &whereClause{conds: []string{strings.Join(parts, " ")}, parent: q}
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

func (q *Query) Exec(db *sql.DB, args ...interface{}) (sql.Result, error) {
	s, err := q.SQL()
	if err != nil {
		return nil, err
	}
	return db.Exec(s, args...)
}

func (q *Query) ExecContext(ctx context.Context, db *sql.DB, args ...interface{}) (sql.Result, error) {
	s, err := q.SQL()
	if err != nil {
		return nil, err
	}
	return db.ExecContext(ctx, s, args...)
}

func (q *Query) Bind(db *sql.DB, v interface{}, args ...interface{}) error {
	s, err := q.SQL()
	if err != nil {
		return err
	}
	rows, err := db.Query(s, args...)
	if err != nil {
		return err
	}
	return bind.Bind(rows, v)
}

func (q *Query) BindContext(ctx context.Context, db *sql.DB, v interface{}, args ...interface{}) error {
	s, err := q.SQL()
	if err != nil {
		return err
	}
	rows, err := db.Query(s, args...)
	if err != nil {
		return err
	}
	return bind.Bind(rows, v)
}

func New() *Query {
	return &Query{}
}
