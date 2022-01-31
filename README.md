[![GoDoc](https://godoc.org/github.com/golobby/orm?status.svg)](https://godoc.org/github.com/golobby/orm)
[![CI](https://github.com/golobby/orm/actions/workflows/ci.yml/badge.svg)](https://github.com/golobby/orm/actions/workflows/ci.yml)
[![CodeQL](https://github.com/golobby/orm/workflows/CodeQL/badge.svg)](https://github.com/golobby/orm/actions?query=workflow%3ACodeQL)
[![Go Report Card](https://goreportcard.com/badge/github.com/golobby/orm)](https://goreportcard.com/report/github.com/golobby/orm)
[![Coverage Status](https://coveralls.io/repos/github/golobby/orm/badge.svg)](https://coveralls.io/github/golobby/orm?branch=master)

# golobby/orm

GoLobby is a simple yet powerfull, fast, safe, customizable, type-safe database toolkit for Golang.

## Why another ORM ?
I started this project as an ORM, but soon after playing with relations for days, I understand that there are two ways,
one is like GORM, magical struct tag stuff + opinionated naming conventions and loading, or way of sqlboiler and generating everything in a
database-first approach.

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

It requires Go `v1.11` or newer versions.

### Installation

To install this package run the following command in the root of your project.

```bash
go get github.com/golobby/orm
```

### Quick Start
ORM uses concept of data mappers or repository for modeling the database, each table is represented with a repository.
Repositories are created from an struct which has the same schema as the table.
```go
package main
type User struct {
	Id   int64
	Name string
	Age  int
}
userRepository := orm.NewRepository(db, orm.Dialects.PostgreSQL, &User{}) // creating a repository for user.
```
when you have the repository you can start messing around.
#### Fetching Data
You can query data with repository object either through filling the object or raw query and binding it to an object.

```go
package main

import (
	"context"
	"github.com/golobby/orm"
	"database/sql"
)

type User struct {
	Id   int64
	Name string
	Age  int
}

func main() {
	userRepository := orm.NewRepository(&sql.DB{}, orm.Dialects.PostgreSQL, &User{}) // creating a repository for user.
	// Filling an object
	user1 := &User{Id: 1}
	err := userRepository.Fill(user1)

	// binding result of query to object
	q, err := userRepository.
		SelectBuilder().
		Where(orm.WhereHelpers.In("parent_id", "1", "2")).
		Build()
	var parents []*User
	userRepository.BindContext(context.Background(), &parents, q)
}
```
You can build any complex queries using orm query builder and then bind the result using the repository object.

#### Saving
You can save using `.Save` method on repository.
```go
package main

import (
	"context"
	"github.com/golobby/orm"
	"database/sql"
)

type User struct {
	Id   int64
	Name string
	Age  int
}

func main() {
	userRepository := orm.NewRepository(&sql.DB{}, orm.Dialects.PostgreSQL, &User{}) // creating a repository for user.
	// Filling an object
	user1 := &User{Id: 1}
	err := userRepository.Save(user1)
}
```

#### Updating
```go
package main

import (
	"context"
	"github.com/golobby/orm"
	"database/sql"
)

type User struct {
	Id   int64
	Name string
	Age  int
}

func main() {
	userRepository := orm.NewRepository(&sql.DB{}, orm.Dialects.PostgreSQL, &User{}) // creating a repository for user.
	// Filling an object
	user1 := &User{Id: 1, Age: 12}
	err := userRepository.Update(user1)
}
```

#### Deleting
```go
package main

import (
	"context"
	"github.com/golobby/orm"
	"database/sql"
)

type User struct {
	Id   int64
	Name string
	Age  int
}

func main() {
	userRepository := orm.NewRepository(&sql.DB{}, orm.Dialects.PostgreSQL, &User{}) // creating a repository for user.
	// Filling an object
	user1 := &User{Id: 1}
	err := userRepository.Delete(user1)
}
```

### Relations
ORM support for relations is still in alpha and API might change, but basically this is how relations can be used in orm.
#### HasMany
```go
package main

import "github.com/golobby/orm"

type Address struct {
	ID      int64
	Content string
}
type User struct {
	ID      int64
	Name    string
	Age     int
	Address Address
}

func main() {
	userRepository := orm.NewRepository(db, orm.Dialects.PostgreSQL, &User{})
	firstUser := &User{
		ID: 1,
	}
	var addresses []*Address

	err = userRepository.
		Entity(firstUser).
		HasMany(&addresses, 
			orm.HasManyConfigurators.PropertyTable("addresses"),
		)

}
```
#### HasOne
```go
package main

import "github.com/golobby/orm"

type Address struct {
	ID      int64
	Content string
}
type User struct {
	ID      int64
	Name    string
	Age     int
	Address Address
}

func main() {
	userRepository := orm.NewRepository(db, orm.Dialects.PostgreSQL, &User{})
	firstUser := &User{
		ID: 1,
	}
	var address *Address

	err = userRepository.
		Entity(firstUser).
		HasOne(address, 
			orm.HasOneConfigurators.PropertyTable("addresses"),
		)

}
```
#### BelongsTo
```go
package main

import "github.com/golobby/orm"

type Address struct {
	ID      int64
	Content string
}
type User struct {
	ID      int64
	Name    string
	Age     int
	Address Address
}

func main() {
	addressRepository := orm.NewRepository(db, orm.Dialects.PostgreSQL, &Address{})
	firstAddress := &Address{
		ID: 1,
	}
	var user *User

	err = addressRepository.
		Entity(firstAddress).
		BelongsTo(user, 
			orm.BelongsToConfigurators.OwnerTable("addresses"),
		)

}
```
#### ManyToMany
COMING SOON :)
# Benchmarks
for CRUD operations on 10000 records
- Create
- Read
- Update
- Delete
<br>
(on Asus ROG G512 with 32 GB of Ram, Core I7 10750)<br>

| ORM                                                    | Miliseconds |
|--------------------------------------------------------|-------------|
| Golobby                                                | 54862       |
| [GORM](https://gorm.io/)                               | 82606       |
| [SQLBoiler](https://github.com/volatiletech/sqlboiler) | 80189       |

[benchmark code](https://github.com/golobby/orm/blob/master/examples/benchmarks/main.go)

## License

GoLobby Sql is released under the [MIT License](http://opensource.org/licenses/mit-license.php).
