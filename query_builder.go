package orm

type SQL interface {
	Build() (string, []interface{})
}

type _QueryBuilder struct {
	SelectBuilder func() *selectStmt
	InsertBuilder func() *insertStmt
	UpdateBuilder func() *updateStmt
	DeleteBuilder func() *deleteStmt
}

var QueryBuilder = _QueryBuilder{
	SelectBuilder: newSelect,
	InsertBuilder: newInsert,
	UpdateBuilder: newUpdate,
	DeleteBuilder: newDelete,
}
