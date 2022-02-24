package querybuilder

import (
	"fmt"
)

type keyValue struct {
	Key   string
	Value interface{}
}

type PlaceholderGenerator func(n int) []string
type placeHolderGenerators struct {
	Postgres PlaceholderGenerator
	MySQL    PlaceholderGenerator
}

var PlaceHolderGenerators = &placeHolderGenerators{
	Postgres: postgresPlaceholder,
	MySQL:    mySQLPlaceHolder,
}

func postgresPlaceholder(n int) []string {
	output := []string{}
	for i := 1; i < n+1; i++ {
		output = append(output, fmt.Sprintf("$%d", i))
	}
	return output
}

func mySQLPlaceHolder(n int) []string {
	output := []string{}
	for i := 0; i < n; i++ {
		output = append(output, "?")
	}

	return output
}
