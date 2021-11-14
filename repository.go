package orm

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/golobby/orm/binder"
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
		metadata:  ObjectMetadataFrom(makeRepositoryFor),
		dialect:   dialect,
		eagerLoad: true,
	}
	return s
}

//Fill the struct
func (s *Repository) Fill(v interface{}) error {
	var q string
	var args []interface{}
	var err error
	pkValue := ObjectHelpers.PKValue(v)
	if pkValue != nil {
		ph := s.dialect.PlaceholderChar
		if s.dialect.IncludeIndexInPlaceholder {
			ph = ph + "1"
		}
		cols := s.metadata.Columns()
		builder := qb.NewQuery().
			Select(cols...).
			From(s.metadata.Table).
			Where(qb.WhereHelpers.Equal(ObjectHelpers.PKColumn(v), ph)).
			WithArgs(pkValue)
		q, args, err = builder.
			Build()
		if err != nil {
			return err
		}

	} else {
		q, args, err = qb.NewQuery().From(s.metadata.Table).Select(s.metadata.Columns()...).Where(qb.WhereHelpers.ForKV(ObjectHelpers.ToMap(v))).Limit(1).Build()
	}
	if err != nil {
		return err
	}
	rows, err := s.conn.Query(q, args...)
	if err != nil {
		return err
	}
	return binder.Bind(rows, v)
}

//Save given object
func (s *Repository) Save(v interface{}) error {
	cols, values := ObjectHelpers.InsertColumnsAndValuesOf(v)
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
	ObjectHelpers.SetPK(v, id)
	return nil
}

//Update object in database
func (s *Repository) Update(v interface{}) error {
	ph := s.dialect.PlaceholderChar
	if s.dialect.IncludeIndexInPlaceholder {
		ph = ph + "1"
	}
	counter := 2
	asMap := ObjectHelpers.ToMap(v)
	phMap := map[string]interface{}{}
	args := []interface{}{}
	for k, v := range asMap {
		thisPh := s.dialect.PlaceholderChar
		if s.dialect.IncludeIndexInPlaceholder {
			thisPh += fmt.Sprint(counter)
		}
		phMap[k] = thisPh
		args = append(args, v)
		counter++
	}
	query := qb.WhereHelpers.Equal(ObjectHelpers.PKColumn(v), ph)
	q, args, err := qb.NewUpdate().
		Table(s.metadata.Table).
		Where(query).WithArgs(ObjectHelpers.PKValue(v)).
		Set(phMap).WithArgs(args...).
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
	query := qb.WhereHelpers.Equal(ObjectHelpers.PKColumn(v), ph)
	q, args, err := qb.NewDelete().
		Table(s.metadata.Table).
		Where(query).
		WithArgs(ObjectHelpers.PKValue(v)).
		Build()
	_, err = s.conn.Exec(q, args...)
	return err
}
func (s *Repository) Exec(q qb.SQL) (sql.Result, error) {
	return s.ExecContext(context.Background(), q)
}

func (s *Repository) ExecContext(ctx context.Context, q qb.SQL) (sql.Result, error) {
	query, args, err := q.Build()
	if err != nil {
		return nil, err
	}
	return s.conn.ExecContext(ctx, query, args...)
}

func (s *Repository) Query(q qb.SQL) (*sql.Rows, error) {
	return s.QueryContext(context.Background(), q)

}
func (s *Repository) QueryContext(ctx context.Context, q qb.SQL) (*sql.Rows, error) {
	query, args, err := q.Build()
	if err != nil {
		return nil, err
	}
	return s.conn.QueryContext(ctx, query, args...)
}

func (s *Repository) Bind(q qb.SQL, out interface{}) error {
	return s.BindContext(context.Background(), q, out)
}

func (s *Repository) BindContext(ctx context.Context, q qb.SQL, out interface{}) error {
	query, args, err := q.Build()
	if err != nil {
		return err
	}
	rows, err := s.conn.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}
	return binder.Bind(rows, out)
}
