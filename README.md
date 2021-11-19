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

`golobby/orm` is built around idea of repositories and models.

### Using Repositories

Repositories are more like EFCore DbSet objects, they map to a database table and hold various information needed for
query generation.
*Note* Since `golobby/orm` uses reflection for creating repositories it's best to build them once at the start of our
application.<br/>

```go
package main

import (
	"database/sql"
	"github.com/golobby/orm"
)

func main() {
	type User struct {
		Id   int    `bind:"id" pk:"true"`
		Name string `bind:"name"`
	}
	db, _ := sql.Open("", "")
	userRepository := orm.NewRepository(db, &User{})
	firstUser := &User{
		Id: 1,
	}
	var secondUser User
	userRepository.Fill(firstUser)      //Fills the struct from database using present fields ( better to have to PK )
	userRepository.Find(2, &secondUser) //Finds given primary key and binds it to the given struct

	newUser := &User{
		Name: "Amirreza",
	}
	userRepository.Save(newUser) //Save given object
	firstUser.Name = "Comrade"
	userRepository.Update(firstUser)  // Updates object in database
	userRepository.Delete(secondUser) // Deletes the object from database

}
```

#### More advance queries using Repositories

```go
package main

import (
	"database/sql"
	"github.com/golobby/orm"
)

func main() {
	type User struct {
		Id   int    `bind:"id" pk:"true"`
		Name string `bind:"name"`
	}
	db, _ := sql.Open("", "")
	userRepository := orm.NewRepository(db, &User{})
	var results []User
	userRepository.Query().
		Where(orm.WhereHelpers.Like("name", "%A%")).
		AndWhere(orm.WhereHelpers.Between("age", "10", "12")).
		Distinct().
		Limit(100).
		Offset(50).Bind(results)
```
### Benchmarks
for CRUD operations on 10000 records
- Create
- Read
- Update
- Delete
<br>
(on Asus ROG G512 with 32 GB of Ram, Core I7 10750)<br>

| ORM                                                    | Miliseconds |
|--------------------------------------------------------|-------------|
| Golobby                                                | 60523       |
| [GORM](https://gorm.io/)                               | 91979       |
| [SQLBoiler](https://github.com/volatiletech/sqlboiler) | 86998       |

[benchmark code](https://github.com/golobby/orm/blob/master/examples/benchmarks/main.go)

## License

GoLobby Sql is released under the [MIT License](http://opensource.org/licenses/mit-license.php).
