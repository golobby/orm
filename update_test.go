package orm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdate(t *testing.T) {

	t.Run("simple update", func(t *testing.T) {
		s, err := NewUpdate("users").
			Where(WhereHelpers.EqualID("$1")).
			Set(KV{
				"name": "'amirreza'",
			}).
			SQL()

		assert.NoError(t, err)
		assert.Equal(t, `UPDATE users WHERE id = $1 SET name = 'amirreza'`, s)
	})
}
