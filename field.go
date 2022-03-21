package orm

import (
	"database/sql/driver"
	"github.com/iancoleman/strcase"
	"reflect"
	"strings"
)

type field struct {
	Name        string
	IsPK        bool
	Virtual     bool
	IsCreatedAt bool
	IsUpdatedAt bool
	IsDeletedAt bool
	Nullable    bool
	Default     any
	Type        reflect.Type
}

type fieldTag struct {
	Name        string
	Virtual     bool
	PK          bool
	Nullable    bool
	Default     string
	IsCreatedAt bool
	IsUpdatedAt bool
	IsDeletedAt bool
}

func fieldMetadataFromTag(t string) fieldTag {
	if t == "" {
		return fieldTag{}
	}
	tuples := strings.Split(t, " ")
	var tag fieldTag
	for _, tuple := range tuples {
		parts := strings.Split(tuple, "=")
		key := parts[0]
		value := parts[1]
		if key == "col" {
			tag.Name = value
		} else if key == "pk" {
			tag.PK = true
		} else if key == "created_at" {
			tag.IsCreatedAt = true
		} else if key == "updated_at" {
			tag.IsUpdatedAt = true
		} else if key == "deleted_at" {
			tag.IsDeletedAt = true
		} else if key == "nullable" {
			tag.Nullable = true
		} else if key == "default" {
			tag.Default = value
		}
		if tag.Name == "_" {
			tag.Virtual = true
		}
	}
	return tag
}
func getFieldConfiguratorFor(fieldConfigurators []*FieldConfigurator, name string) *FieldConfigurator {
	for _, fc := range fieldConfigurators {
		if fc.fieldName == name {
			return fc
		}
	}
	return &FieldConfigurator{}
}

func fieldMetadata(ft reflect.StructField, fieldConfigurators []*FieldConfigurator) []*field {
	// tagParsed := fieldMetadataFromTag(ft.Tag.Get("orm"))
	var fms []*field
	fc := getFieldConfiguratorFor(fieldConfigurators, ft.Name)
	baseFm := &field{}
	baseFm.Type = ft.Type
	fms = append(fms, baseFm)
	if fc.column != "" {
		baseFm.Name = fc.column
	} else {
		baseFm.Name = strcase.ToSnake(ft.Name)
	}
	if strings.ToLower(ft.Name) == "id" || fc.primaryKey {
		baseFm.IsPK = true
	}
	if strings.ToLower(ft.Name) == "createdat" || fc.isCreatedAt {
		baseFm.IsCreatedAt = true
	}
	if strings.ToLower(ft.Name) == "updatedat" || fc.isUpdatedAt {
		baseFm.IsUpdatedAt = true
	}
	if strings.ToLower(ft.Name) == "deletedat" || fc.isDeletedAt {
		baseFm.IsDeletedAt = true
	}
	// if tagParsed.Virtual {
	// 	baseFm.Virtual = true
	// }
	if ft.Type.Kind() == reflect.Struct || ft.Type.Kind() == reflect.Ptr {
		t := ft.Type
		if ft.Type.Kind() == reflect.Ptr {
			t = ft.Type.Elem()
		}
		if !t.Implements(reflect.TypeOf((*driver.Valuer)(nil)).Elem()) {
			for i := 0; i < t.NumField(); i++ {
				fms = append(fms, fieldMetadata(t.Field(i), fieldConfigurators)...)
			}
			fms = fms[1:]
		}
	}
	return fms
}
