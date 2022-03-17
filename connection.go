package orm

import (
	"database/sql"
	"fmt"

	"github.com/jedib0t/go-pretty/table"
)

type connection struct {
	Name     string
	Dialect  *Dialect
	DB       *sql.DB
	Schemas  map[string]*schema
	DBSchema map[string][]columnSpec
}

func (c *connection) validateDatabaseSchema() error {
	err := c.validateAllTablesArePresent()
	if err != nil {
		return err
	}

	return nil
}
func (c *connection) inferedTables() []string {
	var tables []string
	for t, s := range c.Schemas {
		tables = append(tables, t)
		for _, relC := range s.relations {
			if belongsToManyConfig, is := relC.(BelongsToManyConfig); is {
				tables = append(tables, belongsToManyConfig.IntermediateTable)
			}
		}
	}
	return tables
}

func (c *connection) validateAllTablesArePresent() error {
	for _, inferedTable := range c.inferedTables() {
		if _, exists := c.DBSchema[inferedTable]; !exists {
			return fmt.Errorf("orm infered %s but it's not found in your database, your database is out of sync", inferedTable)
		}
	}
	return nil
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
	return c.DB.Exec(q, args...)
}

func (c *connection) query(q string, args ...any) (*sql.Rows, error) {
	return c.DB.Query(q, args...)
}

func (c *connection) queryRow(q string, args ...any) *sql.Row {
	return c.DB.QueryRow(q, args...)
}
