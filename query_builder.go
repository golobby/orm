package orm

type SQL interface {
	Build() (string, []interface{})
}

type _QueryBuilder struct {
	SelectBuilder func() *SelectStmt
	InsertBuilder func() *InsertStmt
	UpdateBuilder func() *UpdateStmt
	DeleteBuilder func() *DeleteStmt
}

var QueryBuilder = _QueryBuilder{
	SelectBuilder: newSelect,
	InsertBuilder: newInsert,
	UpdateBuilder: newUpdate,
	DeleteBuilder: newDelete,
}
