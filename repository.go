package orm

import (
	"database/sql"
	"fmt"

	"github.com/golobby/orm/ds"

	"github.com/golobby/orm/qb"
)

type Repository struct {
	dialect   *Dialect
	conn      *sql.DB
	metadata  *ObjectMetadata
	eagerLoad bool
}

func NewRepository(conn *sql.DB, dialect *Dialect, makeRepositoryFor interface{}) *Repository {
	s := &Repository{
		conn:      conn,
		metadata:  ObjectMetadataFrom(makeRepositoryFor, dialect),
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
	if pkValue != nil {
		ph := s.dialect.PlaceholderChar
		if s.dialect.IncludeIndexInPlaceholder {
			ph = ph + "1"
		}
		builder := qb.NewQuery().
			Select(s.metadata.Columns(true)...).
			From(s.metadata.Table).
			Where(qb.WhereHelpers.Equal(s.pkName(v), ph)).
			WithArgs(pkValue)
		q, args, err = builder.
			Build()
		if err != nil {
			return err
		}

	} else {
		q, args, err = qb.NewQuery().
			From(s.metadata.Table).
			Select(s.metadata.Columns(true)...).
			Where(qb.WhereHelpers.ForKV(s.toMap(v))).Limit(1).Build()
	}
	if err != nil {
		return err
	}
	rows, err := s.conn.Query(q, args...)
	if err != nil {
		return err
	}
	return Bind(rows, v)
}

// Save given object
func (s *Repository) Save(v interface{}) error {
	cols := s.metadata.Columns(false)
	values := s.valuesOf(v, false)
	var phs []string
	if s.dialect.PlaceholderChar == "$" {
		phs = qb.PlaceHolderGenerators.Postgres(len(cols))
	} else {
		phs = qb.PlaceHolderGenerators.MySQL(len(cols))
	}
	q, args, err := qb.NewInsert().
		Table(s.metadata.Table).
		Into(cols...).
		Values(phs...).
		WithArgs(values...).Build()
	if err != nil {
		return err
	}
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
	var kvsWithPh []ds.KV
	var args []interface{}
	for _, kv := range kvs {
		thisPh := s.dialect.PlaceholderChar
		if s.dialect.IncludeIndexInPlaceholder {
			thisPh += fmt.Sprint(counter)
		}
		kvsWithPh = append(kvsWithPh, ds.KV{Key: kv.Key, Value: thisPh})
		args = append(args, kv.Value)
		counter++
	}
	query := qb.WhereHelpers.Equal(s.pkName(v), ph)
	q, args, err := qb.NewUpdate().
		Table(s.metadata.Table).
		Where(query).WithArgs(s.getPkValue(v)).
		Set(kvsWithPh...).WithArgs(args...).
		Build()
	_, err = s.conn.Exec(q, args...)
	return err
}

// Delete the object from database
func (s *Repository) Delete(v interface{}) error {
	ph := s.dialect.PlaceholderChar
	if s.dialect.IncludeIndexInPlaceholder {
		ph = ph + "1"
	}
	query := qb.WhereHelpers.Equal(s.pkName(v), ph)
	q, args, err := qb.NewDelete().
		Table(s.metadata.Table).
		Where(query).
		WithArgs(s.getPkValue(v)).
		Build()
	_, err = s.conn.Exec(q, args...)
	return err
}
