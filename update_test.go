package orm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdate(t *testing.T) {

	t.Run("simple update", func(t *testing.T) {
		s, args, err := NewUpdate().
			Table("users").
			Where(WhereHelpers.Equal("id", "$1")).
			WithArgs(2).
			Set(M{
				"name": "$2",
			}).WithArgs("'amirreza'").
			Build()

		assert.NoError(t, err)
		assert.Equal(t, []interface{}{2, "'amirreza'"}, args)
		assert.Equal(t, `UPDATE users WHERE id = $1 SET name=$2`, s)
	})
}
