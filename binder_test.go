package orm

import (
	"github.com/DATA-DOG/go-sqlmock"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

type User struct {
	ID   int    `bind:"id"`
	Name string `bind:"name"`
}

type ComplexUser struct {
	ID      int    `bind:"id"`
	Name    string `bind:"name"`
	Address Address
}

type Address struct {
	ID   int    `bind:"id"`
	Path string `bind:"path"`
}

func TestBind(t *testing.T) {
	t.Run("single result", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		mock.
			ExpectQuery("SELECT .* FROM users").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "amirreza"))
		rows, err := db.Query(`SELECT * FROM users`)
		assert.NoError(t, err)

		u := &User{}

		err = Bind(rows, u)
		assert.NoError(t, err)

		assert.Equal(t, "amirreza", u.Name)
	})

	t.Run("multi result", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		mock.
			ExpectQuery("SELECT .* FROM users").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow(1, "amirreza").AddRow(2, "milad"))

		rows, err := db.Query(`SELECT * FROM users`)
		assert.NoError(t, err)

		amirreza := &User{}
		milad := &User{}

		err = Bind(rows, []interface{}{amirreza, milad})
		assert.NoError(t, err)

		assert.Equal(t, "amirreza", amirreza.Name)
		assert.Equal(t, "milad", milad.Name)
	})
}

func TestBindNested(t *testing.T) {
	t.Run("single result", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		mock.
			ExpectQuery("SELECT .* FROM users").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "path"}).AddRow(1, "amirreza", "kianpars"))

		rows, err := db.Query(`SELECT users.id, users.name, addresses.path FROM users INNER JOIN addresses ON addresses.user_id = users.id`)
		assert.NoError(t, err)

		u := &ComplexUser{
			Address: Address{},
		}

		err = Bind(rows, u)
		assert.NoError(t, err)

		assert.Equal(t, "amirreza", u.Name)
		assert.Equal(t, "kianpars", u.Address.Path)
	})
	t.Run("multi result", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		mock.
			ExpectQuery("SELECT .* FROM users").
			WillReturnRows(sqlmock.NewRows([]string{"id", "name", "address_id", "address_path"}).
				AddRow(1, "amirreza", 1, "kianpars").
				AddRow(2, "milad", 2, "delfan"))

		rows, err := db.Query(`SELECT users.id, users.name, addresses.id AS address_id, addresses.path AS address_path FROM users INNER JOIN addresses ON addresses.user_id = users.id`)
		assert.NoError(t, err)

		amirreza := &ComplexUser{}
		milad := &ComplexUser{}

		err = Bind(rows, []*ComplexUser{amirreza, milad})
		assert.NoError(t, err)

		assert.Equal(t, "amirreza", amirreza.Name)
		assert.Equal(t, "milad", milad.Name)
		assert.Equal(t, "kianpars", amirreza.Address.Path)
		assert.Equal(t, "delfan", milad.Address.Path)
		assert.Equal(t, 2, milad.Address.ID)
		assert.Equal(t, 1, amirreza.Address.ID)
	})
}
