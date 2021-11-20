package orm

import (
	"fmt"
	"reflect"

	"github.com/golobby/orm/qb"
)

type Entity struct {
	repo *Repository
	obj  interface{}
}

func (r *Repository) Entity(obj interface{}) *Entity {
	return &Entity{r, obj}
}

func (e *Entity) HasMany(out interface{}) error {
	t := reflect.TypeOf(out)
	v := reflect.ValueOf(out)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}
	if t.Kind() == reflect.Slice {
		t = t.Elem()
	}
	target := reflect.New(t).Interface()
	repo := NewRepository(e.repo.conn, e.repo.dialect, target)

	var q string
	var args []interface{}
	for _, field := range e.repo.metadata.Fields {
		if !field.IsRel {
			continue
		}
		if field.RelationMetadata.Table == repo.metadata.Table {
			q, args = qb.NewSelect().
				From(field.RelationMetadata.Table).
				Select(field.RelationMetadata.objectMetadata.Columns(true)...).
				Where(qb.WhereHelpers.Equal(field.RelationMetadata.LeftColumn, field.RelationMetadata.RightColumn)).
				WithArgs(e.repo.getPkValue(e.obj)).
				Build()
		}
	}
	if q == "" {
		return fmt.Errorf("cannot build the query")
	}
	return repo.Bind(out, q, args...)
}
func (e *Entity) HasOne(out interface{}) error {
	t := reflect.TypeOf(out)
	v := reflect.ValueOf(out)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}
	target := reflect.New(t).Interface()
	repo := NewRepository(e.repo.conn, e.repo.dialect, target)

	var q string
	var args []interface{}
	for _, field := range e.repo.metadata.Fields {
		if !field.IsRel {
			continue
		}
		if field.RelationMetadata.Table == repo.metadata.Table {
			q, args = qb.NewSelect().
				From(field.RelationMetadata.Table).
				Select(field.RelationMetadata.objectMetadata.Columns(true)...).
				Where(qb.WhereHelpers.Equal(field.RelationMetadata.LeftColumn, field.RelationMetadata.RightColumn)).
				WithArgs(e.repo.getPkValue(e.obj)).
				Build()
		}
	}
	if q == "" {
		return fmt.Errorf("cannot build the query")
	}
	return repo.Bind(out, q, args...)
}