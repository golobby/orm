package qb

import "fmt"

type Dialect struct {
	DriverName                  string
	PlaceholderChar             string
	IncludeIndexInPlaceholder   bool
	AddTableNameInSelectColumns bool
	PlaceHolderGenerator        func(n int) []string
}

const (
	DialectMySQL = iota + 1
	DialectPostgres
	DialectSQLite
)

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

type PlaceholderGenerator func(n int) []string

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
