package orm

import (
	"database/sql"
	"github.com/gertd/go-pluralize"
)

type EntityConfigurator struct {
	connection        string
	table             string
	this              Entity
	relations         map[string]interface{}
	resolveRelations  []func()
	columnConstraints []*FieldConfigurator
}

func newEntityConfigurator() *EntityConfigurator {
	return &EntityConfigurator{}
}

func (ec *EntityConfigurator) Table(name string) *EntityConfigurator {
	ec.table = name
	return ec
}

func (ec *EntityConfigurator) Connection(name string) *EntityConfigurator {
	ec.connection = name
	return ec
}

func (ec *EntityConfigurator) HasMany(property Entity, config HasManyConfig) *EntityConfigurator {
	if ec.relations == nil {
		ec.relations = map[string]interface{}{}
	}
	ec.resolveRelations = append(ec.resolveRelations, func() {
		if config.PropertyForeignKey != "" && config.PropertyTable != "" {
			ec.relations[config.PropertyTable] = config
			return
		}
		configurator := newEntityConfigurator()
		property.ConfigureEntity(configurator)

		if config.PropertyTable == "" {
			config.PropertyTable = configurator.table
		}

		if config.PropertyForeignKey == "" {
			config.PropertyForeignKey = pluralize.NewClient().Singular(ec.table) + "_id"
		}

		ec.relations[configurator.table] = config

		return
	})
	return ec
}

func (ec *EntityConfigurator) HasOne(property Entity, config HasOneConfig) *EntityConfigurator {
	if ec.relations == nil {
		ec.relations = map[string]interface{}{}
	}
	ec.resolveRelations = append(ec.resolveRelations, func() {
		if config.PropertyForeignKey != "" && config.PropertyTable != "" {
			ec.relations[config.PropertyTable] = config
			return
		}

		configurator := newEntityConfigurator()
		property.ConfigureEntity(configurator)

		if config.PropertyTable == "" {
			config.PropertyTable = configurator.table
		}
		if config.PropertyForeignKey == "" {
			config.PropertyForeignKey = pluralize.NewClient().Singular(ec.table) + "_id"
		}

		ec.relations[configurator.table] = config
		return
	})
	return ec
}

func (ec *EntityConfigurator) BelongsTo(owner Entity, config BelongsToConfig) *EntityConfigurator {
	if ec.relations == nil {
		ec.relations = map[string]interface{}{}
	}
	ec.resolveRelations = append(ec.resolveRelations, func() {
		if config.ForeignColumnName != "" && config.LocalForeignKey != "" && config.OwnerTable != "" {
			ec.relations[config.OwnerTable] = config
			return
		}
		ownerConfigurator := newEntityConfigurator()
		owner.ConfigureEntity(ownerConfigurator)
		if config.OwnerTable == "" {
			config.OwnerTable = ownerConfigurator.table
		}
		if config.LocalForeignKey == "" {
			config.LocalForeignKey = pluralize.NewClient().Singular(ownerConfigurator.table) + "_id"
		}
		if config.ForeignColumnName == "" {
			config.ForeignColumnName = "id"
		}
		ec.relations[ownerConfigurator.table] = config
	})
	return ec
}

func (ec *EntityConfigurator) BelongsToMany(owner Entity, config BelongsToManyConfig) *EntityConfigurator {
	if ec.relations == nil {
		ec.relations = map[string]interface{}{}
	}
	ec.resolveRelations = append(ec.resolveRelations, func() {
		ownerConfigurator := newEntityConfigurator()
		owner.ConfigureEntity(ownerConfigurator)

		if config.OwnerLookupColumn == "" {
			var pkName string
			for _, field := range genericFieldsOf(owner) {
				if field.IsPK {
					pkName = field.Name
				}
			}
			config.OwnerLookupColumn = pkName

		}
		if config.OwnerTable == "" {
			config.OwnerTable = ownerConfigurator.table
		}
		if config.IntermediateTable == "" {
			panic("cannot infer intermediate table yet")
		}
		if config.IntermediatePropertyID == "" {
			config.IntermediatePropertyID = pluralize.NewClient().Singular(ownerConfigurator.table) + "_id"
		}
		if config.IntermediateOwnerID == "" {
			config.IntermediateOwnerID = pluralize.NewClient().Singular(ec.table) + "_id"
		}

		ec.relations[ownerConfigurator.table] = config
	})
	return ec
}

type FieldsConfigurator struct {
	ec *EntityConfigurator
}

func (ec *EntityConfigurator) Fields() *FieldsConfigurator {
	return &FieldsConfigurator{ec: ec}
}

type FieldConfigurator struct {
	fieldName   string
	fc          *FieldsConfigurator
	nullable    sql.NullBool
	primaryKey  bool
	column      string
	isCreatedAt bool
	isUpdatedAt bool
	isDeletedAt bool
}

func (fc *FieldConfigurator) CanBeNull() *FieldConfigurator {
	fc.nullable = sql.NullBool{
		Bool:  true,
		Valid: true,
	}
	return fc
}
func (fc *FieldConfigurator) CannotBeNull() *FieldConfigurator {
	fc.nullable = sql.NullBool{
		Bool:  false,
		Valid: true,
	}
	return fc
}

func (fc *FieldConfigurator) IsPrimaryKey() *FieldConfigurator {
	fc.primaryKey = true
	return fc
}

func (fc *FieldConfigurator) IsCreatedAt() *FieldConfigurator {
	fc.isCreatedAt = true
	return fc
}

func (fc *FieldConfigurator) IsUpdatedAt() *FieldConfigurator {
	fc.isUpdatedAt = true
	return fc
}

func (fc *FieldConfigurator) IsDeletedAt() *FieldConfigurator {
	fc.isDeletedAt = true
	return fc
}

func (fc *FieldConfigurator) ColumnName(name string) *FieldConfigurator {
	fc.column = name
	return fc
}

func (fc *FieldsConfigurator) Field(name string) *FieldConfigurator {
	cc := &FieldConfigurator{fc: fc, fieldName: name}
	fc.ec.columnConstraints = append(fc.ec.columnConstraints, cc)
	return cc
}
func (fc *FieldConfigurator) Also() *FieldsConfigurator {
	return fc.fc
}
