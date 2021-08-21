package query

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQueryBuilding(t *testing.T) {

	t.Run("select should be ok", func(t *testing.T) {
		sql, err := New().Table("users").Select("id", "name").Query().Max("age").Min("weight").Sum("balance").Avg("height").Count("name").SQL()
		assert.NoError(t, err)
		assert.Equal(t, `SELECT id, name, MAX(age), MIN(weight), SUM(balance), AVG(height), COUNT(name) FROM users`, sql)
	})
	t.Run("select with where with like", func(t *testing.T) {
		sql, err := New().Table("users").
			Select("id", "name").Query().
			Where("id", "=", "$1").And(Like("name", "a%")).
			Or(In("name", "'jafar'", "'khadije'")).
			And(In("name", PostgresPlaceholder(2))).
			Query().
			SQL()
		assert.NoError(t, err)
		assert.Equal(t, `SELECT id, name FROM users WHERE id=$1 AND name LIKE a% OR name IN ('jafar', 'khadije') AND name IN ($1, $2)`, sql)
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
