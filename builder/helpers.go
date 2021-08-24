package builder

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
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

//Returns a list of string which are the columns that a struct repreasent based on binder tags.
//Best usage would be to generate these column names in startup.
func ColumnsOf(v interface{}) []string {
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	columns := []string{}

	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if tag, exists := f.Tag.Lookup("bind"); exists {
			columns = append(columns, tag)
		} else {
			columns = append(columns, f.Name)
		}
	}
	return columns
}
