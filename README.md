[![GoDoc](https://godoc.org/github.com/golobby/sql?status.svg)](https://godoc.org/github.com/golobby/sql)
[![CI](https://github.com/golobby/sql/actions/workflows/ci.yml/badge.svg)](https://github.com/golobby/sql/actions/workflows/ci.yml)
[![CodeQL](https://github.com/golobby/sql/workflows/CodeQL/badge.svg)](https://github.com/golobby/config/actions?query=workflow%3ACodeQL)
[![Go Report Card](https://goreportcard.com/badge/github.com/golobby/sql)](https://goreportcard.com/report/github.com/golobby/sql)
[![Coverage Status](https://coveralls.io/repos/github/golobby/config/badge.svg)](https://coveralls.io/github/golobby/sql?branch=master)

# sql 
GoLobby sql is a set of helpers and utilities for simpler usage of `database/sql`.

## Documentation
### Required Go Version
It requires Go `v1.11` or newer versions.

### Installation
To install this package run the following command in the root of your project.

```bash
go get github.com/golobby/sql
```

### Quick Start

The following example demonstrates how to build a query and bind it to a struct.
```go
    var db *sql.DB
    u := User{}
    _ = query.New().Table("users").
			Select("id", "name").Query().
			Where("id", "=", "$1").
			Query().Bind(db, u)
    }
```
of course you can use each sub-package seperately as well as combined usage like above.
for example just building a query:
```go
    q, _ := query.New().Table("users").
			Select("id", "name").Query().
			Where("id", "=", "$1").
			Query().SQL()
    rows, _ := db.Query(q, args...)
    //do smth with rows

```

or usage of bind seperately:
```go
    u := &User{}
    rows, _ := db.Query(`SELECT * FROM users WHERE id = $1`, 1)
    _ = bind.Bind(rows, u)
```

## License
GoLobby Sql is released under the [MIT License](http://opensource.org/licenses/mit-license.php).
