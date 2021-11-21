package orm

import (
	"fmt"
	"reflect"
	"time"

	"github.com/golobby/orm/qb"
)

type Entity struct {
	repo *Repository
	obj  interface{}
}
type BaseModel struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
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
			q, args = qb.newSelect().
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
	ph := e.repo.dialect.PlaceholderChar

	var q string
	var args []interface{}
	for idx, rel := range e.repo.relations {

		if e.repo.dialect.IncludeIndexInPlaceholder {
			ph = ph + fmt.Sprint(idx+1)
		}
		if rel.Table == repo.metadata.Table {
			q, args = qb.newSelect().
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
	return repo.Bind(out, q, args...)
}

func (e *Entity) BelongsTo(out interface{}) error {
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
			q, args = qb.newSelect().
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
	return repo.Bind(out, q, args...)
}

func (e *Entity) Save() error {
	return e.repo.Save(e.obj)
}
func (e *Entity) Update() error {
	return e.repo.Update(e.obj)
}
func (e *Entity) Delete() error {
	return e.repo.Delete(e.obj)
}
