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

# Quick Start

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
| Golobby                                                | 57250       |
| [GORM](https://gorm.io/)                               | 91979       |
| [SQLBoiler](https://github.com/volatiletech/sqlboiler) | 86998       |

[benchmark code](https://github.com/golobby/orm/blob/master/examples/benchmarks/main.go)

## License

GoLobby Sql is released under the [MIT License](http://opensource.org/licenses/mit-license.php).
