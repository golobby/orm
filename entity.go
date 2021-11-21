package orm

import (
	"context"
	"fmt"
	"github.com/gertd/go-pluralize"
	"reflect"
)

type entity struct {
	repo *Repository
	obj  interface{}
}

func (r *Repository) Entity(obj interface{}) *entity {
	return &entity{r, obj}
}

func (e *entity) HasMany(out interface{}) error {
	t := reflect.TypeOf(out)
	v := reflect.ValueOf(out)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}
	if t.Kind() == reflect.Slice {
		t = t.Elem()
	}
	ph := e.repo.dialect.PlaceholderChar
	target := reflect.New(t).Interface()
	repo := NewRepository(e.repo.conn, e.repo.dialect, target)

	var q string
	var args []interface{}
	for idx, rel := range e.repo.relations {
		if e.repo.dialect.IncludeIndexInPlaceholder {
			ph = ph + fmt.Sprint(idx+1)
		}
		if rel.Table == repo.metadata.Table {
			q, args = newSelect().
				From(rel.Table).
				Select(rel.Columns...).
				Where(WhereHelpers.Equal(rel.Lookup, ph)).
				WithArgs(e.repo.getPkValue(e.obj)).
				Build()
		}

	}
	if q == "" {
		return fmt.Errorf("cannot build the query")
	}
	return repo.BindContext(context.Background(), out, q, args...)
}
func (e *entity) ManyToMany(middleType interface{}, out interface{}) error {
	return nil
}
func (e *entity) HasOne(out interface{}) error {
	t := reflect.TypeOf(out)
	v := reflect.ValueOf(out)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}
	target := reflect.New(t).Interface()
	repo := NewRepository(e.repo.conn, e.repo.dialect, target)
	ph := e.repo.dialect.PlaceholderChar

	var q string
	var args []interface{}
	for idx, rel := range e.repo.relations {

		if e.repo.dialect.IncludeIndexInPlaceholder {
			ph = ph + fmt.Sprint(idx+1)
		}
		if rel.Table == repo.metadata.Table {
			q, args = newSelect().
				From(rel.Table).
				Select(rel.Columns...).
				Where(WhereHelpers.Equal(rel.Lookup, ph)).
				WithArgs(e.repo.getPkValue(e.obj)).
				Build()
		}

	}
	if q == "" {
		return fmt.Errorf("cannot build the query")
	}
	return repo.BindContext(context.Background(), out, q, args...)
}

func (e *entity) BelongsTo(out interface{}) error {
	t := reflect.TypeOf(out)
	v := reflect.ValueOf(out)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}
	ph := e.repo.dialect.PlaceholderChar
	if e.repo.dialect.IncludeIndexInPlaceholder {
		ph = ph+"1"
	}
	target := reflect.New(t).Interface()
	repo := NewRepository(e.repo.conn, e.repo.dialect, target)
	ownerTable := tableName(out)
	owner := pluralize.NewClient().Singular(ownerTable)
	ownerIDName := owner + "_id"
	ownerIDidx := 0
	for idx, field := range e.repo.metadata.Fields {
		if field.Name == ownerIDName {
			ownerIDidx = idx
		}
	}
	ownerID:=e.repo.valuesOf(e.obj, true)[ownerIDidx]
	q, args := newSelect().
		From(ownerTable).
		Where(WhereHelpers.Equal(ownerTable + "." + "id", ph)).
		WithArgs(ownerID).
		Build()
	if q == "" {
		return fmt.Errorf("cannot build the query")
	}
	return repo.BindContext(context.Background(), out, q, args...)
}

func (e *entity) Save() error {
	return e.repo.Save(e.obj)
}
func (e *entity) Update() error {
	return e.repo.Update(e.obj)
}
func (e *entity) Delete() error {
	return e.repo.Delete(e.obj)
}
