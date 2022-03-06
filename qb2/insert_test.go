package qb2

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInsert(t *testing.T) {
	t.Run("insert into multiple rows", func(t *testing.T) {
		i := Insert{}
		i.Into = "users"
		i.Columns = []string{"name", "age"}
		i.Values = append(i.Values, []interface{}{"amirreza", 11}, []interface{}{"parsa", 10})
		s := i.String()
		assert.Equal(t, `INSERT INTO users (name,age) VALUES ('amirreza',11),('parsa',10)`, s)
	})

	t.Run("insert into single row", func(t *testing.T) {
		i := Insert{}
		i.Into = "users"
		i.Columns = []string{"name", "age"}
		i.Values = append(i.Values, []interface{}{"amirreza", 11})
		s := i.String()
		assert.Equal(t, `INSERT INTO users (name,age) VALUES ('amirreza',11)`, s)
	})
}
