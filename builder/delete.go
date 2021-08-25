package builder

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

type DeleteStmt struct {
	table string
	where *deleteWhere
}

type deleteWhere struct {
	parent *DeleteStmt
	conds  []string
}

func (w *deleteWhere) And(parts ...string) *deleteWhere {
	w.conds = append(w.conds, "AND "+strings.Join(parts, " "))
	return w
}

func (w *deleteWhere) Stmt() *DeleteStmt {
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

func (d *DeleteStmt) Where(parts ...string) *deleteWhere {
	w := &deleteWhere{conds: []string{strings.Join(parts, " ")}, parent: d}
	d.where = w
	return w
}

func (d *DeleteStmt) SQL() (string, error) {
	return fmt.Sprintf("DELETE FROM %s WHERE %s", d.table, d.where.String()), nil
}

func (d *DeleteStmt) ExecContext(ctx context.Context, db *sql.DB, args ...interface{}) (sql.Result, error) {
	s, err := d.SQL()
	if err != nil {
		return nil, err
	}
	return exec(context.Background(), db, s, args)
}
func (d *DeleteStmt) Exec(db *sql.DB, args ...interface{}) (sql.Result, error) {
	query, err := d.SQL()
	if err != nil {
		return nil, err
	}
	return exec(context.Background(), db, query, args)

}

func NewDelete(table string) *DeleteStmt {
	return &DeleteStmt{table: table}
}
