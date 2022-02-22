package orm

type Executable interface {
	Exec() error
}

type ExecutableQuery struct {
	*Entity
	q      string
	args   []interface{}
	bindTo interface{}
}

func (e *ExecutableQuery) Exec() error {
	db := e.obj.E().getConnection()
	rows, err := db.Query(e.q, e.args...)
	if err != nil {
		return err
	}
	return e.obj.E().getMetadata().Bind(rows, e.bindTo)
}
