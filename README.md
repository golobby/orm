[![GoDoc](https://godoc.org/github.com/golobby/orm?status.svg)](https://godoc.org/github.com/golobby/orm)
[![CI](https://github.com/golobby/orm/actions/workflows/ci.yml/badge.svg)](https://github.com/golobby/orm/actions/workflows/ci.yml)
[![CodeQL](https://github.com/golobby/orm/workflows/CodeQL/badge.svg)](https://github.com/golobby/orm/actions?query=workflow%3ACodeQL)
[![Go Report Card](https://goreportcard.com/badge/github.com/golobby/orm)](https://goreportcard.com/report/github.com/golobby/orm)
[![Coverage Status](https://coveralls.io/repos/github/golobby/orm/badge.svg)](https://coveralls.io/github/golobby/orm?branch=master)

# golobby/orm

GoLobby is a simple yet powerfull, fast, safe, customizable, type-safe database toolkit for Golang.

## Why not an ORM ?
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
