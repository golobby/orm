package qb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDelete(t *testing.T) {
	t.Run("delete without WhereClause", func(t *testing.T) {
		d := Delete{}
		d.Table = "users"
		sql, args := d.ToSql()
		assert.Equal(t, `DELETE FROM users`, sql)
		assert.Empty(t, args)
	})
	t.Run("delete with WhereClause", func(t *testing.T) {
		d := Delete{}
		d.Table = "users"
		d.PlaceHolderGenerator = Dialects.MySQL.PlaceHolderGenerator
		d.Where = &Where{
			Cond: Cond{
				Lhs: "created_at",
				Op:  ">",
				Rhs: "2012-01-10",
			},
		}
		sql, args := d.ToSql()
		assert.Equal(t, `DELETE FROM users WHERE created_at > ?`, sql)
		assert.EqualValues(t, []interface{}{"2012-01-10"}, args)
	})
}
