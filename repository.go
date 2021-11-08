package orm

import (
	"database/sql"
	"fmt"
)

type Repository struct {
	conn     *sql.DB
	metadata *ObjectMetadata
}

func NewRepository(conn *sql.DB, makeRepositoryFor interface{}) *Repository {
	s := &Repository{
		conn:     conn,
		metadata: ObjectMetadataFrom(makeRepositoryFor),
	}
	return s
}

//Fill the struct
func (s *Repository) Fill(v interface{}) error {
	pkValue := ObjectHelpers.PKValue(v)
	if pkValue != nil {
		return NewQueryOnRepository(s).WherePK(pkValue).Bind(v)
	}
	kvs := ObjectHelpers.ToMap(v)
	return NewQueryOnRepository(s).Where(WhereHelpers.ForKV(kvs)).Bind(v)
}

//Save given object
func (s *Repository) Save(v interface{}) error {
	cols, values := ObjectHelpers.InsertColumnsAndValuesOf(v)
	res, err := NewInsert().
		Repository(s).
		Into(cols...).
		WithArgs(values...).
		Exec()
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
	query := WhereHelpers.Equal(ObjectHelpers.PKColumn(v), fmt.Sprint(ObjectHelpers.PKValue(v)))
	_, err := NewUpdate().Repository(s).Where(query).Set(ObjectHelpers.ToMap(v)).Exec()
	return err
}

// Delete the object from database
func (s *Repository) Delete(v interface{}) error {
	query := WhereHelpers.Equal(ObjectHelpers.PKColumn(v), fmt.Sprint(ObjectHelpers.PKValue(v)))
	_, err := NewDelete().Repository(s).Where(query).Exec()
	return err
}

func (s *Repository) Query() *SelectStmt {
	return NewQueryOnRepository(s)
}
