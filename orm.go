package sql

import (
	"database/sql"
	"fmt"
	"github.com/golobby/sql/binder"
	"github.com/golobby/sql/builder"
)

type schema struct {
	con *sql.DB
	md  *ObjectMetadata
}

type model struct {
	*schema
	obj interface{}
}

func (s *schema) Model(obj interface{}) *model {
	return &model{
		schema: s,
		obj:    obj,
	}
}

// Save Saves a model into the DB.
func (m *model) Save() error {
	cols := m.md.Columns()

	res, err := m.con.Exec(builder.NewInsert(m.md.Table).Into(cols...).Values(ObjectHelpers.ValuesOf(m.obj)).SQL())
	if err != nil {
		return err
	}
	pk, err := res.LastInsertId()
	if err != nil {
		return err
	}
	ObjectHelpers.SetPK(m.obj, pk)
	return nil
}

// Fill fills a model inner object using result of a PK query.
func (m *model) Fill() error {
	rows, err := m.con.Query(builder.NewQuery().
		Table(m.md.Table).
		Select(m.md.Columns()...).
		Where(builder.WhereHelpers.Equal(m.md.PrimaryKey, fmt.Sprint(ObjectHelpers.PKValue(m.obj)))).SQL())
	if err != nil {
		return err
	}
	return binder.Bind(rows, m.obj)
}
