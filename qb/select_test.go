package qb

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSelect(t *testing.T) {
	t.Run("only select * from table", func(t *testing.T) {
		s := Select{}
		s.Table = "users"
		str, args, err := s.ToSql()
		assert.NoError(t, err)
		assert.Empty(t, args)
		assert.Equal(t, "SELECT * FROM users", str)
	})
	t.Run("select with where", func(t *testing.T) {
		s := Select{}
		s.Table = "users"
		s.PlaceholderGenerator = Dialects.MySQL.PlaceHolderGenerator
		s.Where = &Where{
			Cond: Cond{
				Lhs: "age",
				Op:  Eq,
				Rhs: 10,
			},
		}
		str, args, err := s.ToSql()
		assert.NoError(t, err)
		assert.EqualValues(t, []interface{}{10}, args)
		assert.Equal(t, "SELECT * FROM users WHERE age = ?", str)
	})
	t.Run("select with order by", func(t *testing.T) {
		s := Select{}
		s.Table = "users"
		s.OrderBy = &OrderBy{
			Columns: []string{"created_at", "updated_at"},
			Order:   OrderByASC,
		}
		str, args, err := s.ToSql()
		assert.NoError(t, err)
		assert.Empty(t, args)
		assert.Equal(t, "SELECT * FROM users ORDER BY created_at,updated_at ASC", str)
	})

	t.Run("select with group by", func(t *testing.T) {
		s := Select{}
		s.Table = "users"
		s.GroupBy = &GroupBy{
			Columns: []string{"created_at", "updated_at"},
		}
		str, args, err := s.ToSql()
		assert.NoError(t, err)
		assert.Empty(t, args)
		assert.Equal(t, "SELECT * FROM users GROUP BY created_at,updated_at", str)
	})

	t.Run("Select with limit", func(t *testing.T) {
		s := Select{}
		s.Table = "users"
		s.Limit = &Limit{N: 10}
		str, args, err := s.ToSql()
		assert.NoError(t, err)
		assert.Empty(t, args)
		assert.Equal(t, "SELECT * FROM users LIMIT 10", str)
	})

	t.Run("Select with offset", func(t *testing.T) {
		s := Select{}
		s.Table = "users"
		s.Offset = &Offset{N: 10}
		str, args, err := s.ToSql()
		assert.NoError(t, err)
		assert.Empty(t, args)
		assert.Equal(t, "SELECT * FROM users OFFSET 10", str)
	})

	t.Run("Select with having", func(t *testing.T) {
		s := Select{}
		s.Table = "users"
		s.PlaceholderGenerator = Dialects.MySQL.PlaceHolderGenerator
		s.Having = &Having{Cond: Cond{
			Lhs: "COUNT(id)",
			Op:  LT,
			Rhs: 5,
		}}
		str, args, err := s.ToSql()
		assert.NoError(t, err)
		assert.Equal(t, []interface{}{5}, args)
		assert.Equal(t, "SELECT * FROM users HAVING COUNT(id) < ?", str)
	})

	t.Run("select with join", func(t *testing.T) {
		s := Select{}
		s.Table = "users"
		s.Selected = &Selected{Columns: []string{"id", "name"}}
		s.Joins = append(s.Joins, &Join{
			Type:  JoinTypeRight,
			Table: "addresses",
			On: JoinOn{
				Lhs: "users.id",
				Rhs: "addresses.user_id",
			},
		})
		str, args, err := s.ToSql()
		assert.NoError(t, err)
		assert.Empty(t, args)
		assert.Equal(t, `SELECT id,name FROM users RIGHT JOIN addresses ON users.id = addresses.user_id`, str)
	})

	t.Run("select with multiple joins", func(t *testing.T) {
		s := Select{}
		s.Table = "users"
		s.Selected = &Selected{Columns: []string{"id", "name"}}
		s.Joins = append(s.Joins, &Join{
			Type:  JoinTypeRight,
			Table: "addresses",
			On: JoinOn{
				Lhs: "users.id",
				Rhs: "addresses.user_id",
			},
		}, &Join{
			Type:  JoinTypeLeft,
			Table: "user_credits",
			On: JoinOn{
				Lhs: "users.id",
				Rhs: "user_credits.user_id",
			},
		})
		sql, args, err := s.ToSql()
		assert.NoError(t, err)
		assert.Empty(t, args)
		assert.Equal(t, `SELECT id,name FROM users RIGHT JOIN addresses ON users.id = addresses.user_id LEFT JOIN user_credits ON users.id = user_credits.user_id`, sql)
	})

	t.Run("select with subquery", func(t *testing.T) {
		s := Select{}
		s.PlaceholderGenerator = Dialects.MySQL.PlaceHolderGenerator
		s.SubQuery = &Select{
			Table: "users",
			Where: &Where{Cond: Cond{
				Lhs: "age",
				Op:  LT,
				Rhs: 10,
			}},
		}
		sql, args, err := s.ToSql()
		assert.NoError(t, err)
		assert.EqualValues(t, []interface{}{10}, args)
		assert.Equal(t, `SELECT * FROM (SELECT * FROM users WHERE age < ? )`, sql)

	})
}
