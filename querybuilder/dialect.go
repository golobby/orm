package querybuilder

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
