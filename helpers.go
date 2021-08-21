package query

import (
	"fmt"
	"strings"
)

func PostgresPlaceholder(n int) string {
	output := []string{}
	for i := 1; i < n+1; i++ {
		output = append(output, fmt.Sprintf("$%d", i))
	}
	return strings.Join(output, ", ")
}
