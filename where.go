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
		rhs, isInterfaceSlice := b.Rhs.([]interface{})
		if isInterfaceSlice {
			phs = b.PlaceHolderGenerator(len(rhs))
			return fmt.Sprintf("%s IN (%s)", b.Lhs, strings.Join(phs, ",")), rhs
		} else if rawThing, isRaw := b.Rhs.(*raw); isRaw {
			return fmt.Sprintf("%s IN (%s)", b.Lhs, rawThing.sql), rawThing.args
		} else {
			panic("Right hand side of Cond when operator is IN should be either a interface{} slice or *raw")
		}

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
	raw  string
	args []interface{}
}

func (w whereClause) ToSql() (string, []interface{}) {
	w.Cond.PlaceHolderGenerator = w.PlaceHolderGenerator
	var base string
	var args []interface{}
	if w.raw != "" {
		base = w.raw
		args = w.args
	} else {
		base, args = w.Cond.ToSql()
	}
	if w.next == nil {
		return base, args
	}
	if w.next != nil {
		next, nextArgs := w.next.ToSql()
		base += " " + w.nextTyp + " " + next
		args = append(args, nextArgs...)
		return base, args
	}

	return base, args
}
