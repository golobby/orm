[![GoDoc](https://godoc.org/github.com/golobby/orm?status.svg)](https://godoc.org/github.com/golobby/orm)
[![CI](https://github.com/golobby/orm/actions/workflows/ci.yml/badge.svg)](https://github.com/golobby/orm/actions/workflows/ci.yml)
[![CodeQL](https://github.com/golobby/orm/workflows/CodeQL/badge.svg)](https://github.com/golobby/orm/actions?query=workflow%3ACodeQL)
[![Go Report Card](https://goreportcard.com/badge/github.com/golobby/orm)](https://goreportcard.com/report/github.com/golobby/orm)
[![Coverage Status](https://coveralls.io/repos/github/golobby/orm/badge.svg)](https://coveralls.io/github/golobby/orm?branch=master)

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

### Quick Start
#### Creating repository
first step to use golobby/orm is creating
a new repository.
```go
package main

import (
	"github.com/golobby/orm"
	"database/sql"
)
type Model struct {
	ID int
	Name string
}
func main() {
	db, _ := sql.Open("postgres", "")
	modelRepository := orm.NewRepository(db, orm.PostgreSQLDialect, &Model{})
}
```
### Inserting a new record
```go
package main

import (
	"github.com/golobby/orm"
	"database/sql"
	"fmt"
)
type Model struct {
	ID int
	Name string
}
func main() {
	db, _ := sql.Open("postgres", "")
	modelRepository := orm.NewRepository(db, orm.PostgreSQLDialect, &Model{})
	amirreza := &Model{
		Name: "Amirreza",
	}
	err := modelRepository.Save(amirreza)
    if err != nil {
        panic(err)
    }
	fmt.Println(amirreza.ID) // primary key is now set to last inserted id
}
```
### Fetching a record from database
```go
package main

import (
	"github.com/golobby/orm"
	"database/sql"
	"fmt"
)
type Model struct {
	ID int
	Name string
}
func main() {
	db, _ := sql.Open("postgres", "")
	modelRepository := orm.NewRepository(db, orm.PostgreSQLDialect, &Model{})
	amirreza := &Model{
		ID: 1,
	}
	err := modelRepository.Fill(amirreza)
    if err != nil {
        panic(err)
    }
	fmt.Println(amirreza.Name) // primary key is now set to last inserted id
}
```
### Updating a record
```go
package main

import (
	"github.com/golobby/orm"
	"database/sql"
)
type Model struct {
	ID int
	Name string
}
func main() {
	db, _ := sql.Open("postgres", "")
	modelRepository := orm.NewRepository(db, orm.PostgreSQLDialect, &Model{})
	amirreza := &Model{
		ID: 1,
		
	}
	amirreza.Name = "comrade"
	err := modelRepository.Update(amirreza)
	if err != nil {
	    panic(err)
	}
}
```
### Deleting a record
```go
package main

import (
	"github.com/golobby/orm"
	"database/sql"
)
type Model struct {
	ID int
	Name string
}
func main() {
	db, _ := sql.Open("postgres", "")
	modelRepository := orm.NewRepository(db, orm.PostgreSQLDialect, &Model{})
	amirreza := &Model{
		ID: 1,
		
	}
	err := modelRepository.Delete(amirreza)
	if err != nil {
	    panic(err)
	}
}
```
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
| Golobby                                                | 57250       |
| [GORM](https://gorm.io/)                               | 91979       |
| [SQLBoiler](https://github.com/volatiletech/sqlboiler) | 86998       |

[benchmark code](https://github.com/golobby/orm/blob/master/examples/benchmarks/main.go)

## License

GoLobby Sql is released under the [MIT License](http://opensource.org/licenses/mit-license.php).
