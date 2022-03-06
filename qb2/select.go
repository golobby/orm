package qb2

import (
	"fmt"
	"strings"
)

type orderByOrder string

const (
	OrderByASC  = "ASC"
	OrderByDesc = "DESC"
)

type OrderBy struct {
	Columns []string
	Order   orderByOrder
}

func (o OrderBy) String() string {
	return fmt.Sprintf("ORDER BY %s %s", strings.Join(o.Columns, ","), o.Order)
}

type GroupBy struct {
	Columns []string
}

func (g GroupBy) String() string {
	return fmt.Sprintf("GROUP BY %s", strings.Join(g.Columns, ","))
}

type joinType string

const (
	JoinTypeInner = "INNER"
	JoinTypeLeft  = "LEFT"
	JoinTypeRight = "RIGHT"
	JoinTypeFull  = "FULL"
	JoinTypeSelf  = "SELF"
)

type JoinOn struct {
	Lhs string
	Rhs string
}

func (j JoinOn) String() string {
	return fmt.Sprintf("%s = %s", j.Lhs, j.Rhs)
}

type Join struct {
	Type  joinType
	Table string
	On    JoinOn
}

func (j Join) String() string {
	return fmt.Sprintf("%s JOIN %s ON %s", j.Type, j.Table, j.On.String())
}

type Limit struct {
	N int
}

func (l Limit) String() string {
	return fmt.Sprintf("LIMIT %d", l.N)
}

type Offset struct {
	N int
}

func (o Offset) String() string {
	return fmt.Sprintf("OFFSET %d", o.N)
}

type Having struct {
	Dialect *Dialect
	Cond    BinaryOp
}

func (h Having) ToSql() (string, []interface{}) {
	h.Cond.Dialect = h.Dialect
	cond, condArgs := h.Cond.ToSql()
	return fmt.Sprintf("HAVING %s", cond), condArgs
}

type Selected struct {
	Columns []string
}

func (s Selected) String() string {
	return fmt.Sprintf("%s", strings.Join(s.Columns, ","))
}

type Select struct {
	Dialect  *Dialect
	Table    string
	SubQuery *Select
	Selected *Selected
	Where    *Where
	OrderBy  *OrderBy
	GroupBy  *GroupBy
	Joins    []*Join
	Limit    *Limit
	Offset   *Offset
	Having   *Having
}

func (s Select) ToSql() (string, []interface{}, error) {
	base := "SELECT"
	var args []interface{}
	//select
	if s.Selected == nil {
		s.Selected = &Selected{
			Columns: []string{"*"},
		}
	}
	base += " " + s.Selected.String()
	// from
	if s.Table == "" && s.SubQuery == nil {
		return "", nil, fmt.Errorf("table name cannot be empty")
	} else if s.Table != "" && s.SubQuery != nil {
		return "", nil, fmt.Errorf("cannot have both table and subquery")
	}
	if s.Table != "" {
		base += " " + "FROM " + s.Table
	}
	if s.SubQuery != nil {
		subQuery, subArgs, err := s.SubQuery.ToSql()
		if err != nil {
			return "", nil, fmt.Errorf("SubQuery: %w", err)
		}
		base += " " + "FROM (" + subQuery + " )"
		args = append(args, subArgs...)
	}
	// Joins
	if s.Joins != nil {
		for _, join := range s.Joins {
			base += " " + join.String()
		}
	}
	// Where
	if s.Where != nil {
		s.Where.Dialect = s.Dialect
		where, whereArgs := s.Where.ToSql()
		base += " WHERE " + where
		args = append(args, whereArgs...)
	}

	// OrderBy
	if s.OrderBy != nil {
		base += " " + s.OrderBy.String()
	}

	// GroupBy
	if s.GroupBy != nil {
		base += " " + s.GroupBy.String()
	}

	// Limit
	if s.Limit != nil {
		base += " " + s.Limit.String()
	}

	// Offset
	if s.Offset != nil {
		base += " " + s.Offset.String()
	}

	// Having
	if s.Having != nil {
		s.Having.Dialect = s.Dialect
		having, havingArgs := s.Having.ToSql()
		base += " " + having
		args = append(args, havingArgs...)
	}

	return base, args, nil
}
