package qb

import (
	"github.com/golobby/orm"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDelete(t *testing.T) {

	t.Run("simple delete equality of id", func(t *testing.T) {
		s, args := (&orm.Delete{}).
			Table("users").
			Where("id", "=", "$1").WithArgs(1).
			Build()
		assert.Equal(t, []interface{}{1}, args)
		assert.Equal(t, `DELETE FROM users WHERE id = $1`, s)
	})
}
