package builder

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQueryBuilding(t *testing.T) {

	t.Run("select with all aggregator functions", func(t *testing.T) {
		sql, err := NewQuery().Table("users").Select("id", "name", SelectHelpers.Max("age"), SelectHelpers.Min("weight"), SelectHelpers.Sum("balance"), SelectHelpers.Avg("height"), SelectHelpers.Count("name")).SQL()
		assert.NoError(t, err)
		assert.Equal(t, `SELECT id, name, MAX(age), MIN(weight), SUM(balance), AVG(height), COUNT(name) FROM users`, sql)
	})

	t.Run("select with comparison where", func(t *testing.T) {
		sql, err := NewQuery().Table("users").
			Select("id", "name").
			Where("id", "=", "$1").
			SQL()
		assert.NoError(t, err)
		assert.Equal(t, `SELECT id, name FROM users WHERE id = $1`, sql)
	})

	t.Run("select with not operator", func(t *testing.T) {
		sql, err := NewQuery().Table("users").
			Select("id", "name").
			Where(WhereHelpers.Not("id", "=", "$1")).
			SQL()
		assert.NoError(t, err)
		assert.Equal(t, `SELECT id, name FROM users WHERE NOT id = $1`, sql)
	})

	t.Run("select with multiple AND, OR", func(t *testing.T) {
		sql, err := NewQuery().Table("users").
			Select("id", "name").
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
		sql, err := NewQuery().Table("users").Select("id").Where(WhereHelpers.Like("name", "%a")).SQL()
		assert.NoError(t, err)
		assert.Equal(t, `SELECT id FROM users WHERE name LIKE %a`, sql)
	})

	t.Run("select WHERE IN", func(t *testing.T) {
		sql, err := NewQuery().Table("users").
			Select("id", "name").
			Where(WhereHelpers.And(WhereHelpers.In("name", "'jafar'", "'khadije'"), WhereHelpers.In("name", PlaceHolderGenerators.Postgres(2)))).
			SQL()
		assert.NoError(t, err)
		assert.Equal(t, `SELECT id, name FROM users WHERE name IN ('jafar', 'khadije') AND name IN ($1, $2)`, sql)
	})

	t.Run("select BETWEEN", func(t *testing.T) {
		sql, err := NewQuery().Table("users").
			Select("id", "name").
			Where(WhereHelpers.Between("age", "10", "18")).SQL()

		assert.NoError(t, err)
		assert.Equal(t, `SELECT id, name FROM users WHERE age BETWEEN 10 AND 18`, sql)
	})

	t.Run("select orderby desc", func(t *testing.T) {
		sql, err := NewQuery().Table("users").Select("id", "name").OrderBy("created_at", "DESC").OrderBy("updated_at", "DESC").SQL()
		assert.NoError(t, err)
		assert.Equal(t, `SELECT id, name FROM users ORDER BY created_at DESC, updated_at DESC`, sql)
	})

	t.Run("select distinct", func(t *testing.T) {
		sql, err := NewQuery().Table("users").Select("name").Distinct().Where(WhereHelpers.Less("age", "10")).SQL()
		assert.NoError(t, err)
		assert.Equal(t, `SELECT DISTINCT name FROM users WHERE age < 10`, sql)
	})

	t.Run("select orderby asc", func(t *testing.T) {
		sql, err := NewQuery().Table("users").Select("id", "name").OrderBy("created_at", "ASC").SQL()
		assert.NoError(t, err)
		assert.Equal(t, `SELECT id, name FROM users ORDER BY created_at ASC`, sql)
	})

	t.Run("select with groupby", func(t *testing.T) {
		sql, err := NewQuery().Table("users").Select("id").GroupBy("name", "age").Query().SQL()
		assert.NoError(t, err)
		assert.Equal(t, `SELECT id FROM users GROUP BY name, age`, sql)
	})

	t.Run("select with join", func(t *testing.T) {
		sql, err := NewQuery().Table("users").Select("id", "name").RightJoin("addresses").On("users.id", "=", "addresses.user_id").Query().SQL()
		assert.NoError(t, err)
		assert.Equal(t, `SELECT id, name FROM users RIGHT JOIN addresses ON users.id = addresses.user_id`, sql)
	})

	t.Run("select with multiple joins", func(t *testing.T) {
		sql, err := NewQuery().Table("users").Select("id", "name").
			RightJoin("addresses").On("users.id", "=", "addresses.user_id").
			Query().
			LeftJoin("user_credits").On("users.id", "=", "user_credits.user_id").Query().
			SQL()
		assert.NoError(t, err)
		assert.Equal(t, `SELECT id, name FROM users RIGHT JOIN addresses ON users.id = addresses.user_id LEFT JOIN user_credits ON users.id = user_credits.user_id`, sql)
	})

}
