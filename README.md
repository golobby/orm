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
#### Builder
Builder package contains a powerful SQL query builder.

The following example demonstrates how to build a query and bind it to a struct.
```go
    var db *sql.DB
    u := User{}
    _ = builder.NewQuery().Table("users").
			Select("id", "name").Query().
			Where("id", "=", "$1").
			Query().Bind(db, u)
    }
```
of course you can use each sub-package seperately as well as combined usage like above.
for example just building a query:
```go
    q, _ := builder.NewQuery().Table("users").
			Select("id", "name").Query().
			Where("id", "=", "$1").
			Query().SQL()
    rows, _ := db.Query(q, args...)
    //do smth with rows

```
#### Binder
Binder package contains functionallity to bind sql.Rows to a struct.
In this example we are binding result of query which contains multiple rows into slice.
```go
    users := []User{&User{}, &User{}}
    rows, _ := db.Query(`SELECT * FROM users`)
    _ = bind.Bind(rows, users)
```

bind also supports nested structs.
```go

type ComplexUser struct {
	ID      int    `bind:"id"`
	Name    string `bind:"name"`
	Address Address
}

type Address struct {
	ID   int    `bind:"id"`
	Path string `bind:"path"`
}

rows, err := db.Query(`SELECT users.id, users.name, addresses.id, addresses.path FROM users INNER JOIN addresses ON addresses.user_id = users.id`)

amirreza := &ComplexUser{}
milad := &ComplexUser{}

err = Bind(rows, []*ComplexUser{amirreza, milad})


assert.Equal(t, "amirreza", amirreza.Name)
assert.Equal(t, "milad", milad.Name)

//Nested struct also has filled
assert.Equal(t, "kianpars", amirreza.Address.Path)
assert.Equal(t, "delfan", milad.Address.Path)
assert.Equal(t, 2, milad.Address.ID)
assert.Equal(t, 1, amirreza.Address.ID)

```
for more info on `bind` see [bind\_test.go](https://github.com/golobby/sql/tree/master/bind/bind_test.go)

## License
GoLobby Sql is released under the [MIT License](http://opensource.org/licenses/mit-license.php).
