package orm

import (
	"fmt"
	"reflect"
)

func postgresPlaceholder(n int) []string {
	output := []string{}
	for i := 1; i < n+1; i++ {
		output = append(output, fmt.Sprintf("$%d", i))
	}
	return output
}

func questionMarks(n int) []string {
	output := []string{}
	for i := 0; i < n; i++ {
		output = append(output, "?")
	}

	return output
}

func entitiesAsList(entities []Entity) []string {
	var output []string

	for _, entity := range entities {
		output = append(output, reflect.TypeOf(entity).Elem().Name())
	}

	return output
}
