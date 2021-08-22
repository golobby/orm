package bind

import (
	"database/sql"
	"testing"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

type User struct {
	ID   int    `bind:"id"`
	Name string `bind:"name"`
}

type ComplexUser struct {
	ID      int     `bind:"id"`
	Name    string  `bind:"name"`
	Address Address
}

type Address struct {
	ID   int    `bind:"id"`
	Path string `bind:"path"`
}

func TestBind(t *testing.T) {
	t.Run("single result", func(t *testing.T) {
		db, err := sql.Open("postgres", "host=127.0.0.1 port=5432 user=connect password=connect dbname=connect sslmode=disable")
		assert.NoError(t, err)

		defer db.Close()

		_, err = db.Exec("DROP TABLE IF EXISTS users")
		assert.NoError(t, err)

		_, err = db.Exec("DROP TABLE IF EXISTS addresses")
		assert.NoError(t, err)

		_, err = db.Exec("CREATE TABLE IF NOT EXISTS users (id SERIAL PRIMARY KEY, name VARCHAR(255))")
		assert.NoError(t, err)

		_, err = db.Exec("DELETE FROM users")
		assert.NoError(t, err)

		_, err = db.Exec("INSERT INTO users (name) VALUES ('amirreza')")
		assert.NoError(t, err)

		rows, err := db.Query(`SELECT * FROM users`)
		assert.NoError(t, err)

		u := &User{}

		err = Bind(rows, u)
		assert.NoError(t, err)

		assert.Equal(t, "amirreza", u.Name)
	})

	t.Run("multi result", func(t *testing.T) {
		db, err := sql.Open("postgres", "host=127.0.0.1 port=5432 user=connect password=connect dbname=connect sslmode=disable")
		assert.NoError(t, err)

		defer db.Close()

		_, err = db.Exec("DROP TABLE IF EXISTS users")
		assert.NoError(t, err)

		_, err = db.Exec("DROP TABLE IF EXISTS addresses")
		assert.NoError(t, err)

		_, err = db.Exec("CREATE TABLE IF NOT EXISTS users (id SERIAL PRIMARY KEY, name VARCHAR(255))")
		assert.NoError(t, err)

		_, err = db.Exec("DELETE FROM users")
		assert.NoError(t, err)

		_, err = db.Exec("INSERT INTO users (name) VALUES ('amirreza')")
		assert.NoError(t, err)

		_, err = db.Exec("INSERT INTO users (name) VALUES ('milad')")
		assert.NoError(t, err)

		assert.NoError(t, err)
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
		db, err := sql.Open("postgres", "host=127.0.0.1 port=5432 user=connect password=connect dbname=connect sslmode=disable")
		assert.NoError(t, err)

		defer db.Close()

		_, err = db.Exec("DROP TABLE IF EXISTS users")
		assert.NoError(t, err)

		_, err = db.Exec("DROP TABLE IF EXISTS addresses")
		assert.NoError(t, err)

		_, err = db.Exec("CREATE TABLE IF NOT EXISTS users (id SERIAL PRIMARY KEY, name VARCHAR(255))")
		assert.NoError(t, err)

		_, err = db.Exec("CREATE TABLE IF NOT EXISTS addresses (id SERIAL PRIMARY KEY, path VARCHAR(255), user_id int)")
		assert.NoError(t, err)

		_, err = db.Exec("DELETE FROM users")
		assert.NoError(t, err)

		_, err = db.Exec("DELETE FROM addresses")
		assert.NoError(t, err)

		_, err = db.Exec("INSERT INTO users (name) VALUES ('amirreza')")
		assert.NoError(t, err)

		_, err = db.Exec("INSERT INTO addresses (path, user_id) VALUES ('kianpars', 1)")
		assert.NoError(t, err)

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

}