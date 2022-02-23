[![GoDoc](https://godoc.org/github.com/golobby/orm?status.svg)](https://godoc.org/github.com/golobby/orm)
[![CI](https://github.com/golobby/orm/actions/workflows/ci.yml/badge.svg)](https://github.com/golobby/orm/actions/workflows/ci.yml)
[![CodeQL](https://github.com/golobby/orm/workflows/CodeQL/badge.svg)](https://github.com/golobby/orm/actions?query=workflow%3ACodeQL)
[![Go Report Card](https://goreportcard.com/badge/github.com/golobby/orm)](https://goreportcard.com/report/github.com/golobby/orm)
[![Coverage Status](https://coveralls.io/repos/github/golobby/orm/badge.svg)](https://coveralls.io/github/golobby/orm?branch=master)

# golobby/orm

GoLobby ORM is a simple yet powerfull, fast, safe, customizable, type-safe database toolkit for Golang.

## Why another ORM ?
GoLobby ORM 

### Why not GORM ?
- GORM is so magical with it's usage of struct tags and sometimes you are basically writing code inside the struct tag.
- GORM uses lots and lots of reflection which makes it slow compare to `database/sql`.
- GORM does not follow go's idiomatic way of error handling.

### Why not SQLBoiler ?
I love sqlboiler, it's safe and generated code is really clean but it has flaws also. 
- sqlboiler uses reflection as well, less than GORM but still in hot-paths there are reflections happening.
- sqlboiler is not comfortable for starting a project from scratch with all the wiring it needs, and also complexity of having a compelete replica of production database in your local env.

## Documentation

### Required Go Version

It requires Go `v1.18` or newer versions.

### Installation

To install this package run the following command in the root of your project.

```bash
go install github.com/golobby/orm
```

### Getting Started
Let's imagine we are going to build a simple blogging application that has 3 entities, `Comment`, `Post`, `Category`.
To start using ORM you need to call **Initialize** method. It gets array of of **ConnectionConfig** objects which has:
- `Name`: Name of the connection, it can be anything you want.
- `Driver`: Name of the driver to be used when opening connection to your database.
- `ConnectionString`: connection string to connect to your db.
- `Entities`: List of entities you want to use for that connection (later we discuss more about entities.)
```go
orm.Initialize(orm.ConnectionConfig{
		Name:             "sqlite3", // Any name
		Driver:           "sqlite3", // can be "postgres" "mysql", or any normal sql driver name
		ConnectionString: ":memory:", // Any connection string that is valid for your driver.
		Entities:         []orm.Entity{&Comment{}, &Post{}, &Category{}}, // List of entities you want to use.
	})
```
Before we go further we need to talk about **Entities**, `Entity` is an interface that you ***need*** to implement for your models/entities to let ORM work with them.
So let's define our entities.

```go
package main

import "github.com/golobby/orm"

type Post struct {
	ID   int
	Text string
}

func (p Post) Schema() *orm.Schema {
	return &orm.Schema{
		Table: "posts",
    }
}

type Comment struct {
	ID int
	PostID int
	Body string
}
func (c Comment) Schema() *orm.Schema {
	return &orm.Schema{
		Table: "comments",
    }
}

type Category struct {
	ID int
	Title string
}
func (c Category) Schema() *orm.Schema {
	return &orm.Schema{Table: "categories"}
}
```
As you see for all of our entities we define a `Schema` method that returns an instance of `Schema` struct defined in orm, `Schema` struct contains all information that `ORM` needs to work with a database entity modeled in Go structs.
In `Schema` struct all fields are optional and can be infered except `Table` field which is mandatory and defines table name of the given struct.

Now let's write simple `CRUD` logic for posts.

```go
package main

import "github.com/golobby/orm"

func createPost(p *Post) error {
	err := orm.Save(p)
    return err
}
func findPost(id int) (*Post, error) {
	return orm.Find[Post](id)
}

func updatePost(p *Post) error {
	return orm.Update(p)
}

func deletePost(p *Post) error {
	return orm.Delete(p)
}

```

now that we have our post in database, let's add some comments to it.

```go
package main

func addCommentsToPost(post *Post, comments []Comment) error {
	return orm.Add[Comment](post, orm.HasManyRelation, comments)
}

func addComments(comments []Comment) error {
	return orm.SaveAll(comments)
}

// you can also create, update, delete, find comments like you saw with posts.
```

finally now we have both our posts and comments in db, let's add some categories.

## License

GoLobby ORM is released under the [MIT License](http://opensource.org/licenses/mit-license.php).
