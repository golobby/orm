package orm

type ToSql interface {
	ToSql() (string, []interface{})
}
