package orm

import (
	"reflect"

	"github.com/gertd/go-pluralize"
)

type RelationType string

const (
	RelationType_HasOne = "has_one"
	RelationType_HasMany = "has_many"
	RelationType_BelongsTo = "belongs_to"
)
type RelationMetadata struct {
	Table string
	Type string
	Lookup string
	Columns []string
}
func (o *ObjectMetadata) relationsOf() []*RelationMetadata {
	var relations []*RelationMetadata
	for _, field := range o.Fields {
		if ! field.Virtual {
			continue
		}
		lookup := pluralize.NewClient().Singular(o.Table)+"_id"
		v := reflect.New(field.Type).Interface()
		md := ObjectMetadataFrom(v, o.dialect)
		relations = append(relations, &RelationMetadata{
			Table:   tableName(field.Name),
			Lookup: lookup,
			Columns: md.Columns(true),
		})
	}
	return relations
}
