package orm

import (
	"database/sql"
	"fmt"

	"github.com/jedib0t/go-pretty/table"
)

type connection struct {
	Name    string
	Dialect *Dialect
	DB      *sql.DB
	Schemas map[string]*schema
	Logger  Logger
}

func (c *connection) Schematic() {
	fmt.Printf("SQL Dialect: %s\n", c.Dialect.DriverName)
	for t, schema := range c.Schemas {
		fmt.Printf("t: %s\n", t)
		w := table.NewWriter()
		w.AppendHeader(table.Row{"SQL Name", "Type", "Is Primary Key", "Is Virtual"})
		for _, field := range schema.fields {
			w.AppendRow(table.Row{field.Name, field.Type, field.IsPK, field.Virtual})
		}
		fmt.Println(w.Render())
		for t, rel := range schema.relations {
			switch rel.(type) {
			case HasOneConfig:
				fmt.Printf("%s 1-1 %s => %+v\n", t, t, rel)
			case HasManyConfig:
				fmt.Printf("%s 1-N %s => %+v\n", t, t, rel)

			case BelongsToConfig:
				fmt.Printf("%s N-1 %s => %+v\n", t, t, rel)

			case BelongsToManyConfig:
				fmt.Printf("%s N-N %s => %+v\n", t, t, rel)
			}
		}
		fmt.Println("")
	}
}

func (c *connection) getSchema(t string) *schema {
	return c.Schemas[t]
}

func (c *connection) setSchema(e Entity, s *schema) {
	var configurator EntityConfigurator
	e.ConfigureEntity(&configurator)
	c.Schemas[configurator.table] = s
}

func GetConnection(name string) *connection {
	return globalConnections[name]
}

func (c *connection) exec(q string, args ...any) (sql.Result, error) {
	globalLogger.Debugf(q)
	globalLogger.Debugf("%v", args)
	return c.DB.Exec(q, args...)
}

func (c *connection) query(q string, args ...any) (*sql.Rows, error) {
	globalLogger.Debugf(q)
	globalLogger.Debugf("%v", args)
	return c.DB.Query(q, args...)
}

func (c *connection) queryRow(q string, args ...any) *sql.Row {
	globalLogger.Debugf(q)
	globalLogger.Debugf("%v", args)
	return c.DB.QueryRow(q, args...)
}
