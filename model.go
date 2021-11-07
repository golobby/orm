package orm

import "fmt"

type Model struct {
	schema *Schema
	obj    interface{}
}

func (s *Schema) NewModel(obj interface{}) *Model {
	return &Model{
		schema: s,
		obj:    obj,
	}
}

// Save Saves a Model into the DB.
func (m *Model) Save() error {
	cols := m.schema.metadata.Columns(m.schema.metadata.PrimaryKey)
	query, _ := NewInsert(m.schema.metadata.Table).Into(cols...).Values(ObjectHelpers.ValuesOf(m.obj)).SQL()
	res, err := m.schema.conn.Exec(query)
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

// Fill fills a Model inner object using result of a PK query.
func (m *Model) Fill() error {
	query, err := NewQuery().
		Table(m.schema.metadata.Table).
		Select(m.schema.metadata.Columns()...).
		Where(WhereHelpers.Equal(m.schema.metadata.PrimaryKey, fmt.Sprint(ObjectHelpers.PKValue(m.obj)))).SQL()
	if err != nil {
		return err
	}
	rows, err := m.schema.conn.Query(query)
	if err != nil {
		return err
	}
	return Bind(rows, m.obj)
}

// Update record in database.
func (m *Model) Update() error {
	kvs := ObjectHelpers.KeyValue(m.obj)
	pk := ObjectHelpers.PKValue(m.obj)
	pkName := m.schema.metadata.PrimaryKey
	_, err := NewUpdate().
		Schema(m.schema).
		Where(WhereHelpers.Equal(pkName, fmt.Sprint(pk))).
		Set(kvs).Exec()
	return err
}

//Delete record of database using primary key
func (m *Model) Delete() error {
	pk := ObjectHelpers.PKValue(m.obj)
	pkName := m.schema.metadata.PrimaryKey
	_, err := NewDelete().
		Schema(m.schema).
		Where(WhereHelpers.Equal(pkName, fmt.Sprint(pk))).
		Exec()
	return err
}

//func (m *Model) FirstOrCreate() error {}
