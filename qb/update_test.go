package qb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdate(t *testing.T) {
	t.Run("update no where", func(t *testing.T) {
		u := Update{}
		u.Table = "users"
		u.PlaceHolderGenerator = Dialects.MySQL.PlaceHolderGenerator
		u.Set = append(u.Set, [2]interface{}{
			"name",
			"amirreza",
		})
		sql, args := u.ToSql()

		assert.Equal(t, `UPDATE users SET name=?`, sql)
		assert.Equal(t, []interface{}{"amirreza"}, args)
	})
	t.Run("update with where", func(t *testing.T) {
		u := Update{}
		u.Table = "users"
		u.PlaceHolderGenerator = Dialects.MySQL.PlaceHolderGenerator
		u.Set = append(u.Set, [2]interface{}{
			"name",
			"amirreza",
		})
		u.Where = &Where{
			Cond: Cond{
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
