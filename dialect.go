package orm

import "database/sql"

type Dialect struct {
	DriverName                  string
	PlaceholderChar             string
	IncludeIndexInPlaceholder   bool
	AddTableNameInSelectColumns bool
	PlaceHolderGenerator        func(n int) []string
	ListTables                  func(db *sql.DB) ([]string, error)
	ListColumns                 func(db *sql.DB, table string) ([]*field, error)
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
		PlaceHolderGenerator:        mySQLPlaceHolder,
	},
	PostgreSQL: &Dialect{
		DriverName:                  "postgres",
		PlaceholderChar:             "$",
		IncludeIndexInPlaceholder:   true,
		AddTableNameInSelectColumns: true,
		PlaceHolderGenerator:        postgresPlaceholder,
	},
	SQLite3: &Dialect{
		DriverName:                  "sqlite3",
		PlaceholderChar:             "?",
		IncludeIndexInPlaceholder:   false,
		AddTableNameInSelectColumns: false,
		PlaceHolderGenerator:        mySQLPlaceHolder,
	},
}
