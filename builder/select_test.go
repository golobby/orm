package builder

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQueryBuilding(t *testing.T) {

	t.Run("select with all aggregator functions", func(t *testing.T) {
		sql, err := NewQuery().Table("users").Select("id", "name", SelectHelpers.Max("age"), SelectHelpers.Min("weight"), SelectHelpers.Sum("balance"), SelectHelpers.Avg("height"), SelectHelpers.Count("name")).Query().SQL()
		assert.NoError(t, err)
		assert.Equal(t, `SELECT id, name, MAX(age), MIN(weight), SUM(balance), AVG(height), COUNT(name) FROM users`, sql)
	})

	t.Run("select with comparison where", func(t *testing.T) {
		sql, err := NewQuery().Table("users").
			Select("id", "name").Query().
			Where("id", "=", "$1").
			SQL()
		assert.NoError(t, err)
		assert.Equal(t, `SELECT id, name FROM users WHERE id = $1`, sql)
	})

	t.Run("select with not operator", func(t *testing.T) {
		sql, err := NewQuery().Table("users").
			Select("id", "name").Query().
			Where(WhereHelpers.Not("id", "=", "$1")).
			SQL()
		assert.NoError(t, err)
		assert.Equal(t, `SELECT id, name FROM users WHERE NOT id = $1`, sql)
	})

	t.Run("select with multiple AND, OR", func(t *testing.T) {
		sql, err := NewQuery().Table("users").
			Select("id", "name").Query().
			Where(WhereHelpers.And(
				WhereHelpers.Or(
					"id = $1",
					WhereHelpers.Less("age", "10")),
				WhereHelpers.Equal("name", "'asghar'"))).
			SQL()
		assert.NoError(t, err)
		assert.Equal(t, `SELECT id, name FROM users WHERE id = $1 OR age < 10 AND name = 'asghar'`, sql)

	})

	t.Run("select with Like", func(t *testing.T) {
		sql, err := NewQuery().Table("users").Select("id").Query().Where(WhereHelpers.Like("name", "%a")).SQL()
		assert.NoError(t, err)
		assert.Equal(t, `SELECT id FROM users WHERE name LIKE %a`, sql)
	})

	t.Run("select WHERE IN", func(t *testing.T) {
		sql, err := NewQuery().Table("users").
			Select("id", "name").Query().
			Where(WhereHelpers.And(WhereHelpers.In("name", "'jafar'", "'khadije'"), WhereHelpers.In("name", PlaceHolderGenerators.Postgres(2)))).
			SQL()
		assert.NoError(t, err)
		assert.Equal(t, `SELECT id, name FROM users WHERE name IN ('jafar', 'khadije') AND name IN ($1, $2)`, sql)
	})

	t.Run("select BETWEEN", func(t *testing.T) {
		sql, err := NewQuery().Table("users").
			Select("id", "name").Query().
			Where(WhereHelpers.Between("age", "10", "18")).SQL()

		assert.NoError(t, err)
		assert.Equal(t, `SELECT id, name FROM users WHERE age BETWEEN 10 AND 18`, sql)
	})

	t.Run("select orderby desc", func(t *testing.T) {
		sql, err := NewQuery().Table("users").Select("id", "name").Query().OrderBy("created_at", "updated_at").Desc().Query().SQL()
		assert.NoError(t, err)
		assert.Equal(t, `SELECT id, name FROM users ORDER BY created_at, updated_at DESC`, sql)
	})

	t.Run("select orderby asc", func(t *testing.T) {
		sql, err := NewQuery().Table("users").Select("id", "name").Query().OrderBy("created_at").Query().SQL()
		assert.NoError(t, err)
		assert.Equal(t, `SELECT id, name FROM users ORDER BY created_at`, sql)
	})

	t.Run("select with groupby", func(t *testing.T) {
		sql, err := NewQuery().Table("users").Select("id").Query().GroupBy("name", "age").Query().SQL()
		assert.NoError(t, err)
		assert.Equal(t, `SELECT id FROM users GROUP BY name, age`, sql)
	})

	t.Run("select with join", func(t *testing.T) {
		sql, err := NewQuery().Table("users").Select("id", "name").Query().RightJoin("addresses").On("users.id", "=", "addresses.user_id").Query().SQL()
		assert.NoError(t, err)
		assert.Equal(t, `SELECT id, name FROM users RIGHT JOIN addresses ON users.id = addresses.user_id`, sql)
	})

	t.Run("select with multiple joins", func(t *testing.T) {
		sql, err := NewQuery().Table("users").Select("id", "name").Query().
			RightJoin("addresses").On("users.id", "=", "addresses.user_id").
			Query().
			LeftJoin("user_credits").On("users.id", "=", "user_credits.user_id").Query().
			SQL()
		assert.NoError(t, err)
		assert.Equal(t, `SELECT id, name FROM users RIGHT JOIN addresses ON users.id = addresses.user_id LEFT JOIN user_credits ON users.id = user_credits.user_id`, sql)
	})

}