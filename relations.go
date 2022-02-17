package orm

import (
	"context"
	"fmt"
	"github.com/gertd/go-pluralize"
	"reflect"
)

type hasManyConfig struct {
	PropertyTable      string
	PropertyForeignKey string
}
type HasManyConfigurator func(config *hasManyConfig)

var HasManyConfigurators = &struct {
	PropertyTable      func(name string) HasManyConfigurator
	PropertyForeignKey func(name string) HasManyConfigurator
}{
	PropertyTable: func(name string) HasManyConfigurator {
		return func(config *hasManyConfig) {
			config.PropertyTable = name
		}
	},
	PropertyForeignKey: func(name string) HasManyConfigurator {
		return func(config *hasManyConfig) {
			config.PropertyForeignKey = name
		}
	},
}

func (e *entity) HasMany(out IEntity, configs ...HasManyConfigurator) error {
	s := _globalORM.metadatas[out.Table()]
	c := &hasManyConfig{}
	for _, config := range configs {
		config(c)
	}
	//settings default config values
	if c.PropertyTable == "" {
		c.PropertyTable = tableName(out)
	}
	if c.PropertyForeignKey == "" {
		c.PropertyForeignKey = pluralize.NewClient().Singular(s.Table) + "_id"
	}
	t := reflect.TypeOf(out)
	v := reflect.ValueOf(out)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}
	if t.Kind() == reflect.Slice {
		t = t.Elem()
	}
	ph := _globalORM.dialect.PlaceholderChar
	target := reflect.New(t).Interface()
	repo := Initialize(_globalORM.conn, _globalORM.dialect, target)
	if _globalORM.dialect.IncludeIndexInPlaceholder {
		ph = ph + fmt.Sprint(1)
	}
	var q string
	var args []interface{}

	q, args = newSelect().
		From(c.PropertyTable).
		Where(WhereHelpers.Equal(c.PropertyForeignKey, ph)).
		WithArgs(_globalORM.getPkValue(e.obj)).
		Build()

	if q == "" {
		return fmt.Errorf("cannot build the query")
	}
	return s.BindContext(context.Background(), out, q, args...)
}

type HasOneConfig struct {
	PropertyTable      string
	PropertyForeignKey string
}
type HasOneConfigurator func(config *HasOneConfig)

var HasOneConfigurators = &struct {
	PropertyTable      func(name string) HasOneConfigurator
	PropertyForeignKey func(name string) HasOneConfigurator
}{
	PropertyTable: func(name string) HasOneConfigurator {
		return func(config *HasOneConfig) {
			config.PropertyTable = name
		}
	},
	PropertyForeignKey: func(name string) HasOneConfigurator {
		return func(config *HasOneConfig) {
			config.PropertyForeignKey = name
		}
	},
}

func (e *entity) HasOne(out IEntity, configs ...HasOneConfigurator) error {
	s := _globalORM.metadatas[out.Table()]
	c := &HasOneConfig{}
	for _, config := range configs {
		config(c)
	}
	//settings default config values
	if c.PropertyTable == "" {
		c.PropertyTable = tableName(out)
	}
	if c.PropertyForeignKey == "" {
		c.PropertyForeignKey = pluralize.NewClient().Singular(s.Table) + "_id"
	}
	t := reflect.TypeOf(out)
	v := reflect.ValueOf(out)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}
	if t.Kind() == reflect.Slice {
		t = t.Elem()
	}
	ph := _globalORM.dialect.PlaceholderChar
	target := reflect.New(t).Interface()
	repo := Initialize(_globalORM.conn, _globalORM.dialect, target)
	if _globalORM.dialect.IncludeIndexInPlaceholder {
		ph = ph + fmt.Sprint(1)
	}
	var q string
	var args []interface{}

	q, args = newSelect().
		From(c.PropertyTable).
		Where(WhereHelpers.Equal(c.PropertyForeignKey, ph)).
		WithArgs(_globalORM.getPkValue(e.obj)).
		Build()

	if q == "" {
		return fmt.Errorf("cannot build the query")
	}
	return repo.BindContext(context.Background(), out, q, args...)
}

type BelongsToConfig struct {
	OwnerTable        string
	LocalForeignKey   string
	ForeignColumnName string
}
type BelongsToConfigurator func(config *BelongsToConfig)

var BelongsToConfigurators = &struct {
	OwnerTable        func(name string) BelongsToConfigurator
	LocalKey          func(name string) BelongsToConfigurator
	ForeignColumnName func(name string) BelongsToConfigurator
}{
	OwnerTable: func(name string) BelongsToConfigurator {
		return func(config *BelongsToConfig) {
			config.OwnerTable = name
		}
	},
	LocalKey: func(name string) BelongsToConfigurator {
		return func(config *BelongsToConfig) {
			config.LocalForeignKey = name
		}
	},
	ForeignColumnName: func(name string) BelongsToConfigurator {
		return func(config *BelongsToConfig) {
			config.ForeignColumnName = name
		}
	},
}

func (e *entity) BelongsTo(out interface{}, configs ...BelongsToConfigurator) error {
	c := &BelongsToConfig{}
	for _, config := range configs {
		config(c)
	}
	if c.OwnerTable == "" {
		c.OwnerTable = tableName(out)
	}
	if c.LocalForeignKey == "" {
		c.LocalForeignKey = pluralize.NewClient().Singular(c.OwnerTable) + "_id"
	}
	if c.ForeignColumnName == "" {
		c.ForeignColumnName = "id"
	}
	t := reflect.TypeOf(out)
	v := reflect.ValueOf(out)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}
	ph := _globalORM.dialect.PlaceholderChar
	if _globalORM.dialect.IncludeIndexInPlaceholder {
		ph = ph + "1"
	}
	target := reflect.New(t).Interface()
	repo := Initialize(_globalORM.conn, _globalORM.dialect, target)

	ownerIDidx := 0
	for idx, field := range _globalORM.metadata.Fields {
		if field.Name == c.LocalForeignKey {
			ownerIDidx = idx
		}
	}

	ownerID := _globalORM.valuesOf(e.obj, true)[ownerIDidx]

	q, args := newSelect().
		From(c.OwnerTable).
		Where(WhereHelpers.Equal(c.ForeignColumnName, ph)).
		WithArgs(ownerID).Build()

	return repo.BindContext(context.Background(), out, q, args...)
}

type ManyToManyConfig struct {
	IntermediateTable         string
	IntermediateLocalColumn   string
	IntermediateForeignColumn string
	ForeignTable              string
	ForeignLookupColumn       string
}
type ManyToManyConfigurator func(config *ManyToManyConfig)

var ManyToManyConfigurators = &struct {
	IntermediateTable         func(name string) ManyToManyConfigurator
	IntermediateLocalColumn   func(name string) ManyToManyConfigurator
	IntermediateForeignColumn func(name string) ManyToManyConfigurator
}{
	IntermediateTable: func(name string) ManyToManyConfigurator {
		return func(config *ManyToManyConfig) {
			config.IntermediateTable = name
		}
	},
	IntermediateLocalColumn: func(name string) ManyToManyConfigurator {
		return func(config *ManyToManyConfig) {
			config.IntermediateLocalColumn = name
		}
	},
	IntermediateForeignColumn: func(name string) ManyToManyConfigurator {
		return func(config *ManyToManyConfig) {
			config.IntermediateForeignColumn = name
		}
	},
}

var tableer = reflect.TypeOf((Table)(nil))

func (e *entity) ManyToMany(out interface{}, configs ...ManyToManyConfigurator) error {
	fmt.Println("Inaj")
	c := &ManyToManyConfig{}
	for _, config := range configs {
		config(c)
	}
	if c.IntermediateTable == "" {
		return fmt.Errorf("no way to infer many to many intermediate table yet.")
	}
	if c.IntermediateLocalColumn == "" {
		table := _globalORM.metadata.Table
		table = pluralize.NewClient().Singular(table)
		c.IntermediateLocalColumn = table + "_id"
	}
	t := reflect.TypeOf(out)
	v := reflect.New(t).Interface()

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if c.IntermediateForeignColumn == "" {
		table := tableName(v)
		c.IntermediateForeignColumn = pluralize.NewClient().Singular(table) + "_id"
	}
	if c.ForeignTable == "" {
		c.IntermediateForeignColumn = tableName(v)
	}

	sub, _ := newSelect().From(c.IntermediateTable).Where(c.IntermediateLocalColumn, "=", fmt.Sprint(_globalORM.getPkValue(e.obj))).Build()
	q, args := newSelect().
		From(c.ForeignTable).
		Where(c.ForeignLookupColumn, "in", sub).
		Build()

	fmt.Println(q)
	return Initialize(_globalORM.conn, _globalORM.dialect, out).
		BindContext(context.Background(), out, q, args...)

}
