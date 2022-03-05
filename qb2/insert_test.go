package qb2

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestInsert(t *testing.T) {
	t.Run("insert into multiple rows", func(t *testing.T) {
		i := Insert{}
		i.Into = "users"
		i.Columns = []string{"name", "age"}
		i.Values = append(i.Values, []string{"amirreza", "11"}, []string{"parsa", "10"})
		s := i.String()
		assert.Equal(t, `INSERT INTO users (name,age) VALUES ('amirreza',11),('parsa',10)`, s)
	})
}
