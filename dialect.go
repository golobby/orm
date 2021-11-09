package orm

type Dialect struct {
	PlaceholderChar           string
	IncludeIndexInPlaceholder bool
}

var MySQLDialect = &Dialect{
	PlaceholderChar:           "?",
	IncludeIndexInPlaceholder: false,
}
var PostgreSQLDialect = &Dialect{
	PlaceholderChar:           "$",
	IncludeIndexInPlaceholder: true,
}
