package builder

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strings"

	"github.com/golobby/sql/binder"
)

type placeHolderGenerators struct {
	Postgres func(n int) string
	MySQL    func(n int) string
}

var PlaceHolderGenerators = &placeHolderGenerators{
	Postgres: postgresPlaceholder,
	MySQL:    mySQLPlaceHolder,
}

func postgresPlaceholder(n int) string {
	output := []string{}
	for i := 1; i < n+1; i++ {
		output = append(output, fmt.Sprintf("$%d", i))
	}
	return strings.Join(output, ", ")
}

func mySQLPlaceHolder(n int) string {
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

type objectHelpers struct {
	ColumnsOf func(v interface{}) []string
	TableName func(v interface{}) string
}

var ObjectHelpers = &objectHelpers{
	//Returns a list of string which are the columns that a struct repreasent based on binder tags.
	//Best usage would be to generate these column names in startup.
	ColumnsOf: columnsOf,
	// Returns a string which is the table name ( by convention is TYPEs ) of given object
	TableName: tableName,
}

func tableName(v interface{}) string {
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	parts := strings.Split(t.Name(), ".")
	return parts[len(parts)-1] + "s"
}

func columnsOf(v interface{}) []string {
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

type objectMetadata struct {
	Table   string
	Columns []string
}

func ObjectMetadataFrom(v interface{}) *objectMetadata {
	return &objectMetadata{
		Table:   ObjectHelpers.TableName(v),
		Columns: ObjectHelpers.ColumnsOf(v),
	}
}
