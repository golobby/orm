package orm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPostgresPlaceholder(t *testing.T) {
	t.Run("for 5 it should have 5", func(t *testing.T) {
		phs := postgresPlaceholder(5)
		assert.EqualValues(t, []string{"$1", "$2", "$3", "$4", "$5"}, phs)
	})
}