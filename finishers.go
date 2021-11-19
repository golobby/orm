package orm

import (
	"context"
	"database/sql"
	"github.com/golobby/orm/qb"
)

func (s *Repository) injectRepoMetadata(q qb.SQL) {
	switch q.(type) {
	case *qb.SelectStmt:
		selectStmt := q.(*qb.SelectStmt)
		selectStmt.From(s.metadata.Table)
	case *qb.InsertStmt:
		insertStmt := q.(*qb.InsertStmt)
		insertStmt.Table(s.metadata.Table)
	case *qb.UpdateStmt:
		updateStmt := q.(*qb.UpdateStmt)
		updateStmt.Table(s.metadata.Table)
	case *qb.DeleteStmt:
		deleteStmt := q.(*qb.DeleteStmt)
		deleteStmt.Table(s.metadata.Table)
	}
}

func (s *Repository) Exec(q qb.SQL) (sql.Result, error) {
	s.injectRepoMetadata(q)
	return s.ExecContext(context.Background(), q)
}

func (s *Repository) ExecContext(ctx context.Context, q qb.SQL) (sql.Result, error) {
	s.injectRepoMetadata(q)
	query, args, err := q.Build()
	if err != nil {
		return nil, err
	}
	return s.conn.ExecContext(ctx, query, args...)
}

func (s *Repository) Query(q qb.SQL) (*sql.Rows, error) {
	s.injectRepoMetadata(q)
	return s.QueryContext(context.Background(), q)
}

func (s *Repository) QueryContext(ctx context.Context, q qb.SQL) (*sql.Rows, error) {
	s.injectRepoMetadata(q)
	query, args, err := q.Build()
	if err != nil {
		return nil, err
	}
	return s.conn.QueryContext(ctx, query, args...)
}

func (s *Repository) Bind(q qb.SQL, out interface{}) error {
	s.injectRepoMetadata(q)
	return s.BindContext(context.Background(), q, out)
}

func (s *Repository) BindContext(ctx context.Context, q qb.SQL, out interface{}) error {
	s.injectRepoMetadata(q)
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

	builder.Select(field.RelationMetadata.objectMetadata.Columns(true)...)
	builder.LeftJoin(table, qb.WhereHelpers.Equal(field.RelationMetadata.LeftColumn, table+"."+field.RelationMetadata.RightColumn))
	for _, innerField := range field.RelationMetadata.objectMetadata.Fields {
		if innerField.IsRel {
			resolveRelations(builder, innerField)
		}
	}
}

func (s *Repository) FillWithRelations(v interface{}) error {
	var q string
	var args []interface{}
	var err error
	pkValue := getPkValue(v)
	ph := s.dialect.PlaceholderChar
	if s.dialect.IncludeIndexInPlaceholder {
		ph = ph + "1"
	}
	builder := qb.NewQuery().
		Select(s.metadata.Columns(true)...).
		From(s.metadata.Table).
		Where(qb.WhereHelpers.Equal(s.pkName(v), ph)).
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
