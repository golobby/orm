package qb2

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDelete(t *testing.T) {
	t.Run("delete without where", func(t *testing.T) {
		d := Delete{}
		d.From = "users"
		assert.Equal(t, `DELETE FROM users`, d.String())
	})
	t.Run("delete with where", func(t *testing.T) {
		d := Delete{}
		d.From = "users"
		d.Where = &Where{
			BinaryOp: BinaryOp{
				Lhs: "created_at",
				Op:  ">",
				Rhs: "2012-01-10",
			},
		}
		assert.Equal(t, `DELETE FROM users WHERE created_at > '2012-01-10'`, d.String())
	})
}
