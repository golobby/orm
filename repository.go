package orm

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/golobby/orm/qb"
)

type Repository struct {
	dialect   *Dialect
	conn      *sql.DB
	metadata  *ObjectMetadata
	relations []*RelationMetadata
	eagerLoad bool
}

func (r *Repository) Schema() *ObjectMetadata {
	return r.metadata
}

func NewRepository(conn *sql.DB, dialect *Dialect, makeRepositoryFor interface{}) *Repository {
	md := ObjectMetadataFrom(makeRepositoryFor, dialect)
	s := &Repository{
		conn:      conn,
		metadata:  md,
		dialect:   dialect,
		eagerLoad: true,
		relations: md.relationsOf(),
	}
	return s
}

// Fill the struct
func (s *Repository) Fill(v interface{}) error {
	var q string
	var args []interface{}
	var err error
	pkValue := s.getPkValue(v)
	ph := s.dialect.PlaceholderChar
	if s.dialect.IncludeIndexInPlaceholder {
		ph = ph + "1"
	}
	builder := qb.newSelect().
		Select(s.metadata.Columns(true)...).
		From(s.metadata.Table).
		Where(WhereHelpers.Equal(s.metadata.pkName(), ph)).
		WithArgs(pkValue)
	q, args = builder.
		Build()
	rows, err := s.conn.Query(q, args...)
	if err != nil {
		return err
	}
	return s.metadata.Bind(rows, v)
}
func (s *Repository) SelectBuilder() *SelectStmt {
	return qb.newSelect().From(s.metadata.Table).Select(s.metadata.Columns(true)...)
}

// Save given object
func (s *Repository) Save(v interface{}) error {
	cols := s.metadata.Columns(false)
	values := s.valuesOf(v, false)
	var phs []string
	if s.dialect.PlaceholderChar == "$" {
		phs = PlaceHolderGenerators.Postgres(len(cols))
	} else {
		phs = PlaceHolderGenerators.MySQL(len(cols))
	}
	q, args := qb.newInsert().
		Table(s.metadata.Table).
		Into(cols...).
		Values(phs...).
		WithArgs(values...).Build()

	res, err := s.conn.Exec(q, args...)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	s.setPkValue(v, id)
	return nil
}

// Update object in database
func (s *Repository) Update(v interface{}) error {
	ph := s.dialect.PlaceholderChar
	if s.dialect.IncludeIndexInPlaceholder {
		ph = ph + "1"
	}
	counter := 2
	kvs := s.toMap(v)
	var kvsWithPh []KV
	var args []interface{}
	for _, kv := range kvs {
		thisPh := s.dialect.PlaceholderChar
		if s.dialect.IncludeIndexInPlaceholder {
			thisPh += fmt.Sprint(counter)
		}
		kvsWithPh = append(kvsWithPh, KV{Key: kv.Key, Value: thisPh})
		args = append(args, kv.Value)
		counter++
	}
	query := WhereHelpers.Equal(s.metadata.pkName(), ph)
	q, args := qb.newUpdate().
		Table(s.metadata.Table).
		Where(query).WithArgs(s.getPkValue(v)).
		Set(kvsWithPh...).WithArgs(args...).
		Build()
	_, err := s.conn.Exec(q, args...)
	return err
}

// Delete the object from database
func (s *Repository) Delete(v interface{}) error {
	ph := s.dialect.PlaceholderChar
	if s.dialect.IncludeIndexInPlaceholder {
		ph = ph + "1"
	}
	query := WhereHelpers.Equal(s.metadata.pkName(), ph)
	q, args := qb.newDelete().
		Table(s.metadata.Table).
		Where(query).
		WithArgs(s.getPkValue(v)).
		Build()
	_, err := s.conn.Exec(q, args...)
	return err
}

func (s *Repository) Bind(out interface{}, q string, args ...interface{}) error {
	return s.BindContext(context.Background(), out, q, args...)
}

func (s *Repository) BindContext(ctx context.Context, out interface{}, q string, args ...interface{}) error {
	rows, err := s.conn.QueryContext(ctx, q, args...)
	if err != nil {
		return err
	}
	return s.metadata.Bind(rows, out)
}

func (s *Repository) DB() *sql.DB {
	return s.conn
}
