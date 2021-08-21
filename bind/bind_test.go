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

func TestBind(t *testing.T) {
	db, err := sql.Open("postgres", "host=127.0.0.1 port=5432 user=connect password=connect dbname=connect sslmode=disable")
	assert.NoError(t, err)
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS users (id SERIAL PRIMARY KEY, name VARCHAR(255))")

	assert.NoError(t, err)
	_, err = db.Exec("INSERT INTO users (name) VALUES ('amirreza')")

	assert.NoError(t, err)
	rows, err := db.Query(`SELECT * FROM users`)
	assert.NoError(t, err)

	u := &User{}
	err = Bind(rows, u)
	assert.NoError(t, err)

	assert.Equal(t, "amirreza", u.Name)
}
