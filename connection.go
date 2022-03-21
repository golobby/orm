package orm

import (
	"database/sql"
	"fmt"

	"github.com/jedib0t/go-pretty/table"
)

type connection struct {
	Name                    string
	Dialect                 *Dialect
	DB                      *sql.DB
	Schemas                 map[string]*schema
	DBSchema                map[string][]columnSpec
	DatabaseValidations bool
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

func (c *connection) validateTablesSchemas() error {
	// check for entity tables: there should not be any struct field that does not have a coresponding column
	for table, sc := range c.Schemas {
		if columns, exists := c.DBSchema[table]; exists {
			for _, f := range sc.fields {
				found := false
				for _, c := range columns {
					if c.Name == f.Name {
						found = true
					}
				}
				if !found {
					return fmt.Errorf("column %s not found while it was inferred", f.Name)
				}
			}
		} else {
			return fmt.Errorf("tables are out of sync, %s was inferred but not present in database", table)
		}
	}

	// check for relation tables: for HasMany,HasOne relations check if OWNER pk column is in PROPERTY,
	// for BelongsToMany check intermediate table has 2 pk for two entities

	for table, sc := range c.Schemas {
		for _, rel := range sc.relations {
			switch rel.(type) {
			case BelongsToConfig:
				columns := c.DBSchema[table]
				var found bool
				for _, col := range columns {
					if col.Name == rel.(BelongsToConfig).LocalForeignKey {
						found = true
					}
				}
				if !found {
					return fmt.Errorf("cannot find local foreign key %s for relation", rel.(BelongsToConfig).LocalForeignKey)
				}
			case BelongsToManyConfig:
				columns := c.DBSchema[rel.(BelongsToManyConfig).IntermediateTable]
				var foundOwner bool
				var foundProperty bool

				for _, col := range columns {
					if col.Name == rel.(BelongsToManyConfig).IntermediateOwnerID {
						foundOwner = true
					}
					if col.Name == rel.(BelongsToManyConfig).IntermediatePropertyID {
						foundProperty = true
					}
				}
				if !foundOwner || !foundProperty {
					return fmt.Errorf("table schema for %s is not correct one of foreign keys is not present", rel.(BelongsToManyConfig).IntermediateTable)
				}
			}
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
		for _, rel := range schema.relations {
			switch rel.(type) {
			case HasOneConfig:
				fmt.Printf("%s 1-1 %s => %+v\n", t, rel.(HasOneConfig).PropertyTable, rel)
			case HasManyConfig:
				fmt.Printf("%s 1-N %s => %+v\n", t, rel.(HasManyConfig).PropertyTable, rel)

			case BelongsToConfig:
				fmt.Printf("%s N-1 %s => %+v\n", t, rel.(BelongsToConfig).OwnerTable, rel)

			case BelongsToManyConfig:
				fmt.Printf("%s N-N %s => %+v\n", t, rel.(BelongsToManyConfig).IntermediateTable, rel)
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
