package query

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQueryBuilding(t *testing.T) {

	t.Run("select should be ok", func(t *testing.T) {
		sql, err := New().Table("users").Select("id", "name").Query().SQL()
		assert.NoError(t, err)
		assert.Equal(t, `SELECT id, name FROM users`, sql)
	})
	t.Run("select with where", func(t *testing.T) {
		sql, err := New().Table("users").Select("id", "name").Query().Where("id", "=", "$1").Query().SQL()
		assert.NoError(t, err)
		assert.Equal(t, `SELECT id, name FROM users WHERE id=$1`, sql)
	})
	t.Run("select orderby desc", func(t *testing.T) {
		sql, err := New().Table("users").Select("id", "name").Query().OrderBy("created_at").Desc().Query().SQL()
		assert.NoError(t, err)
		assert.Equal(t, `SELECT id, name FROM users ORDER BY created_at DESC`, sql)
	})
	t.Run("select orderby asc", func(t *testing.T) {
		sql, err := New().Table("users").Select("id", "name").Query().OrderBy("created_at").Query().SQL()
		assert.NoError(t, err)
		assert.Equal(t, `SELECT id, name FROM users ORDER BY created_at`, sql)
	})

}
