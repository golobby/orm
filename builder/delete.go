package builder

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

type deleteStmt struct {
	table string
	where *deleteWhere
}

type deleteWhere struct {
	parent *deleteStmt
	conds  []string
}

func (w *deleteWhere) And(parts ...string) *deleteWhere {
	w.conds = append(w.conds, "AND "+strings.Join(parts, " "))
	return w
}

func (w *deleteWhere) Stmt() *deleteStmt {
	return w.parent
}
func (w *deleteWhere) Or(parts ...string) *deleteWhere {
	w.conds = append(w.conds, "OR "+strings.Join(parts, " "))
	return w
}
func (w *deleteWhere) Not(parts ...string) *deleteWhere {
	w.conds = append(w.conds, "NOT "+strings.Join(parts, " "))
	return w
}

func (w *deleteWhere) String() string {
	return strings.Join(w.conds, " ")
}

func (d *deleteStmt) Where(parts ...string) *deleteWhere {
	w := &deleteWhere{conds: []string{strings.Join(parts, " ")}, parent: d}
	d.where = w
	return w
}

func (d *deleteStmt) SQL() (string, error) {
	return fmt.Sprintf("DELETE FROM %s WHERE %s", d.table, d.where.String()), nil
}

func (d *deleteStmt) ExecContext(ctx context.Context, db *sql.DB, args ...interface{}) (sql.Result, error) {
	s, err := d.SQL()
	if err != nil {
		return nil, err
	}
	return exec(context.Background(), db, s, args)
}
func (d *deleteStmt) Exec(db *sql.DB, args ...interface{}) (sql.Result, error) {
	query, err := d.SQL()
	if err != nil {
		return nil, err
	}
	return exec(context.Background(), db, query, args)

}

func NewDelete(table string) *deleteStmt {
	return &deleteStmt{table: table}
}
