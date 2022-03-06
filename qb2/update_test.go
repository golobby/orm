package qb2

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdate(t *testing.T) {
	t.Run("update no where", func(t *testing.T) {
		u := Update{}
		u.Table = "users"
		u.Dialect = Dialects.MySQL
		u.Set = append(u.Set, updateTuple{
			Key:   "name",
			Value: "amirreza",
		})
		sql, args := u.ToSql()

		assert.Equal(t, `UPDATE users SET name=?`, sql)
		assert.Equal(t, []interface{}{"amirreza"}, args)
	})
	t.Run("update with where", func(t *testing.T) {
		u := Update{}
		u.Table = "users"
		u.Dialect = Dialects.MySQL
		u.Set = append(u.Set, updateTuple{
			Key:   "name",
			Value: "amirreza",
		})
		u.Where = &Where{
			BinaryOp: BinaryOp{
				Lhs: "age",
				Op:  LT,
				Rhs: 18,
			},
		}
		sql, args := u.ToSql()
		assert.Equal(t, `UPDATE users SET name=? WHERE age < ?`, sql)
		assert.Equal(t, []interface{}{"amirreza", 18}, args)

	})
}
