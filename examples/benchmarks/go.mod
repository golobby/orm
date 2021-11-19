module github.com/golobby/orm/examples

go 1.17
replace github.com/golobby/orm => ../..

require github.com/golobby/orm v1.0.0

require (
	github.com/mattn/go-sqlite3 v1.14.9
	gorm.io/driver/sqlite v1.2.4
	gorm.io/gorm v1.22.3
)

require (
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.3 // indirect
)
