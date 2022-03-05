package qm

import (
	"fmt"
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

type groupBy struct {
	cols []string
}

func (g groupBy) Modify(s *qb.Select) {
	s.GroupBy(g.cols...)
}

func GroupBy(columns ...string) QM {
	return &groupBy{
		cols: columns,
	}
}

type in struct {
	column string
	values []interface{}
}

func WhereIn(column string, values ...interface{}) QM {
	return &in{
		column: column,
		values: values,
	}
}

func (i in) Modify(s *qb.Select) {
	var valuesStr []string
	for _, value := range i.values {
		valuesStr = append(valuesStr, fmt.Sprint(value))
	}
	s.Where(qb.WhereHelpers.In(i.column, valuesStr...))
}

type between struct {
	column string
	upper  interface{}
	lower  interface{}
}

func WhereBetween(column string, upper, lower interface{}) QM {
	return &between{
		column: column,
		upper:  upper,
		lower:  lower,
	}
}

func (b between) Modify(s *qb.Select) {
	s.Where(qb.WhereHelpers.Between(b.column, fmt.Sprint(b.lower), fmt.Sprint(b.upper)))
}

type equal struct {
	column string
	value  interface{}
}

func WhereEQ() QM {

}
