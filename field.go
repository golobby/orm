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

func getFieldConfiguratorFor(fieldConfigurators []*FieldConfigurator, name string) *FieldConfigurator {
	for _, fc := range fieldConfigurators {
		if fc.fieldName == name {
			return fc
		}
	}
	return &FieldConfigurator{}
}

func fieldMetadata(ft reflect.StructField, fieldConfigurators []*FieldConfigurator) []*field {
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
