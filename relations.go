package orm

import (
	"context"
	"fmt"
	"github.com/gertd/go-pluralize"
	"reflect"
)
type HasManyConfig struct {
	PropertyTable      string
	PropertyForeignKey string
}
type HasManyConfigurator func(config *HasManyConfig)
type _HasManyDefaultConfigurators struct {
	PropertyTable      func(name string) HasManyConfigurator
	PropertyForeignKey func(name string) HasManyConfigurator
}
var HasManyConfigurators = &_HasManyDefaultConfigurators{
	PropertyTable: func(name string) HasManyConfigurator {
		return func(config *HasManyConfig) {
			config.PropertyTable = name
		}
	},
	PropertyForeignKey: func(name string) HasManyConfigurator {
		return func(config *HasManyConfig) {
			config.PropertyForeignKey = name
		}
	},
}
func (e *entity) HasMany(out interface{}, configs...HasManyConfigurator) error {
	c := &HasManyConfig{}
	for _, config := range configs {
		config(c)
	}
	//settings default config values
	if c.PropertyTable == "" {
		c.PropertyTable = tableName(out)
	}
	if c.PropertyForeignKey == "" {
		c.PropertyForeignKey = pluralize.NewClient().Singular(e.repo.metadata.Table)+"_id"
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
	ph := e.repo.dialect.PlaceholderChar
	target := reflect.New(t).Interface()
	repo := NewRepository(e.repo.conn, e.repo.dialect, target)
	if e.repo.dialect.IncludeIndexInPlaceholder {
		ph = ph + fmt.Sprint(1)
	}
	var q string
	var args []interface{}

	q, args = newSelect().
		From(c.PropertyTable).
		Where(WhereHelpers.Equal(c.PropertyForeignKey, ph)).
		WithArgs(e.repo.getPkValue(e.obj)).
		Build()

	if q == "" {
		return fmt.Errorf("cannot build the query")
	}
	return repo.BindContext(context.Background(), out, q, args...)
}
type HasOneConfig struct {
	PropertyTable      string
	PropertyForeignKey string
}
type HasOneConfigurator func(config *HasOneConfig)
type _HasOneDefaultConfigurators struct {
	PropertyTable      func(name string) HasOneConfigurator
	PropertyForeignKey func(name string) HasOneConfigurator
}
var HasOneConfigurators = &_HasOneDefaultConfigurators{
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

func (e *entity) HasOne(out interface{}, configs ...HasOneConfigurator) error {
	c := &HasOneConfig{}
	for _, config := range configs {
		config(c)
	}
	//settings default config values
	if c.PropertyTable == "" {
		c.PropertyTable = tableName(out)
	}
	if c.PropertyForeignKey == "" {
		c.PropertyForeignKey = pluralize.NewClient().Singular(e.repo.metadata.Table)+"_id"
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
	ph := e.repo.dialect.PlaceholderChar
	target := reflect.New(t).Interface()
	repo := NewRepository(e.repo.conn, e.repo.dialect, target)
	if e.repo.dialect.IncludeIndexInPlaceholder {
		ph = ph + fmt.Sprint(1)
	}
	var q string
	var args []interface{}

	q, args = newSelect().
		From(c.PropertyTable).
		Where(WhereHelpers.Equal(c.PropertyForeignKey, ph)).
		WithArgs(e.repo.getPkValue(e.obj)).
		Build()

	if q == "" {
		return fmt.Errorf("cannot build the query")
	}
	return repo.BindContext(context.Background(), out, q, args...)
}
type BelongsToConfig struct {
	OwnerTable      string
	LocalForeignKey string
}
type BelongsToConfigurator func(config *BelongsToConfig)
type _BelongsToConfigurator struct {
	OwnerTable      func(name string) BelongsToConfigurator
	LocalKey func(name string) BelongsToConfigurator
}
var BelongsToConfigurators = &_BelongsToConfigurator{
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
}
func (e *entity) BelongsTo(out interface{}, configs ...BelongsToConfigurator) error {
	c := &BelongsToConfig{}
	for _, config := range configs {
		config(c)
	}
	if c.OwnerTable == "" {
		c.OwnerTable = tableName(out)
	}
	if c.LocalForeignKey == ""{
		c.LocalForeignKey = pluralize.NewClient().Singular(c.OwnerTable)+"_id"
	}
	t := reflect.TypeOf(out)
	v := reflect.ValueOf(out)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
		v = v.Elem()
	}
	ph := e.repo.dialect.PlaceholderChar
	if e.repo.dialect.IncludeIndexInPlaceholder {
		ph = ph+"1"
	}
	target := reflect.New(t).Interface()
	repo := NewRepository(e.repo.conn, e.repo.dialect, target)

	ownerIDidx := 0
	for idx, field := range e.repo.metadata.Fields {
		if field.Name == c.LocalForeignKey {
			ownerIDidx = idx
		}
	}

	ownerID:=e.repo.valuesOf(e.obj, true)[ownerIDidx]

	q, args := newSelect().
		From(c.OwnerTable).
		Where(WhereHelpers.Equal(c.LocalForeignKey, ph)).
		WithArgs(ownerID).Build()

	return repo.BindContext(context.Background(), out, q, args...)
}