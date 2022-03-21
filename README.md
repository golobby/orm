[![GoDoc](https://godoc.org/github.com/golobby/orm/?status.svg)](https://godoc.org/github.com/golobby/orm)
[![CI](https://github.com/golobby/orm/actions/workflows/ci.yml/badge.svg)](https://github.com/golobby/orm/actions/workflows/ci.yml)
[![CodeQL](https://github.com/golobby/orm/actions/workflows/codeql-analysis.yml/badge.svg)](https://github.com/golobby/orm/actions/workflows/codeql-analysis.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/golobby/orm)](https://goreportcard.com/report/github.com/golobby/orm)
[![Coverage Status](https://coveralls.io/repos/github/golobby/orm/badge.svg?r=1)](https://coveralls.io/github/golobby/orm?branch=master)

# Golobby ORM

GoLobby ORM is a lightweight yet powerful, fast, customizable, type-safe object-relational mapper for the Go programming language.

## Table Of Contents
  * [Features](#features)
    + [Introduction](#introduction)
    + [Creating a new Entity](#creating-a-new-entity)
      - [Conventions](#conventions)
        * [Timestamps](#timestamps)
        * [Column names](#column-names)
        * [Primary Key](#primary-key)
    + [Initializing ORM](#initializing-orm)
    + [Fetching an entity from a database](#fetching-an-entity-from-a-database)
    + [Saving entities or Insert/Update](#saving-entities-or-insert-update)
    + [Using raw SQL](#using-raw-sql)
    + [Deleting entities](#deleting-entities)
    + [Relationships](#relationships)
      - [HasMany](#hasmany)
      - [HasOne](#hasone)
      - [BelongsTo](#belongsto)
      - [BelongsToMany](#belongstomany)
      - [Saving with relation](#saving-with-relation)
    + [Query Builder](#query-builder)
      - [Finishers](#finishers)
        * [All](#all)
        * [Get](#get)
        * [Update](#update)
        * [Delete](#delete)
      - [Select](#select)
        * [Column names](#column-names-1)
        * [Table](#table)
        * [Where](#where)
        * [Order By](#order-by)
        * [Limit](#limit)
        * [Offset](#offset)
        * [First, Latest](#first-latest)
      - [Update](#update)
        * [Where](#where-1)
        * [Table](#table-1)
        * [Set](#set)
      - [Delete](#delete)
        * [Table](#table-2)
        * [Where](#where-2)
    + [Database Validations](#database-validations)
  * [License](#license)

## Introduction
GoLobby ORM is an object-relational mapper (ORM) that makes it enjoyable to interact with your database. 
When using Golobby ORM, each database table has a corresponding "Entity" to interact with that table using elegant APIs.

## Features
- Elegant and easy-to-use APIs with the help of Generics.
- Type-safety.
- Using reflection at startup to be fast during runtime. 
- No code generation!
- Query builder for various query types.
- Binding query results to entities.
- Supports different kinds of relationship/Association types:
    - One to one
    - One to Many
    - Many to Many

## Quick Start

The following example demonstrates how to use the GoLobby ORM.

```go
package main

import "github.com/golobby/orm"

// User entity
type User struct {
  ID        int64
  FirstName string
  LastName  string
  Email     string
  orm.Timestamps
}

// It will be called by ORM to setup entity.
func (u User) ConfigureEntity(e *orm.EntityConfigurator) {
    // Specify related database table for the entity.
    e.Table("users")
}

func main() {
  // Setup ORM
  err := orm.Initialize(orm.ORMConfig{LogLevel: orm.LogLevelDev}, orm.ConnectionConfig{
    // Name:          "default",  // Optional. Specify connection names if you have more than on database.
    Driver:           "sqlite3",  // Database type. Currently supported sqlite3, mysql, mariadb, postgresql. 
    ConnectionString: ":memory:", // Database DSN.
  })
  
  if err != nil {
	  panic(err)
  }
  
  // Find user by primary key (ID)
  user, err := orm.Find[User](1)
  
  // Update entity
  user.Email = "jack@mail.com"
  
  // Save entity
  orm.Save(&user)
}
```

### Creating a new Entity
Let's create a new `Entity` to represent `User` in our application.

```go
package main

import "github.com/golobby/orm"

type User struct {
  ID       int64
  Name     string
  LastName string
  Email    string
  orm.Timestamps
}

func (u User) ConfigureEntity(e *orm.EntityConfigurator) {
    e.Table("users").
      Connection("default") // You can omit connection name if you only have one.
	
}
```
As you see, our user entity is nothing else than a simple struct and two methods.
Entities in GoLobby ORM are implementations of `Entity` interface, which defines two methods:
- ConfigureEntity: configures table, fields, and also relations to other entities.
#### Conventions
We have standard conventions and we encourage you to follow, but if you want to change them for any reason you can use `Field` method to customize how ORM
inferres meta data from your `Entity`.

##### Column names
GoLobby ORM for each struct field(except slice, arrays, maps, and other nested structs) assumes a respective column named using snake case syntax.
If you want a custom column name, you should specify it in `ConfigureEntity` method using `Field()` method.
```go
package main

type User struct {
  Name string
}

func (u User) ConfigureEntity(e *orm.EntityConfigurator) {
    e.Field("Name").ColumnName("custom_name_for_column")

    e.Table("users")
}
```
##### Timestamps
for having `created_at`, `updated_at`, `deleted_at` timestamps in your entities you can embed `orm.Timestamps` struct in your entity,
```go
type User struct {
  ID       int64
  Name     string
  LastName string
  Email    string
  orm.Timestamps
}

```
Also, if you want custom names for them, you can do it like this.
```go
type User struct {
    ID       int64
    Name     string
    LastName string
    Email    string
    MyCreatedAt sql.NullTime
    MyUpdatedAt sql.NullTime
    MyDeletedAt sql.NullTime
}
func (u User) ConfigureEntity(e *orm.EntityConfigurator) {
    e.Field("MyCreatedAt").IsCreatedAt() // this will make ORM to use MyCreatedAt as created_at column
    e.Field("MyUpdatedAt").IsUpdatedAt() // this will make ORM to use MyUpdatedAt as created_at column
    e.Field("MyDeletedAt").IsDeletedAt() // this will make ORM to use MyDeletedAt as created_at column

    e.Table("users")
}
```
As always you use `Field` method for configuring how ORM behaves to your struct field.

##### Primary Key
GoLobby ORM assumes that each entity has a primary key named `id`; if you want a custom primary key called, you need to specify it in entity struct.
```go
package main

type User struct {
	PK int64
}
func (u User) ConfigureEntity(e *orm.EntityConfigurator) {
    e.Field("PK").IsPrimaryKey() // this will make ORM use PK field as primary key.
    e.Table("users")
}
```

### Initializing ORM
After creating our entities, we need to initialize GoLobby ORM.
```go
package main

import "github.com/golobby/orm"

func main() {
  orm.Initialize(orm.ConnectionConfig{
    // Name:             "default", You should specify connection name if you have multiple connections
    Driver:           "sqlite3",
    ConnectionString: ":memory:",
  })
}
```
After this step, we can start using ORM.
### Fetching an entity from a database
GoLobby ORM makes it trivial to fetch entities from a database using its primary key.
```go
user, err := orm.Find[User](1)
```
`orm.Find` is a generic function that takes a generic parameter that specifies the type of `Entity` we want to query and its primary key value.
You can also use custom queries to get entities from the database.
```go

user, err := orm.Query[User]().Where("id", 1).First()
user, err := orm.Query[User]().WherePK(1).First()
```
GoLobby ORM contains a powerful query builder, which you can use to build `Select`, `Update`, and `Delete` queries, but if you want to write a raw SQL query, you can.
```go
users, err := orm.QueryRaw[User](`SELECT * FROM users`)
```

### Saving entities or Insert/Update
GoLobby ORM makes it easy to persist an `Entity` to the database using `Save` method, it's an UPSERT method, if the primary key field is not zero inside the entity
it will go for an update query; otherwise, it goes for the insert.
```go
// this will insert entity into the table
err := orm.Save(&User{Name: "Amirreza"}) // INSERT INTO users (name) VALUES (?) , "Amirreza"
```
```go
// this will update entity with id = 1
orm.Save(&User{ID: 1, Name: "Amirreza2"}) // UPDATE users SET name=? WHERE id=?, "Amirreza2", 1
```
Also, you can do custom update queries using query builder or raw SQL again as well.
```go
res, err := orm.Query[User]().Where("id", 1).Update(orm.KV{"name": "amirreza2"})
```

### Using raw SQL

```go
_, affected, err := orm.ExecRaw[User](`UPDATE users SET name=? WHERE id=?`, "amirreza", 1)
```
### Deleting entities  
It is also easy to delete entities from a database.
```go
err := orm.Delete(user)
```
You can also use query builder or raw SQL.
```go
_, affected, err := orm.Query[Post]().WherePK(1).Delete()

_, affected, err := orm.Query[Post]().Where("id", 1).Delete()

```
```go
_, affected, err := orm.ExecRaw[Post](`DELETE FROM posts WHERE id=?`, 1)
```
### Relationships
GoLobby ORM makes it easy to have entities that have relationships with each other. Configuring relations is using `ConfigureEntity` method, as you will see.
#### HasMany
```go
type Post struct {}

func (p Post) ConfigureEntity(e *orm.EntityConfigurator) {
    e.Table("posts").HasMany(&Comment{}, orm.HasManyConfig{})
}
```
As you can see, we are defining a `Post` entity that has a `HasMany` relation with `Comment`. You can configure how GoLobby ORM queries `HasMany` relation with `orm.HasManyConfig` object; by default, it will infer all fields for you.
Now you can use this relationship anywhere in your code.
```go
comments, err := orm.HasMany[Comment](post).All()
```
`HasMany` and other related functions in GoLobby ORM return `QueryBuilder`, and you can use them like other query builders and create even more
complex queries for relationships. for example, you can start a query to get all comments of a post made today.
```go
todayComments, err := orm.HasMany[Comment](post).Where("created_at", "CURDATE()").All()
```
#### HasOne
Configuring a `HasOne` relation is like `HasMany`.
```go
type Post struct {}

func (p Post) ConfigureEntity(e *orm.EntityConfigurator) {
    e.Table("posts").HasOne(&HeaderPicture{}, orm.HasOneConfig{})
}
```
As you can see, we are defining a `Post` entity that has a `HasOne` relation with `HeaderPicture`. You can configure how GoLobby ORM queries `HasOne` relation with `orm.HasOneConfig` object; by default, it will infer all fields for you.
Now you can use this relationship anywhere in your code.
```go
picture, err := orm.HasOne[HeaderPicture](post)
```
`HasOne` also returns a query builder, and you can create more complex queries for relations.
#### BelongsTo
```go
type Comment struct {}

func (c Comment) ConfigureEntity(e *orm.EntityConfigurator) {
    e.Table("comments").BelongsTo(&Post{}, orm.BelongsToConfig{})
}
```
As you can see, we are defining a `Comment` entity that has a `BelongsTo` relation with `Post` that we saw earlier. You can configure how GoLobby ORM queries `BelongsTo` relation with `orm.BelongsToConfig` object; by default, it will infer all fields for you.
Now you can use this relationship anywhere in your code.
```go
post, err := orm.BelongsTo[Post](comment).First()
```
#### BelongsToMany
```go
type Post struct {}

func (p Post) ConfigureEntity(e *orm.EntityConfigurator) {
    e.Table("posts").BelongsToMany(&Category{}, orm.BelongsToManyConfig{IntermediateTable: "post_categories"})
}

type Category struct{}

func(c Category) ConfigureEntity(r *orm.EntityConfigurator) {
    e.Table("categories").BelongsToMany(&Post{}, orm.BelongsToManyConfig{IntermediateTable: "post_categories"})
}

```
We are defining a `Post` entity and a `Category` entity with a `many2many` relationship; as you can see, we must configure the IntermediateTable name, which GoLobby ORM cannot infer.
Now you can use this relationship anywhere in your code.
```go
categories, err := orm.BelongsToMany[Category](post).All()
```
#### Saving with relation
You may need to save an entity that has some kind of relationship with another entity; in that case, you can use `Add` method.
```go
orm.Add(post, comments...) // inserts all comments passed in and also sets all post_id to the primary key of the given post.
```

### Query Builder
GoLobby ORM contains a powerful query builder to help you build complex queries with ease. QueryBuilder is accessible from `orm.Query[Entity]` method
which will create a new query builder for you with given type parameter.
Query builder can build `SELECT`,`UPDATE`,`DELETE` queries for you.

#### Finishers
Finishers are methods on QueryBuilder that will some how touch database, so use them with caution.
##### All
All will generate a `SELECT` query from QueryBuilder, execute it on database and return results in a slice of OUTPUT. It's useful for queries that have multiple results.
```go
posts, err := orm.Query[Post]().All() 
```
##### Get
Get will generate a `SELECT` query from QueryBuilder, execute it on database and return results in an instance of type parameter `OUTPUT`. It's useful for when you know your query has single result.
```go
post, err := orm.Query[Post]().First().Get()
```
##### Update
Update will generate an `UPDATE` query from QueryBuilder and executes it, returns rows affected by query and any possible error.
```go
rowsAffected, err := orm.Query[Post]().WherePK(1).Set("body", "body jadid").Update()
```
##### Delete
Delete will generate a `DELETE` query from QueryBuilder and executes it, returns rows affected by query and any possible error.
```go
rowsAffected, err := orm.Query[Post]().WherePK(1).Delete()
```
#### Select
Let's start with `Select` queries.
Each `Select` query consists of following:
```sql
SELECT [column names] FROM [table name] WHERE [cond1 AND/OR cond2 AND/OR ...] ORDER BY [column] [ASC/DESC] LIMIT [N] OFFSET [N] GROUP BY [col]
```
Query builder has methods for constructing each part, of course not all of these parts are necessary.
##### Column names
for setting column names to select use `Select` method as following:
```go
orm.Query[Post]().Select("id", "title")
```
##### Table
for setting table name for select use `Table` method as following:
```go
orm.Query[Post]().Table("users")
```
##### Where
for adding where conditions based on what kind of where you want you can use any of following:
```go
orm.Query[Post]().Where("name", "amirreza") // Equal mode: WHERE name = ?, ["amirreza"]
orm.Query[Post]().Where("age", "<", 19) // Operator mode: WHERE age < ?, [19]
orm.Query[Post]().WhereIn("id", 1,2,3,4,5) // WhereIn: WHERE id IN (?,?,?,?,?), [1,2,3,4,5]
```
You can also chain these together.
```go
orm.Query[Post]().
	Where("name", "amirreza").
	AndWhere("age", "<", 10).
	OrWhere("id", "!=", 1)
    // WHERE name = ? AND age < ? OR id != ?, ["amirreza", 10, 1]
```
##### Order By
You can set order by of query using `OrderBy` as following.
```go
orm.Query[Post]().OrderBy("id", orm.ASC) // ORDER BY id ASC
orm.Query[Post]().OrderBy("id", orm.DESC) // ORDER BY id DESC
```

##### Limit
You can set limit setting of query using `Limit` as following
```go
orm.Query[Post]().Limit1(1) // LIMIT 1
```

##### Offset
You can set limit setting of query using `Offset` as following
```go
orm.Query[Post]().Offset(1) // OFFSET 1
```

##### First, Latest
You can use `First`, `Latest` method which are also executers of query as you already seen to get first or latest record.
```go
orm.Query[Post]().First() // SELECT * FROM posts ORDER BY id ASC LIMIT 1
orm.Query[Post]().Latest() // SELECT * FROM posts ORDER BY id DESC LIMIT 1
```
#### Update
Each `Update` query consists of following:
```sql
UPDATE [table name] SET [col=val] WHERE [cond1 AND/OR cond2 AND/OR ...]
```
##### Where
Just like select where stuff, same code.

##### Table
Same as select.

##### Set
You can use `Set` method to set value.
```go
orm.Query[Message]().
  Where("id", 1).
  Set("read", true, "seen", true).
  Update() // UPDATE posts SET read=?, seen=? WHERE id = ?, [true, true, 1]
```

#### Delete
Each `Delete` query consists of following:
```sql
DELETE FROM [table name] WHERE [cond1 AND/OR cond2 AND/OR ...]
```
##### Table
Same as Select and Update.
##### Where
Same as Select and Update.
### Database Validations
Golobby ORM can validate your database state and compare it to your entities and if your database and code are not in sync give you error.
Currently there are two database validations possible:
1. Validate all necessary tables exists.
2. Validate all tables contain necessary columns.
You can enable database validations feature by enabling `DatabaseValidations` flag in your ConnectionConfig.
```go
return orm.SetupConnections(orm.ConnectionConfig{
    Name:                    "default",
    DB:                      db,
    Dialect:                 orm.Dialects.SQLite3,
    Entities:                []orm.Entity{&Post{}, &Comment{}, &Category{}, &HeaderPicture{}},
    DatabaseValidations: true,
  })
```
## License
GoLobby ORM is released under the [MIT License](http://opensource.org/licenses/mit-license.php).
