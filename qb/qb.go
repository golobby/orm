package qb

type ToSql interface {
	ToSql() (string, []interface{})
}
