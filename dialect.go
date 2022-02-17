package orm

type dialect struct {
	DriverName                  string
	PlaceholderChar             string
	IncludeIndexInPlaceholder   bool
	AddTableNameInSelectColumns bool
}

const (
	DialectMySQL = iota + 1
	DialectPostgres
	DialectSQLite
)

var dialects = &struct {
	MySQL      *dialect
	PostgreSQL *dialect
	SQLite3    *dialect
}{
	MySQL: &dialect{
		DriverName:                  "mysql",
		PlaceholderChar:             "?",
		IncludeIndexInPlaceholder:   false,
		AddTableNameInSelectColumns: true,
	},
	PostgreSQL: &dialect{
		DriverName:                  "postgres",
		PlaceholderChar:             "$",
		IncludeIndexInPlaceholder:   true,
		AddTableNameInSelectColumns: true,
	},
	SQLite3: &dialect{
		DriverName:                  "sqlite3",
		PlaceholderChar:             "?",
		IncludeIndexInPlaceholder:   false,
		AddTableNameInSelectColumns: false,
	},
}
