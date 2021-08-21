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
    q, err := query.New().Table("users").
			Select("id", "name").Query().
			Where("id", "=", "$1").
			Query().
			SQL()
    rows, err := db.Query(q, 1)
    if err != nil {
        panic(err)
    }
    u := User{}
    err = bind.Bind(rows, u)
    if err != nil {
        panic(err)
    }
```

## License
GoLobby Sql is released under the [MIT License](http://opensource.org/licenses/mit-license.php).
