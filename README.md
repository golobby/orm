[![GoDoc](https://godoc.org/github.com/golobby/orm?status.svg)](https://godoc.org/github.com/golobby/orm)

[//]: # ([![CI]&#40;https://github.com/golobby/orm/actions/workflows/ci.yml/badge.svg&#41;]&#40;https://github.com/golobby/orm/actions/workflows/ci.yml&#41;)

[//]: # ([![CodeQL]&#40;https://github.com/golobby/orm/workflows/CodeQL/badge.svg&#41;]&#40;https://github.com/golobby/orm/actions?query=workflow%3ACodeQL&#41;)

[![Go Report Card](https://goreportcard.com/badge/github.com/golobby/orm)](https://goreportcard.com/report/github.com/golobby/orm)
[![Coverage Status](https://coveralls.io/repos/github/golobby/orm/badge.svg)](https://coveralls.io/github/golobby/orm?branch=master)

# golobby/orm

GoLobby ORM is a simple yet powerfull, fast, safe, customizable, type-safe database toolkit for Golang.

## Features
- Simple, type safe, elegant API with help of `Generics`
- Minimum reflection usage, mostly at startup of application
- No code generation
- Query builder for various query types
- Binding query results to entities
- Supports relationship/Association types
    - HasMany
    - HasOne
    - BelongsTo
    - BelongsToMany (ManyToMany)

## Documentation
- [Blog Example](https://github.com/golobby/orm/tree/master/blog_example.md)
- [API Documentation](https://github.com/golobby/orm/tree/master/api.md)

### Introduction
GolobbyORM an object-relational mapper (ORM) that makes it enjoyable to interact with your database. 
When using GolobbyORM, each database table has a corresponding "Entity" that is used to interact with that table.
In addition to retrieving records from the database table, GolobbyORM entities allow you to insert,
update, and delete records from the table as well.

### Creating a new Entity
Lets create a new `Entity` to represent `User` in our application.

```go
package main

import "github.com/golobby/orm"

type User struct {
  ID       int64
  Name     string
  LastName string
  Email    string
}

func (u User) ConfigureEntity(e *orm.EntityConfigurator) {
    e.Table("users")
}

func (u User) ConfigureRelations(r *orm.RelationConfigurator) {}
```
as you see our user entity is nothing else than a simple struct and two methods.
Entities in GolobbyORM are implementations of `Entity` interface which defines two methods:
- ConfigureEntity: configures table and database connection.
- ConfigureRelations: configures relations that `Entity` has with other relations.

#### Conventions
##### Column names
GolobbyORM for each struct field(except slice, arrays, maps and other nested structs) assumes a respective column named using snake case syntax.
if you want to have a custom column name you should specify it in entity struct.
```go
package main
type User struct {
	Name string `orm:"column=username"` // now this field will be mapped to `username` column in sql database. 
}
```
##### Primary Key
GolobbyORM assumes that each entity has primary key named `id`, if you want to have a custom named primary key you need to specify it in entity struct.
```go
package main
type User struct {
	PK int64 `orm:"pk=true"`
}
```

#### Fetching an entity from database
GolobbyORM makes it trivial to fetch entity from database using its primary key.
```go
user, err := orm.Find[User](1)
```
`orm.Find` is a generic function that takes a generic parameter that specifies the type of `Entity` we want to query and it's primary key value.
You can also use custom queries to get entities from database.
```go

```
## License

GoLobby ORM is released under the [MIT License](http://opensource.org/licenses/mit-license.php).
