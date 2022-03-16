package orm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func setup(t *testing.T) {
	err := SetupConnection(ConnectionConfig{
		Driver:           "sqlite3",
		ConnectionString: ":memory:",
	})
	// orm.Schematic()
	_, err = GetConnection("default").Connection.Exec(`CREATE TABLE IF NOT EXISTS posts (id INTEGER PRIMARY KEY, body text, created_at TIMESTAMP, updated_at TIMESTAMP, deleted_at TIMESTAMP)`)
	_, err = GetConnection("default").Connection.Exec(`CREATE TABLE IF NOT EXISTS emails (id INTEGER PRIMARY KEY, post_id INTEGER, email text)`)
	_, err = GetConnection("default").Connection.Exec(`CREATE TABLE IF NOT EXISTS header_pictures (id INTEGER PRIMARY KEY, post_id INTEGER, link text)`)
	_, err = GetConnection("default").Connection.Exec(`CREATE TABLE IF NOT EXISTS comments (id INTEGER PRIMARY KEY, post_id INTEGER, body text)`)
	_, err = GetConnection("default").Connection.Exec(`CREATE TABLE IF NOT EXISTS categories (id INTEGER PRIMARY KEY, title text)`)
	_, err = GetConnection("default").Connection.Exec(`CREATE TABLE IF NOT EXISTS post_categories (post_id INTEGER, category_id INTEGER, PRIMARY KEY(post_id, category_id))`)
	assert.NoError(t, err)
}

type Object struct {
	ID   int64
	Name string
	Timestamps
}

func (o Object) ConfigureEntity(e *EntityConfigurator) {
	e.Table("objects").Connection("default")
}

func TestGenericFieldsOf(t *testing.T) {
	t.Run("fields of with id and timestamps embedded", func(t *testing.T) {
		fs := genericFieldsOf(&Object{})
		assert.Len(t, fs, 5)
		assert.Equal(t, "id", fs[0].Name)
		assert.True(t, fs[0].IsPK)
		assert.Equal(t, "name", fs[1].Name)
		assert.Equal(t, "created_at", fs[2].Name)
		assert.Equal(t, "updated_at", fs[3].Name)
		assert.Equal(t, "deleted_at", fs[4].Name)
	})
}

func TestGenericValuesOf(t *testing.T) {
	t.Run("values of", func(t *testing.T) {

		setup(t)
		vs := genericValuesOf(Object{}, true)
		assert.Len(t, vs, 5)
	})
}

func TestEntityConfigurator(t *testing.T) {
	t.Run("test has many with user provided values", func(t *testing.T) {
		setup(t)
		var ec EntityConfigurator
		ec.Table("users").Connection("default").HasMany(Object{}, HasManyConfig{
			"objects", "user_id",
		})

	})

}
