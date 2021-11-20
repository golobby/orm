package qb

import (
	"github.com/golobby/orm/ds"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdate(t *testing.T) {

	t.Run("simple update", func(t *testing.T) {
		s, args := NewUpdate().
			Table("users").
			Where(WhereHelpers.Equal("id", "$1")).
			Set(ds.KV{
				Key:   "name",
				Value: "$2",
			}).
			WithArgs(2).
			Build()

		assert.Equal(t, []interface{}{2}, args)
		assert.Equal(t, `UPDATE users SET name=$2 WHERE id = $1`, s)
	})
}
