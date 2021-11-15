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
	pkValue := pkValue(v)
	if pkValue != nil {
		ph := s.dialect.PlaceholderChar
		if s.dialect.IncludeIndexInPlaceholder {
			ph = ph + "1"
		}
		builder := qb.NewQuery().
			Select(s.metadata.Columns()...).
			From(s.metadata.Table).
			Where(qb.WhereHelpers.Equal(pkName(v), ph)).
			WithArgs(pkValue)
		q, args, err = builder.
			Build()
		if err != nil {
			return err
		}

	} else {
		q, args, err = qb.NewQuery().From(s.metadata.Table).Select(s.metadata.Columns()...).Where(qb.WhereHelpers.ForKV(keyValueOf(v))).Limit(1).Build()
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

//Save given object
func (s *Repository) Save(v interface{}) error {
	cols, values := colsAndValsForInsert(v)
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
	setPrimaryKeyFor(v, id)
	return nil
}

//Update object in database
func (s *Repository) Update(v interface{}) error {
	ph := s.dialect.PlaceholderChar
	if s.dialect.IncludeIndexInPlaceholder {
		ph = ph + "1"
	}
	counter := 2
	asMap := keyValueOf(v)
	phMap := map[string]interface{}{}
	var args []interface{}
	for k, v := range asMap {
		thisPh := s.dialect.PlaceholderChar
		if s.dialect.IncludeIndexInPlaceholder {
			thisPh += fmt.Sprint(counter)
		}
		phMap[k] = thisPh
		args = append(args, v)
		counter++
	}
	query := qb.WhereHelpers.Equal(pkName(v), ph)
	q, args, err := qb.NewUpdate().
		Table(s.metadata.Table).
		Where(query).WithArgs(pkValue(v)).
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
	query := qb.WhereHelpers.Equal(pkName(v), ph)
	q, args, err := qb.NewDelete().
		Table(s.metadata.Table).
		Where(query).
		WithArgs(pkValue(v)).
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
	return Bind(rows, out)
}

func resolveRelations(builder *qb.SelectStmt, field *FieldMetadata) {
	table := field.RelationMetadata.Table
	if field.RelationMetadata.Type == RelationTypeHasOne {
		builder.Select(field.RelationMetadata.objectMetadata.Columns()...)
		builder.LeftJoin(table, qb.WhereHelpers.Equal(field.RelationMetadata.LeftColumn, table+"."+field.RelationMetadata.RightColumn))
		for _, innerField := range field.RelationMetadata.objectMetadata.Fields {
			if innerField.IsRel {
				resolveRelations(builder, innerField)
			}
		}
	} else if field.RelationMetadata.Type == RelationTypeHasMany {
		//builder.Select(field.RelationMetadata.objectMetadata.Columns()...)
		//builder.LeftJoin(table, qb.WhereHelpers.Equal(field.RelationMetadata.LeftColumn,))
	}
}

func (s *Repository) FillWithRelations(v interface{}) error {
	var q string
	var args []interface{}
	var err error
	pkValue := pkValue(v)
	ph := s.dialect.PlaceholderChar
	if s.dialect.IncludeIndexInPlaceholder {
		ph = ph + "1"
	}
	builder := qb.NewQuery().
		Select(s.metadata.Columns()...).
		From(s.metadata.Table).
		Where(qb.WhereHelpers.Equal(pkName(v), ph)).
		WithArgs(pkValue)
	for _, field := range s.metadata.Fields {
		if field.RelationMetadata == nil {
			continue
		}
		resolveRelations(builder, field)
	}
	q, args, err = builder.
		Build()
	if err != nil {
		return err
	}
	rows, err := s.conn.Query(q, args...)
	if err != nil {
		return err
	}
	return Bind(rows, v)
}
