package builder

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

type updateStmt struct {
	table string
	where *updateWhere
}

type updateWhere struct {
	parent *updateStmt
	conds  []string
}

func (w *updateWhere) And(parts ...string) *updateWhere {
	w.conds = append(w.conds, "AND "+strings.Join(parts, " "))
	return w
}

func (w *updateWhere) Stmt() *updateStmt {
	return w.parent
}
func (w *updateWhere) Or(parts ...string) *updateWhere {
	w.conds = append(w.conds, "OR "+strings.Join(parts, " "))
	return w
}
func (w *updateWhere) Not(parts ...string) *updateWhere {
	w.conds = append(w.conds, "NOT "+strings.Join(parts, " "))
	return w
}

func (w *updateWhere) String() string {
	return strings.Join(w.conds, " ")
}

func (d *updateStmt) Where(parts ...string) *updateWhere {
	w := &updateWhere{conds: []string{strings.Join(parts, " ")}, parent: d}
	d.where = w
	return w
}

func (d *updateStmt) SQL() (string, error) {
	return fmt.Sprintf("DELETE FROM %s WHERE %s", d.table, d.where.String()), nil
}

func (d *updateStmt) ExecContext(ctx context.Context, db *sql.DB, args ...interface{}) (sql.Result, error) {
	s, err := d.SQL()
	if err != nil {
		return nil, err
	}
	return exec(context.Background(), db, s, args)
}
func (d *updateStmt) Exec(db *sql.DB, args ...interface{}) (sql.Result, error) {
	query, err := d.SQL()
	if err != nil {
		return nil, err
	}
	return exec(context.Background(), db, query, args)

}
func NewUpdate(table string) *updateStmt {
	return &updateStmt{table: table}
}
