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
	// Returns a list of string which are the columns that a struct repreasent based on binder tags.
	// Best usage would be to generate these column names in startup.
	ColumnsOf func(v interface{}) []string
	// Returns a string which is the table name ( by convention is TYPEs ) of given object
	TableName func(v interface{}) string
	// Returns a list of values of the given object, usefull for passing as args of sql exec or query
	ValuesOf func(v interface{}) []interface{}
	PrimaryKeyOf func(v interface{}) string
}

var ObjectHelpers = &objectHelpers{
	ColumnsOf: columnsOf,
	TableName: tableName,

	ValuesOf: valuesOf,
	PrimaryKeyOf: primaryKeyOf,
}

func valuesOf(o interface{}) []interface{} {
	v := reflect.ValueOf(o)
	if v.Type().Kind() == reflect.Ptr {
		v = v.Elem()
	}
	values := []interface{}{}
	for i := 0; i < v.NumField(); i++ {
		values = append(values, v.Field(i).Interface())
	}
	return values
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
func primaryKeyOf(v interface{}) string {
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()	
	}
	for i:=0;i<t.NumField();i++ {
		if tag, exists := t.Field(i).Tag.Lookup("pk"); exists {
			if tag == "true" {
				if name, exist:= t.Field(i).Tag.Lookup("bind");exist {
					return name
				}
				return t.Field(i).Name
			}
		}
	}
	return ""
}

type ObjectMetadata struct {
	// Name of the table that the object represents
	Table   string
	// List of columns that this object has.
	Columns func(...string) []string
	// primary key of this struct
	PrimaryKey string
	// index of the relation fields
	RelationField map[string]int
}

func ObjectMetadataFrom(v interface{}) *ObjectMetadata {
	return &ObjectMetadata{
		Table: ObjectHelpers.TableName(v),
		Columns: func(blacklist ...string) []string {
			allColumns := ObjectHelpers.ColumnsOf(v)
			blacklisted := strings.Join(blacklist, ";")
			columns := []string{}
			for _, col := range allColumns {
				if !strings.Contains(blacklisted, col) {
					columns = append(columns, col)
				}
			}
			return columns
		},
		PrimaryKey: primaryKeyOf(v),
	}
}

type functionCall func(string) string

type aggregators struct {
	Min   functionCall
	Max   functionCall
	Count functionCall
	Avg   functionCall
	Sum   functionCall
}

var Aggregators = &aggregators{
	Min:   makeFunctionFormatter("MIN"),
	Max:   makeFunctionFormatter("MAX"),
	Count: makeFunctionFormatter("COUNT"),
	Avg:   makeFunctionFormatter("AVG"),
	Sum:   makeFunctionFormatter("SUM"),
}

func makeFunctionFormatter(function string) func(string) string {
	return func(column string) string {
		return fmt.Sprintf("%s(%s)", function, column)
	}
}
