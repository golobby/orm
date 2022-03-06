package qb2

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpdate(t *testing.T) {
	t.Run("update no where", func(t *testing.T) {
		u := Update{}
		u.Table = "users"
		u.Set = append(u.Set, updateTuple{
			Key:   "name",
			Value: "amirreza",
		})
		assert.Equal(t, `UPDATE users SET name='amirreza'`, u.String())
	})
	t.Run("update with where", func(t *testing.T) {
		u := Update{}
		u.Table = "users"
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
		assert.Equal(t, `UPDATE users SET name='amirreza' WHERE age < 18`, u.String())

	})
}
