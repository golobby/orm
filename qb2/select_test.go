package qb2

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSelect(t *testing.T) {
	t.Run("only select * from table", func(t *testing.T) {
		s := Select{}
		s.Table = "users"
		str, err := s.String()
		assert.NoError(t, err)
		assert.Equal(t, "SELECT * FROM users", str)
	})
	t.Run("select with where", func(t *testing.T) {
		s := Select{}
		s.Table = "users"
		s.Where = &Where{
			BinaryOp: BinaryOp{
				Lhs: "age",
				Op:  Eq,
				Rhs: 10,
			},
		}
		str, err := s.String()
		assert.NoError(t, err)
		assert.Equal(t, "SELECT * FROM users WHERE age = 10", str)
	})
	t.Run("select with order by", func(t *testing.T) {
		s := Select{}
		s.Table = "users"
		s.OrderBy = &OrderBy{
			Columns: []string{"created_at", "updated_at"},
			Order:   OrderByASC,
		}
		str, err := s.String()
		assert.NoError(t, err)
		assert.Equal(t, "SELECT * FROM users ORDER BY created_at,updated_at ASC", str)
	})

	t.Run("select with group by", func(t *testing.T) {
		s := Select{}
		s.Table = "users"
		s.GroupBy = &GroupBy{
			Columns: []string{"created_at", "updated_at"},
		}
		str, err := s.String()
		assert.NoError(t, err)
		assert.Equal(t, "SELECT * FROM users GROUP BY created_at,updated_at", str)
	})

	t.Run("Select with limit", func(t *testing.T) {
		s := Select{}
		s.Table = "users"
		s.Limit = &Limit{N: 10}
		str, err := s.String()
		assert.NoError(t, err)
		assert.Equal(t, "SELECT * FROM users LIMIT 10", str)
	})

	t.Run("Select with offset", func(t *testing.T) {
		s := Select{}
		s.Table = "users"
		s.Offset = &Offset{N: 10}
		str, err := s.String()
		assert.NoError(t, err)
		assert.Equal(t, "SELECT * FROM users OFFSET 10", str)
	})

	t.Run("Select with having", func(t *testing.T) {
		s := Select{}
		s.Table = "users"
		s.Having = &Having{Cond: BinaryOp{
			Lhs: "COUNT(id)",
			Op:  LT,
			Rhs: 5,
		}}
		str, err := s.String()
		assert.NoError(t, err)
		assert.Equal(t, "SELECT * FROM users HAVING COUNT(id) < 5", str)
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
		sql, err := s.String()
		assert.NoError(t, err)
		assert.Equal(t, `SELECT id,name FROM users RIGHT JOIN addresses ON users.id = addresses.user_id`, sql)
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
		sql, err := s.String()
		assert.NoError(t, err)
		assert.Equal(t, `SELECT id,name FROM users RIGHT JOIN addresses ON users.id = addresses.user_id LEFT JOIN user_credits ON users.id = user_credits.user_id`, sql)
	})
}
