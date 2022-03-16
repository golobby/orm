package orm

import (
	"database/sql"
	"fmt"
)

type Dialect struct {
	DriverName                  string
	PlaceholderChar             string
	IncludeIndexInPlaceholder   bool
	AddTableNameInSelectColumns bool
	PlaceHolderGenerator        func(n int) []string
	QueryListTables             string
	QueryTableSchema            string
}

func getListOfTables(query string) func(db *sql.DB) ([]string, error) {
	return func(db *sql.DB) ([]string, error) {
		rows, err := db.Query(query)
		if err != nil {
			return nil, err
		}
		var tables []string
		for rows.Next() {
			var table string
			err = rows.Scan(&table)
			if err != nil {
				return nil, err
			}
			tables = append(tables, table)
		}
		return tables, nil
	}
}

type columnSpec struct {
	//0|id|INTEGER|0||1
	Name         string
	Type         string
	Nullable     bool
	DefaultValue sql.NullString
	IsPrimaryKey bool
}

func getTableSchema(query string) func(db *sql.DB, query string) ([]columnSpec, error) {
	return func(db *sql.DB, table string) ([]columnSpec, error) {
		rows, err := db.Query(fmt.Sprintf(query, table))
		if err != nil {
			return nil, err
		}
		var output []columnSpec
		for rows.Next() {
			var cs columnSpec
			var nullable string
			var pk int
			err = rows.Scan(&cs.Name, &cs.Type, &nullable, &cs.DefaultValue, &pk)
			if err != nil {
				return nil, err
			}
			cs.Nullable = nullable == "notnull"
			cs.IsPrimaryKey = pk == 1
			output = append(output, cs)
		}
		return output, nil

	}
}

var Dialects = &struct {
	MySQL      *Dialect
	PostgreSQL *Dialect
	SQLite3    *Dialect
}{
	MySQL: &Dialect{
		DriverName:                  "mysql",
		PlaceholderChar:             "?",
		IncludeIndexInPlaceholder:   false,
		AddTableNameInSelectColumns: true,
		PlaceHolderGenerator:        questionMarks,
		QueryListTables:             "SHOW TABLES",
		QueryTableSchema:            "DESCRIBE %s",
	},
	PostgreSQL: &Dialect{
		DriverName:                  "postgres",
		PlaceholderChar:             "$",
		IncludeIndexInPlaceholder:   true,
		AddTableNameInSelectColumns: true,
		PlaceHolderGenerator:        postgresPlaceholder,
		QueryListTables:             `\dt`,
		QueryTableSchema:            `\d %s`,
	},
	SQLite3: &Dialect{
		DriverName:                  "sqlite3",
		PlaceholderChar:             "?",
		IncludeIndexInPlaceholder:   false,
		AddTableNameInSelectColumns: false,
		PlaceHolderGenerator:        questionMarks,
		QueryListTables:             "SELECT name FROM sqlite_schema WHERE type='table'",
		QueryTableSchema:            `SELECT name,type,"notnull","dflt_value","pk" FROM PRAGMA_TABLE_INFO('%s')`,
	},
}
