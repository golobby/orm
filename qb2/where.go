package qb2

import "fmt"

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
	Lhs string
	Op  binaryOp
	Rhs interface{}
}

func (b BinaryOp) String() string {
	switch b.Rhs.(type) {
	case string:
		return fmt.Sprintf("%s %s '%s'", b.Lhs, b.Op, b.Rhs.(string))
	default:
		return fmt.Sprintf("%s %s %s", b.Lhs, b.Op, fmt.Sprint(b.Rhs))
	}
}

type Where struct {
	Or  *Where
	And *Where
	BinaryOp
}

func (w Where) String() string {
	base := w.BinaryOp.String()
	if w.And != nil && w.Or != nil {
		return base
	}
	if w.And != nil {
		return fmt.Sprintf("%s AND %s", base, w.And.String())
	}

	if w.Or != nil {
		return fmt.Sprintf("%s OR %s", base, w.Or.String())
	}
	return base
}
