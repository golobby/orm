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


# Quick Start
## Builder
Query package supports *almost* whole SQL syntax, keywords and functions but there are probably some rough edges and some unsupported missed ones, so if there is anything missing just
add an issue and let's talk about it.<br>

I took a lot of inspiration from *Laravel*'s *Eloquent* APIs when designing the query builder API.
Each SQL stmt type has it's own Go representation struct with various useful helper methods.
### Insert
- Into
- Values
- PlaceHolderGenerator
- Exec
- ExecContext
- SQL

### Update
- Where
- WhereNot
- OrWhere
- AndWhere
- Set
- SQL
- Exec
- ExecContext

### Select
- Select
- Where
- WhereNot
- OrWhere
- AndWhere
- Having
- Limit
- Offset
- Take (alias of Limit)
- Skip (alias of Offset)
- Joins
    - InnerJoin
    - RightJoin
    - LeftJoin
    - FullOuterJoin
- OrderBy
- GroupBy
- Exec
- ExecContext
- Bind
- BindContext
- SQL

### Delete
- Where
- WhereNot
- OrWhere
- AndWhere
- Exec
- ExecContext
- SQL

## Binder
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
