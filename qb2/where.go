package qb2

import (
	"fmt"
	"strings"
)

type binaryOp string

const (
	Eq      = "="
	GT      = ">"
	LT      = "<"
	GE      = ">="
	LE      = "<="
	NE      = "!="
	Between = "BETWEEN"
	Like    = "LIKE"
	In      = "IN"
)

type BinaryOp struct {
	Dialect *Dialect
	Lhs     string
	Op      binaryOp
	Rhs     interface{}
}

func (b BinaryOp) ToSql() (string, []interface{}) {
	var phs []string
	if b.Op == In {
		phs = b.Dialect.PlaceHolderGenerator(len(b.Rhs.([]interface{})))
		return fmt.Sprintf("%s IN (%s)", b.Lhs, strings.Join(phs, ",")), b.Rhs.([]interface{})
	} else {
		phs = b.Dialect.PlaceHolderGenerator(1)
		return fmt.Sprintf("%s %s %s", b.Lhs, b.Op, pop(&phs)), []interface{}{b.Rhs}
	}
}

type Where struct {
	Dialect *Dialect
	Or      *Where
	And     *Where
	BinaryOp
}

func (w Where) ToSql() (string, []interface{}) {
	w.BinaryOp.Dialect = w.Dialect
	base, args := w.BinaryOp.ToSql()
	if w.And != nil && w.Or != nil {
		return base, args
	}
	if w.And != nil {
		and, andArgs := w.And.ToSql()
		base += " AND " + and
		args = append(args, andArgs)
		return base, args
	}

	if w.Or != nil {
		or, orArgs := w.And.ToSql()
		base += " OR " + or
		args = append(args, orArgs)
		return base, args
	}
	return base, args
}
