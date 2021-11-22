package orm

import (
	"context"
	"database/sql"
	"fmt"
)

type Repository struct {
	dialect   *dialect
	conn      *sql.DB
	metadata  *objectMetadata
	eagerLoad bool
}

func (r *Repository) Schema() *objectMetadata {
	return r.metadata
}

func NewRepository(conn *sql.DB, dialect *dialect, makeRepositoryFor interface{}) *Repository {
	md := objectMetadataFrom(makeRepositoryFor, dialect)
	s := &Repository{
		conn:      conn,
		metadata:  md,
		dialect:   dialect,
		eagerLoad: true,
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
	builder := newSelect().
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
func (s *Repository) SelectBuilder() *selectStmt {
	return newSelect().From(s.metadata.Table).Select(s.metadata.Columns(true)...)
}
func (s *Repository) InsertBuilder() *insertStmt {
	return newInsert().Table(s.metadata.Table).Into(s.metadata.Columns(true)...)
}
func (s *Repository) UpdateBuilder() *updateStmt {
	return newUpdate().Table(s.metadata.Table)
}
func (s *Repository) DeleteBuilder() *deleteStmt {
	return newDelete().Table(s.metadata.Table)
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
	q, args := newInsert().
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
	var kvsWithPh []keyValue
	var args []interface{}
	for _, kv := range kvs {
		thisPh := s.dialect.PlaceholderChar
		if s.dialect.IncludeIndexInPlaceholder {
			thisPh += fmt.Sprint(counter)
		}
		kvsWithPh = append(kvsWithPh, keyValue{Key: kv.Key, Value: thisPh})
		args = append(args, kv.Value)
		counter++
	}
	query := WhereHelpers.Equal(s.metadata.pkName(), ph)
	q, args := newUpdate().
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
	q, args := newDelete().
		Table(s.metadata.Table).
		Where(query).
		WithArgs(s.getPkValue(v)).
		Build()
	_, err := s.conn.Exec(q, args...)
	return err
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
