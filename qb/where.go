package qb

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
	PlaceHolderGenerator func(n int) []string

	Lhs string
	Op  binaryOp
	Rhs interface{}
}

func (b BinaryOp) ToSql() (string, []interface{}) {
	var phs []string
	if b.Op == In {
		phs = b.PlaceHolderGenerator(len(b.Rhs.([]interface{})))
		return fmt.Sprintf("%s IN (%s)", b.Lhs, strings.Join(phs, ",")), b.Rhs.([]interface{})
	} else {
		phs = b.PlaceHolderGenerator(1)
		return fmt.Sprintf("%s %s %s", b.Lhs, b.Op, pop(&phs)), []interface{}{b.Rhs}
	}
}

type Where struct {
	PlaceHolderGenerator func(n int) []string
	Or                   *Where
	And                  *Where
	BinaryOp
}

func (w Where) ToSql() (string, []interface{}) {
	w.BinaryOp.PlaceHolderGenerator = w.PlaceHolderGenerator
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
