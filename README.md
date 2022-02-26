[![GoDoc](https://godoc.org/github.com/golobby/orm?status.svg)](https://godoc.org/github.com/golobby/orm)

[//]: # ([![CI]&#40;https://github.com/golobby/orm/actions/workflows/ci.yml/badge.svg&#41;]&#40;https://github.com/golobby/orm/actions/workflows/ci.yml&#41;)

[//]: # ([![CodeQL]&#40;https://github.com/golobby/orm/workflows/CodeQL/badge.svg&#41;]&#40;https://github.com/golobby/orm/actions?query=workflow%3ACodeQL&#41;)

[//]: # ([![Go Report Card]&#40;https://goreportcard.com/badge/github.com/golobby/orm&#41;]&#40;https://goreportcard.com/report/github.com/golobby/orm&#41;)
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

### Examples
- [Blog Example](https://github.com/golobby/orm#blog-example)

### Required Go Version

It requires Go `v1.18` or newer versions.

### Installation

To install this package run the following command in the root of your project.

```bash
go install github.com/golobby/orm
```

### Getting Started
Let's imagine we are going to build a simple blogging application that has 3 entities, `Comment`, `Post`, `Category`. To
start using ORM you need to call **Initialize** method. It gets array of of **ConnectionConfig** objects which has:

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
#### Creating database entities 
Before we go further we need to talk about **Entities**, `Entity` is an interface that you ***need*** to implement for
your models/entities to let ORM work with them. So let's define our entities.

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
	ID     int
	PostID int
	Body   string
}

func (c Comment) Schema() *orm.Schema {
	return &orm.Schema{
		Table: "comments",
	}
}

type Category struct {
	ID    int
	Title string
}

func (c Category) Schema() *orm.Schema {
	return &orm.Schema{Table: "categories"}
}
```

As you see for all of our entities we define a `Schema` method that returns an instance of `Schema` struct defined in
orm, `Schema` struct contains all information that `ORM` needs to work with a database entity modeled in Go structs.
`Schema` has two public fields, `Connection` and `Table`, `Table` is mandatory for all usecases and `Connection` is mandatory for 
applications with more than 1 connection.

#### Create, Find, Update, Delete
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
#### Insert with relation
now that we have our post in database, let's add some comments to it. notice that comments are in relation with posts and the relation from posts view is a hasMany relationship and from comments is a belongsTo relationship.

```go
package main

func addCommentsToPost(post *Post, comments []Comment) error {
	return orm.Add[Comment](post, orm.BelongsToRelation, comments)
}

func addComments(comments []Comment) error {
	return orm.SaveAll(comments)
}

// you can also create, update, delete, find comments like you saw with posts.
```

finally, now we have both our posts and comments in db, let's add some categories.

```go
package main

func addCategoryToPost(post *Post, category *Category) error {
	return orm.Add[Category](post, orm.ManyToManyRelation, category)
}


```
#### Custom query
Now what if you want to do some complex query for example to get some posts that were created today ?

```go
package main

import "github.com/golobby/orm"

func getTodayPosts() ([]Post, error) {
	posts, err := orm.Query[Post](
		orm.
			Select().
			Where("created_at", "<", "NOW()").
			Where("created_at", ">", "TODAY()").
			OrderBy("id", "desc"))
    return posts, err
}
```
basically you can use all orm power to run any custom query, you can build any custom query using orm query builder but you can even run raw queries and use orm power to bind them to your entities.
You can see querybuilder docs in [query builder package](https://github.com/golobby/orm/tree/master/querybuilder)
```go
package main

import "github.com/golobby/orm"

func getTodayPosts() ([]Post, error) {
	return orm.RawQuery[Post]("SELECT * FROM posts WHERE created_at < NOW() and created_at > TODAY()")
}
```


### API Documentation
If you prefer (like myself) a more api oriented document this part is for you. Almost all functionalities of ORM is exposed thorough
simple functions of ORM, there are 2 or 3 types you need to know about:
- `Schema`: All data that ORM needs for working with a struct as database model, all of it's fields can be infered at startup except `Table` which is mandatory. Ofcourse you can fill any field you want and instead of ORM default that one would be used through your application.

- `Entity`: Interface which all structs that are database entities should implement, it has only one method and that just returns the `Schema` that we talk about in `Schema` section above.

Now let's talk about ORM functions. also please note that since Go1.18 is on the horizon we are using generic feature extensively to give
a really nice type-safe api.


- Basic CRUD APIs
	- `Insert`
	- `Find`
	- `Save`
	- `SaveAll`
	- `Update`
	- `Delete`

*Note*: for relationship we try to explain them using post/comment/category sample.

- Relationships
	- `Add`: This is a relation function, inserts `items` into database and also creates necessary wiring of relationships based on `relationType`.
	- `BelongsTo`: This defines a hasMany inverse relationship, relationship of a `Comment -> Post`, each comment belongs to a post.
	- `BelongsToMany`: Relationship of `Post <-> Category`, each `Post` has categories and each `Category` has posts.
	- `HasMany`: Relationship of `Post -> Comment`, each post has many comments.
	- `HasOne`: 


- Custom and Raw queries
	- `Exec`
	- `ExecRaw`
	- `Query`
	- `RawQuery`


## License

GoLobby ORM is released under the [MIT License](http://opensource.org/licenses/mit-license.php).
