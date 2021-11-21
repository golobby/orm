package orm

type dialect struct {
	PlaceholderChar           string
	IncludeIndexInPlaceholder bool
	AddTableNameInSelectColumns bool
}
type dialects struct {
	MySQL *dialect
	PostgreSQL *dialect
	SQLite3 *dialect
}

var Dialects = &dialects{
	MySQL:      _MySQLDialect,
	PostgreSQL: _PostgreSQLDialect,
	SQLite3:    _Sqlite3SQLDialect,
}

var _MySQLDialect = &dialect{
	PlaceholderChar:           "?",
	IncludeIndexInPlaceholder: false,
	AddTableNameInSelectColumns : true,

}
var _PostgreSQLDialect = &dialect{
	PlaceholderChar:           "$",
	IncludeIndexInPlaceholder: true,
	AddTableNameInSelectColumns: true,
}
var _Sqlite3SQLDialect = &dialect{
	PlaceholderChar:           "?",
	IncludeIndexInPlaceholder: false,
	AddTableNameInSelectColumns: false,
}
