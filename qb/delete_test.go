package qb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDelete(t *testing.T) {

	t.Run("simple delete equality of id", func(t *testing.T) {
		s, args, err := NewDelete().
			Table("users").
			Where("id", "=", "$1").WithArgs(1).
			Build()
		assert.NoError(t, err)
		assert.Equal(t, []interface{}{1}, args)
		assert.Equal(t, `DELETE FROM users WHERE id = $1`, s)
	})
}