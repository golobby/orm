package orm_test

import (
	"testing"

	"github.com/golobby/orm"
	"github.com/stretchr/testify/assert"
)

type Post struct {
	ID   int64
	Body string
}

func (p Post) ConfigureEntity(e *orm.EntityConfigurator) {
	e.Table("posts")
}

func (p *Post) Categories() ([]Category, error) {
	return orm.BelongsToMany[Category](p, orm.BelongsToManyConfig{})
}

func (p *Post) Comments() ([]Comment, error) {
	return orm.HasMany[Comment](p, orm.HasManyConfig{})
}

type Comment struct {
	ID     int64
	PostID int64
	Body   string
}

func (c Comment) ConfigureEntity(e *orm.EntityConfigurator) {
	e.Table("comments")
}

func (c *Comment) Post() (Post, error) {
	return orm.BelongsTo[Post](c, orm.BelongsToConfig{})
}

type Category struct {
	ID    int64
	Title string
}

func (c Category) ConfigureEntity(e *orm.EntityConfigurator) {
	e.Table("categories")
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
	_, err = orm.GetConnection("default").Connection.Exec(`CREATE TABLE IF NOT EXISTS comments (id INTEGER PRIMARY KEY, post_id INTEGER, body text)`)
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
func TestAdd(t *testing.T) {
	setup(t)
	post := &Post{
		Body: "my body for insert",
	}
	err := orm.Insert(post)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), post.ID)

	err = orm.Add(post, orm.RelationType_HasMany, []Comment{
		{
			PostID: post.ID,
			Body:   "comment 1",
		},
		{
			PostID: post.ID,
			Body:   "comment 2",
		},
	}...)
	assert.NoError(t, err)

}

func TestSave(t *testing.T) {
	t.Run("save should insert", func(t *testing.T) {
		setup(t)
		post := &Post{
			Body: "1",
		}
		assert.NoError(t, orm.Save(post))
		assert.Equal(t, int64(1), post.ID)
	})

	t.Run("save should update", func(t *testing.T) {
		setup(t)
		post := &Post{
			Body: "1",
		}
		assert.NoError(t, orm.Save(post))
		assert.Equal(t, int64(1), post.ID)

		post.Body += "2"
		assert.NoError(t, orm.Save(post))

		myPost, err := orm.Find[Post](1)
		assert.NoError(t, err)

		assert.Equal(t, post, &myPost)
	})

}

func TestHasMany(t *testing.T) {
	setup(t)
	post := &Post{
		Body: "first post",
	}
	assert.NoError(t, orm.Save(post))
	assert.Equal(t, int64(1), post.ID)

	assert.NoError(t, orm.Save(&Comment{
		PostID: post.ID,
		Body:   "comment 1",
	}))
	assert.NoError(t, orm.Save(&Comment{
		PostID: post.ID,
		Body:   "comment 2",
	}))

	comments, err := orm.HasMany[Comment](post, orm.HasManyConfig{})
	assert.NoError(t, err)

	assert.Len(t, comments, 2)

	assert.Equal(t, post.ID, comments[0].PostID)
	assert.Equal(t, post.ID, comments[1].PostID)
}

func TestBelongsTo(t *testing.T) {
	setup(t)
	post := &Post{
		Body: "first post",
	}
	assert.NoError(t, orm.Save(post))
	assert.Equal(t, int64(1), post.ID)

	comment := &Comment{
		PostID: post.ID,
		Body:   "comment 1",
	}
	assert.NoError(t, orm.Save(comment))

	post2, err := orm.BelongsTo[Post](comment, orm.BelongsToConfig{})
	assert.NoError(t, err)

	assert.Equal(t, *post, post2)
}

func TestHasOne(t *testing.T) {
	setup(t)
	post := &Post{
		Body: "first post",
	}
	assert.NoError(t, orm.Save(post))
	assert.Equal(t, int64(1), post.ID)

	comment := &Comment{
		PostID: post.ID,
		Body:   "comment 1",
	}
	assert.NoError(t, orm.Save(comment))

	c1, err := orm.HasOne[Comment](post, orm.HasOneConfig{})
	assert.NoError(t, err)

	assert.Equal(t, comment.PostID, c1.PostID)
}

func TestBelongsToMany(t *testing.T) {
	setup(t)

	post := &Post{
		Body: "first Post",
	}

	assert.NoError(t, orm.Save(post))
	assert.Equal(t, int64(1), post.ID)

	category := &Category{
		Title: "first category",
	}
	assert.NoError(t, orm.Save(category))
	assert.Equal(t, int64(1), category.ID)

	_, _, err := orm.ExecRaw[Category](`INSERT INTO post_categories (post_id, category_id) VALUES (?,?)`, post.ID, category.ID)
	assert.NoError(t, err)

	categories, err := orm.BelongsToMany[Category](post, orm.BelongsToManyConfig{IntermediateTable: "post_categories"})
	assert.NoError(t, err)

	assert.Len(t, categories, 1)
}
