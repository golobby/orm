package qm

import (
	"github.com/golobby/orm/internal/qb"
)

type QM interface {
	Modify(p *qb.Select)
}

type table struct {
	name string
}

func Table(name string) QM {
	return &table{name: name}
}

func (t table) Modify(p *qb.Select) {
	p.Table(t.name)
}

type columns struct {
	cols []string
}

func Select(cols ...string) QM {
	return &columns{cols: cols}
}

func (c columns) Modify(p *qb.Select) {
	p.Select(c.cols...)
}

type where struct {
	parts []interface{}
}

func Where(parts ...interface{}) QM {
	return &where{parts: parts}
}

func (w where) Modify(p *qb.Select) {
	p.Where(w.parts...)
}

type orderBy struct {
	column string
	order  string
}

func (o orderBy) Modify(p *qb.Select) {
	p.OrderBy(o.order, o.column)
}

func Order(column string, order string) QM {
	return &orderBy{column: column, order: order}
}
