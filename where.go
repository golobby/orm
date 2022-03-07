package orm

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

type Cond struct {
	PlaceHolderGenerator func(n int) []string

	Lhs string
	Op  binaryOp
	Rhs interface{}
}

func (b Cond) ToSql() (string, []interface{}) {
	var phs []string
	if b.Op == In {
		phs = b.PlaceHolderGenerator(len(b.Rhs.([]interface{})))
		return fmt.Sprintf("%s IN (%s)", b.Lhs, strings.Join(phs, ",")), b.Rhs.([]interface{})
	} else {
		phs = b.PlaceHolderGenerator(1)
		return fmt.Sprintf("%s %s %s", b.Lhs, b.Op, pop(&phs)), []interface{}{b.Rhs}
	}
}

const (
	nextType_AND = "AND"
	nextType_OR  = "OR"
)

type whereClause struct {
	PlaceHolderGenerator func(n int) []string
	nextTyp              string
	next                 *whereClause
	Cond
}

func (w whereClause) ToSql() (string, []interface{}) {
	w.Cond.PlaceHolderGenerator = w.PlaceHolderGenerator
	base, args := w.Cond.ToSql()
	if w.next == nil {
		return base, args
	}
	if w.next != nil {
		next, nextArgs := w.next.ToSql()
		base += " " + w.nextTyp + " " + next
		args = append(args, nextArgs)
		return base, args
	}

	return base, args
}
