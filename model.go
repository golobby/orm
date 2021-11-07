package orm

type Model struct {
	repository *Repository
	obj        interface{}
}

func (s *Repository) NewModel(obj interface{}) *Model {
	return &Model{
		repository: s,
		obj:        obj,
	}
}

// Save Saves a Model into the DB.
func (m *Model) Save() error {
	return m.repository.Save(m.obj)
}

// Fill fills a Model inner object using result of a PK query.
func (m *Model) Fill() error {
	return m.repository.Fill(m.obj)

}

// Update record in database.
func (m *Model) Update() error {
	return m.repository.Update(m.obj)
}

//Delete record of database using primary key
func (m *Model) Delete() error {
	return m.repository.Delete(m.obj)
}

func (m *Model) Query() *SelectStmt {
	return NewQueryOnRepository(m.repository)
}
