package builder

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/golobby/sql/binder"
)

func PostgresPlaceholder(n int) string {
	output := []string{}
	for i := 1; i < n+1; i++ {
		output = append(output, fmt.Sprintf("$%d", i))
	}
	return strings.Join(output, ", ")
}

func MySQLPlaceHolder(n int) string {
	output := []string{}
	for i := 0; i < n; i++ {
		output = append(output, "?")
	}

	return strings.Join(output, ", ")
}
func _query(ctx context.Context, db *sql.DB, query string, args ...interface{}) (*sql.Rows, error) {
	return db.QueryContext(ctx, query, args...)
}

func _bind(ctx context.Context, db *sql.DB, v interface{}, query string, args ...interface{}) error {
	rows, err := _query(ctx, db, query, args...)
	if err != nil {
		return err
	}

	return binder.Bind(rows, v)
}

func exec(ctx context.Context, db *sql.DB, stmt string, args ...interface{}) (sql.Result, error) {
	return db.ExecContext(ctx, stmt, args...)
}
