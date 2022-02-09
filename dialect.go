package orm

type dialect struct {
	PlaceholderChar             string
	IncludeIndexInPlaceholder   bool
	AddTableNameInSelectColumns bool
}

var Dialects = &struct {
	MySQL      *dialect
	PostgreSQL *dialect
	SQLite3    *dialect
}{
	MySQL: &dialect{
		PlaceholderChar:             "?",
		IncludeIndexInPlaceholder:   false,
		AddTableNameInSelectColumns: true,
	},
	PostgreSQL: &dialect{
		PlaceholderChar:             "$",
		IncludeIndexInPlaceholder:   true,
		AddTableNameInSelectColumns: true,
	},
	SQLite3: &dialect{
		PlaceholderChar:             "?",
		IncludeIndexInPlaceholder:   false,
		AddTableNameInSelectColumns: false,
	},
}
