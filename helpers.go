package orm

import (
	"context"
	"database/sql"
	"fmt"
)

type PlaceholderGenerator func(n int) []string
type placeHolderGenerators struct {
	Postgres PlaceholderGenerator
	MySQL    PlaceholderGenerator
}

var PlaceHolderGenerators = &placeHolderGenerators{
	Postgres: postgresPlaceholder,
	MySQL:    mySQLPlaceHolder,
}

func postgresPlaceholder(n int) []string {
	output := []string{}
	for i := 1; i < n+1; i++ {
		output = append(output, fmt.Sprintf("$%d", i))
	}
	return output
}

func mySQLPlaceHolder(n int) []string {
	output := []string{}
	for i := 0; i < n; i++ {
		output = append(output, "?")
	}

	return output
}

func _query(ctx context.Context, db *sql.DB, query string, args ...interface{}) (*sql.Rows, error) {
	return db.QueryContext(ctx, query, args...)
}

func _bind(ctx context.Context, db *sql.DB, v interface{}, query string, args ...interface{}) error {
	rows, err := _query(ctx, db, query, args...)
	if err != nil {
		return err
	}

	return Bind(rows, v)
}

func exec(ctx context.Context, db *sql.DB, stmt string, args ...interface{}) (sql.Result, error) {
	return db.ExecContext(ctx, stmt, args...)
}
