[![GoDoc](https://godoc.org/github.com/golobby/orm?status.svg)](https://godoc.org/github.com/golobby/sql)
[![CI](https://github.com/golobby/orm/actions/workflows/ci.yml/badge.svg)](https://github.com/golobby/sql/actions/workflows/ci.yml)
[![CodeQL](https://github.com/golobby/orm/workflows/CodeQL/badge.svg)](https://github.com/golobby/config/actions?query=workflow%3ACodeQL)
[![Go Report Card](https://goreportcard.com/badge/github.com/golobby/orm)](https://goreportcard.com/report/github.com/golobby/sql)
[![Coverage Status](https://coveralls.io/repos/github/golobby/orm/badge.svg)](https://coveralls.io/github/golobby/sql?branch=master)

# ORM
GoLobby is a simple yet powerfull, fast, safe, customizable, type-safe ORM.

## Documentation
### Required Go Version
It requires Go `v1.11` or newer versions.

### Installation
To install this package run the following command in the root of your project.

```bash
go get github.com/golobby/orm
```

# Quick Start
`golobby/orm` has 3 main features.
### Schema
Schema is a type that holds different type info and database info to help the query builder engine build a query simpler.
```go
var userSchema = orm.NewSchema(&sql.DB{}, &User{})
query, err := orm.NewQuery().
    Schema(userSchema).
    Where(orm.WhereHelpers.EqualID("1")).SQL() //"SELECT id, name FROM users WHERE id = 1"

userSchema.NewModel(&User{Id: 1}).Fill() // fills given User object with data from database
userSchema.NewModel(&User{
	Id:   1,
	Name: "amirreza",
}).Save() //Saved given object into database using the schema.
```
### QueryBuilder
Abstract SQL syntax into a Go API with builder pattern.
### Bind
Bind feature sql.Rows to a struct.
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
