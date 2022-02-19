package orm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQueryBuilding(t *testing.T) {

	t.Run("select with all aggregator functions", func(t *testing.T) {
		sql, args := Select().
			From("users").
			Select("id", "name", Aggregators.Max("age"), Aggregators.Min("weight"), Aggregators.Sum("balance"), Aggregators.Avg("height"), Aggregators.Count("name")).
			Build()
		assert.Empty(t, args)
		assert.Equal(t, `SELECT id, name, MAX(age), MIN(weight), SUM(balance), AVG(height), COUNT(name) FROM users`, sql)
	})

	t.Run("select with comparison where", func(t *testing.T) {
		sql, args := Select().From("users").
			Select("id", "name").
			Where("id", "=", "$1").
			WithArgs(1).
			Build()
		assert.Equal(t, []interface{}{1}, args)
		assert.Equal(t, `SELECT id, name FROM users WHERE id = $1`, sql)
	})

	t.Run("select with not operator", func(t *testing.T) {
		sql, args := Select().From("users").
			Select("id", "name").
			Where(WhereHelpers.Not("id", "=", "$1")).
			WithArgs(1).
			Build()
		assert.Equal(t, []interface{}{1}, args)
		assert.Equal(t, `SELECT id, name FROM users WHERE NOT id = $1`, sql)
	})

	t.Run("select with multiple AND, OR", func(t *testing.T) {
		sql, args := Select().From("users").
			Select("id", "name").
			Where(WhereHelpers.And(
				WhereHelpers.Or(
					"id = $1",
					WhereHelpers.Less("age", "$2")),
				WhereHelpers.Equal("name", "$3"))).
			WithArgs(1, 10, "'asghar'").
			Build()
		assert.Equal(t, []interface{}{1, 10, "'asghar'"}, args)
		assert.Equal(t, `SELECT id, name FROM users WHERE id = $1 OR age < $2 AND name = $3`, sql)

	})

	t.Run("select with Like", func(t *testing.T) {
		sql, args := Select().
			From("users").
			Select("id").
			Where(WhereHelpers.Like("name", "%a")).Build()
		assert.Empty(t, args)
		assert.Equal(t, `SELECT id FROM users WHERE name LIKE %a`, sql)
	})

	t.Run("select WHERE IN", func(t *testing.T) {
		sql, args := Select().From("users").
			Select("id", "name").
			Where(WhereHelpers.And(WhereHelpers.In("name", "$1", "$2"))).
			WithArgs("'jafar'", "'khadije'").
			Build()
		assert.Equal(t, []interface{}{"'jafar'", "'khadije'"}, args)
		assert.Equal(t, `SELECT id, name FROM users WHERE name IN ($1, $2)`, sql)
	})

	t.Run("select BETWEEN", func(t *testing.T) {
		sql, args := Select().From("users").
			Select("id", "name").
			Where(WhereHelpers.Between("age", "$1", "$2")).WithArgs(10, 18).Build()

		assert.Equal(t, []interface{}{10, 18}, args)
		assert.Equal(t, `SELECT id, name FROM users WHERE age BETWEEN $1 AND $2`, sql)
	})

	t.Run("select orderby desc", func(t *testing.T) {
		sql, args := Select().
			From("users").
			Select("id", "name").
			OrderBy("created_at", "DESC").
			OrderBy("updated_at", "DESC").
			Build()
		assert.Empty(t, args)
		assert.Equal(t, `SELECT id, name FROM users ORDER BY created_at DESC, updated_at DESC`, sql)
	})

	t.Run("select distinct", func(t *testing.T) {
		sql, args := Select().
			From("users").
			Select("name").
			Distinct().
			Where(WhereHelpers.Less("age", "$1")).
			WithArgs(10).
			Build()
		assert.Equal(t, []interface{}{10}, args)
		assert.Equal(t, `SELECT DISTINCT name FROM users WHERE age < $1`, sql)
	})

	t.Run("select orderby asc", func(t *testing.T) {
		sql, args := Select().
			From("users").
			Select("id", "name").
			OrderBy("created_at", "ASC").
			Build()
		assert.Empty(t, args)
		assert.Equal(t, `SELECT id, name FROM users ORDER BY created_at ASC`, sql)
	})

	t.Run("select with groupby", func(t *testing.T) {
		sql, args := Select().From("users").Select("id").GroupBy("name", "age").Build()
		assert.Empty(t, args)
		assert.Equal(t, `SELECT id FROM users GROUP BY name, age`, sql)
	})

	t.Run("select with join", func(t *testing.T) {
		sql, args := Select().
			From("users").
			Select("id", "name").
			RightJoin("addresses", "users.id", "=", "addresses.user_id").
			Build()
		assert.Empty(t, args)
		assert.Equal(t, `SELECT id, name FROM users RIGHT JOIN addresses ON users.id = addresses.user_id`, sql)
	})

	t.Run("select with multiple joins", func(t *testing.T) {
		sql, args := Select().From("users").Select("id", "name").
			RightJoin("addresses", "users.id", "=", "addresses.user_id").
			LeftJoin("user_credits", "users.id", "=", "user_credits.user_id").
			Build()
		assert.Empty(t, args)
		assert.Equal(t, `SELECT id, name FROM users RIGHT JOIN addresses ON users.id = addresses.user_id LEFT JOIN user_credits ON users.id = user_credits.user_id`, sql)
	})

	t.Run("select with limit and offset", func(t *testing.T) {
		sql, args := Select().From("users").Take(10).Skip(10).Build()
		assert.Empty(t, args)
		assert.Equal(t, `SELECT * FROM users LIMIT 10 OFFSET 10`, sql)

	})

	t.Run("select with having", func(t *testing.T) {
		sql, args := Select().From("users").Having("COUNT(users) > 10").Build()
		assert.Empty(t, args)
		assert.Equal(t, `SELECT * FROM users HAVING COUNT(users) > 10`, sql)
	})

}
