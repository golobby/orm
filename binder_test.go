package orm

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

type User struct {
	ID   int
	Name string
}
type ComplexUser struct {
	ID      int
	Name    string
	Address Address `orm:"in_rel=true with=addresses left=id right=user_id"`
}
func (c ComplexUser) Table() string {
	return "users"
}

type Address struct {
	ID   int
	Path string
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
		md := ObjectMetadataFrom(u, Sqlite3SQLDialect)
		err = md.Bind(rows, u)
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

		md := ObjectMetadataFrom(&User{}, Sqlite3SQLDialect)
		var users []*User
		err = md.Bind(rows, &users)
		assert.NoError(t, err)

		assert.Equal(t, "amirreza", users[0].Name)
		assert.Equal(t, "milad", users[1].Name)
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
		md := ObjectMetadataFrom(u, Sqlite3SQLDialect)
		err = md.Bind(rows, u)
		assert.NoError(t, err)

		assert.Equal(t, "amirreza", u.Name)
		assert.Equal(t, "kianpars", u.Address.Path)
	})
	t.Run("multi result", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		assert.NoError(t, err)
		mock.
			ExpectQuery("SELECT .* FROM users").
			WillReturnRows(sqlmock.NewRows([]string{"users.id", "users.name", "addresses.id", "addresses.path"}).
				AddRow(1, "amirreza", 1, "kianpars").
				AddRow(2, "milad", 2, "delfan"))

		rows, err := db.Query(`SELECT users.id, users.name, addresses.id, addresses.path FROM users INNER JOIN addresses ON addresses.user_id = users.id`)
		assert.NoError(t, err)

		md := ObjectMetadataFrom(&ComplexUser{}, Sqlite3SQLDialect)
		var users []*ComplexUser
		err = md.Bind(rows, &users)
		assert.NoError(t, err)

		assert.Equal(t, "amirreza", users[0].Name)
		assert.Equal(t, "milad", users[1].Name)
		assert.Equal(t, "kianpars", users[0].Address.Path)
		assert.Equal(t, "delfan", users[1].Address.Path)
		assert.Equal(t, 2, users[1].Address.ID)
		assert.Equal(t, 1, users[0].Address.ID)
	})
}