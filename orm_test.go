package orm_test

import (
	"github.com/golobby/orm"
	"github.com/stretchr/testify/assert"

	"testing"
)

type Post struct {
	ID   int64
	Body string
}

func (p Post) Schema() *orm.Schema {
	return &orm.Schema{
		Table: "posts",
	}
}

func (p *Post) Categories() ([]Category, error) {
	return orm.BelongsToMany[Category](p, orm.BelongsToManyConfig{})
}

func (p *Post) Comments() ([]Comment, error) {
	return orm.HasMany[Comment](p, orm.HasManyConfig{})
}

type Comment struct {
	ID   int
	Body string
}

func (c Comment) Schema() *orm.Schema {
	return &orm.Schema{Table: "comments"}
}

func (c *Comment) Post() (Post, error) {
	return orm.BelongsTo[Post](c, orm.BelongsToConfig{})
}

type Category struct {
	ID    int
	Title string
}

func (c Category) Schema() *orm.Schema {
	return &orm.Schema{Table: "categories"}
}

func (c Category) Posts() ([]Post, error) {
	return orm.BelongsToMany[Post](c, orm.BelongsToManyConfig{})
}

// enough models let's test

func setup(t *testing.T) {
	err := orm.Initialize(orm.ConnectionConfig{
		Name:             "default",
		Driver:           "sqlite3",
		ConnectionString: ":memory:",
		Entities:         []orm.Entity{&Comment{}, &Post{}, &Category{}},
	})

	_, err = orm.GetConnection("default").Connection.Exec(`CREATE TABLE IF NOT EXISTS posts (id INTEGER PRIMARY KEY, body text)`)
	_, err = orm.GetConnection("default").Connection.Exec(`CREATE TABLE IF NOT EXISTS comments (id INTEGER PRIMARY KEY, body text)`)
	_, err = orm.GetConnection("default").Connection.Exec(`CREATE TABLE IF NOT EXISTS categories (id INTEGER PRIMARY KEY, title text)`)
	_, err = orm.GetConnection("default").Connection.Exec(`CREATE TABLE IF NOT EXISTS post_categories (post_id INTEGER, category_id INTEGER, PRIMARY KEY(post_id, category_id))`)
	assert.NoError(t, err)

}

func TestFind(t *testing.T) {
	setup(t)
	err := orm.Insert(&Post{
		Body: "my body for insert",
	})
	assert.NoError(t, err)

	post, err := orm.Find[Post](1)
	assert.NoError(t, err)
	assert.Equal(t, "my body for insert", post.Body)
	assert.Equal(t, int64(1), post.ID)
}

func TestInsert(t *testing.T) {
	setup(t)
	post := &Post{
		Body: "my body for insert",
	}
	err := orm.Insert(post)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), post.ID)
	var p Post
	assert.NoError(t,
		orm.GetConnection("default").Connection.QueryRow(`SELECT id, body FROM posts where id = ?`, 1).Scan(&p.ID, &p.Body))

	assert.Equal(t, "my body for insert", p.Body)
}

func TestUpdate(t *testing.T) {
	setup(t)
	post := &Post{
		Body: "my body for insert",
	}
	err := orm.Insert(post)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), post.ID)

	post.Body += " update text"
	assert.NoError(t, orm.Update(post))

	var body string
	assert.NoError(t,
		orm.GetConnection("default").Connection.QueryRow(`SELECT body FROM posts where id = ?`, post.ID).Scan(&body))

	assert.Equal(t, "my body for insert update text", body)
}

func TestDelete(t *testing.T) {
	setup(t)
	post := &Post{
		Body: "my body for insert",
	}
	err := orm.Insert(post)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), post.ID)

	assert.NoError(t, orm.Delete(post))

	var count int
	assert.NoError(t,
		orm.GetConnection("default").Connection.QueryRow(`SELECT count(id) FROM posts where id = ?`, post.ID).Scan(&count))

	assert.Equal(t, 0, count)
}
