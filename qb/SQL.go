package qb

type SQL interface {
	Build() (string, []interface{})
}
