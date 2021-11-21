package orm

import (
	"reflect"
	"strings"

	"github.com/iancoleman/strcase"
)

type fieldMetadata struct {
	Name             string
	IsPK    		 bool
	Virtual 		 bool
	Type    reflect.Type
}

type fieldTag struct {
	Name  string
	Virtual bool
	PK    bool
}

type HasFields interface {
	Fields() []*fieldMetadata
}

func fieldMetadataFromTag(t string) fieldTag {
	if t == "" {
		return fieldTag{}
	}
	tuples := strings.Split(t, " ")
	var tag fieldTag
	kv := map[string]string{}
	for _, tuple := range tuples {
		parts := strings.Split(tuple, "=")
		key := parts[0]
		value := parts[1]
		kv[key] = value
		if key == "name" {
			tag.Name = value
		} else if key == "pk" {
			tag.PK = true
		} else if key == "virtual" {
			tag.Virtual = true
		}
	}
	return tag
}

func fieldsOf(obj interface{}, dialect *dialect) []*fieldMetadata {
	hasFields, is := obj.(HasFields)
	if is {
		return hasFields.Fields()
	}
	t := reflect.TypeOf(obj)
	for t.Kind() == reflect.Ptr {
		t = t.Elem()

	}
	if t.Kind() == reflect.Slice {
		t = t.Elem()
		for t.Kind() == reflect.Ptr {
			t = t.Elem()
		}
	}

	var fms []*fieldMetadata
	for i := 0; i < t.NumField(); i++ {
		ft := t.Field(i)
		tagParsed := fieldMetadataFromTag(ft.Tag.Get("orm"))
		fm := &fieldMetadata{}
		fm.Type = ft.Type
		if tagParsed.Name != "" {
			fm.Name = tagParsed.Name
		} else {
			fm.Name = strcase.ToSnake(ft.Name)
		}
		if tagParsed.PK || strings.ToLower(ft.Name) == "id" {
			fm.IsPK = true
		}
		if tagParsed.Virtual || ft.Type.Kind() == reflect.Struct || ft.Type.Kind() == reflect.Slice || ft.Type.Kind() == reflect.Ptr {
			fm.Virtual = true
		}
		fms = append(fms, fm)
	}
	return fms
}