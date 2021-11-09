package orm

import (
	"context"
	"database/sql"
	"fmt"
)

type Repository struct {
	dialect  *Dialect
	conn     *sql.DB
	metadata *ObjectMetadata
}

func NewRepository(conn *sql.DB, dialect *Dialect, makeRepositoryFor interface{}) *Repository {
	s := &Repository{
		conn:     conn,
		metadata: ObjectMetadataFrom(makeRepositoryFor),
		dialect:  dialect,
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
		q, args, err = NewQuery().Select(s.metadata.Columns()...).Table(s.metadata.Table).Where(WhereHelpers.Equal(ObjectHelpers.PKColumn(v), ph)).WithArgs(pkValue).SQL()
		if err != nil {
			return err
		}

	} else {
		q, args, err = NewQuery().Table(s.metadata.Table).Select(s.metadata.Columns()...).Where(WhereHelpers.ForKV(ObjectHelpers.ToMap(v))).Limit(1).SQL()
	}
	if err != nil {
		return err
	}
	return _bind(context.Background(), s.conn, v, q, args...)
}

//Save given object
func (s *Repository) Save(v interface{}) error {
	cols, values := ObjectHelpers.InsertColumnsAndValuesOf(v)
	var phs []string
	if s.dialect.PlaceholderChar == "$" {
		phs = postgresPlaceholder(len(cols))
	} else {
		phs = mySQLPlaceHolder(len(cols))
	}
	q, args, err := NewInsert().
		Table(s.metadata.Table).
		Into(cols...).
		Values(phs...).
		WithArgs(values...).SQL()
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
	query := WhereHelpers.Equal(ObjectHelpers.PKColumn(v), ph)
	q, args, err := NewUpdate().
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
	query := WhereHelpers.Equal(ObjectHelpers.PKColumn(v), ph)
	q, args, err := NewDelete().
		Table(s.metadata.Table).
		Where(query).
		WithArgs(ObjectHelpers.PKValue(v)).
		SQL()
	_, err = s.conn.Exec(q, args...)
	return err
}
