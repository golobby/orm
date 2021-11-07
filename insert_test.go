package orm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInsert(t *testing.T) {

	t.Run("simple insert for psql", func(t *testing.T) {
		sql, err := NewInsert("users").
			Into("name", "password").
			Values("'amirreza'", "'admin'").
			PlaceHolderGenerator(PlaceHolderGenerators.Postgres).
			SQL()

		assert.NoError(t, err)
		assert.Equal(t, "INSERT INTO users (name, password) VALUES ($1, $2)", sql)
	})

	t.Run("simple insert for mysql", func(t *testing.T) {
		sql, err := NewInsert("users").
			Into("name", "password").
			Values("'amirreza'", "'admin'").
			PlaceHolderGenerator(PlaceHolderGenerators.MySQL).
			SQL()

		assert.NoError(t, err)
		assert.Equal(t, "INSERT INTO users (name, password) VALUES (?, ?)", sql)
	})

}
