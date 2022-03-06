package qb2

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInsert(t *testing.T) {
	t.Run("insert into multiple rows", func(t *testing.T) {
		i := Insert{}
		i.Into = "users"
		i.dialect = Dialects.MySQL
		i.Columns = []string{"name", "age"}
		i.Values = append(i.Values, []interface{}{"amirreza", 11}, []interface{}{"parsa", 10})
		s, args := i.ToSql()
		assert.Equal(t, `INSERT INTO users (name,age) VALUES (?,?),(?,?)`, s)
		assert.EqualValues(t, []interface{}{"amirreza", 11, "parsa", 10}, args)
	})

	t.Run("insert into single row", func(t *testing.T) {
		i := Insert{}
		i.Into = "users"
		i.dialect = Dialects.MySQL
		i.Columns = []string{"name", "age"}
		i.Values = append(i.Values, []interface{}{"amirreza", 11})
		s, args := i.ToSql()
		assert.Equal(t, `INSERT INTO users (name,age) VALUES (?,?)`, s)
		assert.Equal(t, []interface{}{"amirreza", 11}, args)
	})
}
