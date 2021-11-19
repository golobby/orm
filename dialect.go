package orm

type Dialect struct {
	PlaceholderChar           string
	IncludeIndexInPlaceholder bool
	AddTableNameInSelectColumns bool
}

var MySQLDialect = &Dialect{
	PlaceholderChar:           "?",
	IncludeIndexInPlaceholder: false,
	AddTableNameInSelectColumns : true,

}
var PostgreSQLDialect = &Dialect{
	PlaceholderChar:           "$",
	IncludeIndexInPlaceholder: true,
	AddTableNameInSelectColumns: true,
}
var Sqlite3SQLDialect = &Dialect{
	PlaceholderChar:           "?",
	IncludeIndexInPlaceholder: false,
	AddTableNameInSelectColumns: false,
}
