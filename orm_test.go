package orm_test

import (
	"database/sql"
	"testing"

	"github.com/golobby/orm"
	"github.com/stretchr/testify/assert"
)

type AuthorEmail struct {
	ID    int64
	Email string
}

func (a AuthorEmail) ConfigureEntity(e *orm.EntityConfigurator) {
	e.
		Table("emails").
		Connection("default").
		BelongsTo(&Post{}, orm.BelongsToConfig{})
}

type HeaderPicture struct {
	ID     int64
	PostID int64
	Link   string
}

func (h HeaderPicture) ConfigureEntity(e *orm.EntityConfigurator) {
	e.Table("header_pictures").BelongsTo(&Post{}, orm.BelongsToConfig{})
}

type Post struct {
	ID        int64
	BodyText  string
	CreatedAt sql.NullTime
	UpdatedAt sql.NullTime
	DeletedAt sql.NullTime
}

func (p Post) ConfigureEntity(e *orm.EntityConfigurator) {
	e.Field("BodyText").ColumnName("body")
	e.Field("ID").ColumnName("id")
	e.
		Table("posts").
		HasMany(Comment{}, orm.HasManyConfig{}).
		HasOne(HeaderPicture{}, orm.HasOneConfig{}).
		HasOne(AuthorEmail{}, orm.HasOneConfig{}).
		BelongsToMany(Category{}, orm.BelongsToManyConfig{IntermediateTable: "post_categories"})

}

func (p *Post) Categories() ([]Category, error) {
	return orm.BelongsToMany[Category](p).All()
}

func (p *Post) Comments() *orm.QueryBuilder[Comment] {
	return orm.HasMany[Comment](p)
}

type Comment struct {
	ID     int64
	PostID int64
	Body   string
}

func (c Comment) ConfigureEntity(e *orm.EntityConfigurator) {
	e.Table("comments").BelongsTo(&Post{}, orm.BelongsToConfig{})
}

func (c *Comment) Post() (Post, error) {
	return orm.BelongsTo[Post](c).Get()
}

type Category struct {
	ID    int64
	Title string
}

func (c Category) ConfigureEntity(e *orm.EntityConfigurator) {
	e.Table("categories").BelongsToMany(Post{}, orm.BelongsToManyConfig{IntermediateTable: "post_categories"})
}

func (c Category) Posts() ([]Post, error) {
	return orm.BelongsToMany[Post](c).All()
}

// enough models let's test
// Entities is mandatory
// Errors should be carried

func setup() error {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		return err
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS posts (id INTEGER PRIMARY KEY, body text, created_at TIMESTAMP, updated_at TIMESTAMP, deleted_at TIMESTAMP)`)
	if err != nil {
		return err
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS emails (id INTEGER PRIMARY KEY, post_id INTEGER, email text)`)
	if err != nil {
		return err
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS header_pictures (id INTEGER PRIMARY KEY, post_id INTEGER, link text)`)
	if err != nil {
		return err
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS comments (id INTEGER PRIMARY KEY, post_id INTEGER, body text)`)
	if err != nil {
		return err
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS categories (id INTEGER PRIMARY KEY, title text)`)
	if err != nil {
		return err
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS post_categories (post_id INTEGER, category_id INTEGER, PRIMARY KEY(post_id, category_id))`)
	if err != nil {
		return err
	}
	return orm.SetupConnections(orm.ConnectionConfig{
		Name:                "default",
		DB:                  db,
		Dialect:             orm.Dialects.SQLite3,
		Entities:            []orm.Entity{&Post{}, &Comment{}, &Category{}, &HeaderPicture{}},
		DatabaseValidations: true,
	})
}

func TestFind(t *testing.T) {
	err := setup()
	assert.NoError(t, err)
	err = orm.InsertAll(&Post{
		BodyText: "my body for insert",
	})

	assert.NoError(t, err)

	post, err := orm.Find[Post](1)
	assert.NoError(t, err)
	assert.Equal(t, "my body for insert", post.BodyText)
	assert.Equal(t, int64(1), post.ID)
}

func TestInsert(t *testing.T) {
	err := setup()
	assert.NoError(t, err)
	post := &Post{
		BodyText: "my body for insert",
	}
	err = orm.Insert(post)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), post.ID)
	var p Post
	assert.NoError(t,
		orm.GetConnection("default").DB.QueryRow(`SELECT id, body FROM posts where id = ?`, 1).Scan(&p.ID, &p.BodyText))

	assert.Equal(t, "my body for insert", p.BodyText)
}
func TestInsertAll(t *testing.T) {
	err := setup()
	assert.NoError(t, err)
	post1 := &Post{
		BodyText: "Body1",
	}
	post2 := &Post{
		BodyText: "Body2",
	}

	post3 := &Post{
		BodyText: "Body3",
	}

	err = orm.InsertAll(post1, post2, post3)
	assert.NoError(t, err)
	var counter int
	assert.NoError(t, orm.GetConnection("default").DB.QueryRow(`SELECT count(id) FROM posts`).Scan(&counter))
	assert.Equal(t, 3, counter)

}
func TestUpdateORM(t *testing.T) {
	err := setup()
	assert.NoError(t, err)
	post := &Post{
		BodyText: "my body for insert",
	}
	err = orm.Insert(post)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), post.ID)

	post.BodyText += " update text"
	assert.NoError(t, orm.Update(post))

	var body string
	assert.NoError(t,
		orm.GetConnection("default").DB.QueryRow(`SELECT body FROM posts where id = ?`, post.ID).Scan(&body))

	assert.Equal(t, "my body for insert update text", body)
}

func TestDeleteORM(t *testing.T) {
	err := setup()
	assert.NoError(t, err)
	post := &Post{
		BodyText: "my body for insert",
	}
	err = orm.Insert(post)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), post.ID)

	assert.NoError(t, orm.Delete(post))

	var count int
	assert.NoError(t,
		orm.GetConnection("default").DB.QueryRow(`SELECT count(id) FROM posts where id = ?`, post.ID).Scan(&count))

	assert.Equal(t, 0, count)
}
func TestAdd_HasMany(t *testing.T) {
	err := setup()
	assert.NoError(t, err)
	post := &Post{
		BodyText: "my body for insert",
	}
	err = orm.Insert(post)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), post.ID)

	err = orm.Add(post, []orm.Entity{
		Comment{
			Body: "comment 1",
		},
		Comment{
			Body: "comment 2",
		},
	}...)
	// orm.Query(qm.WhereBetween())
	assert.NoError(t, err)
	var count int
	assert.NoError(t, orm.GetConnection("default").DB.QueryRow(`SELECT COUNT(id) FROM comments`).Scan(&count))
	assert.Equal(t, 2, count)

	comment, err := orm.Find[Comment](1)
	assert.NoError(t, err)

	assert.Equal(t, int64(1), comment.PostID)

}

func TestAdd_ManyToMany(t *testing.T) {
	err := setup()
	assert.NoError(t, err)
	post := &Post{
		BodyText: "my body for insert",
	}
	err = orm.Insert(post)
	assert.NoError(t, err)
	assert.Equal(t, int64(1), post.ID)

	err = orm.Add(post, []orm.Entity{
		&Category{
			Title: "cat 1",
		},
		&Category{
			Title: "cat 2",
		},
	}...)
	assert.NoError(t, err)
	var count int
	assert.NoError(t, orm.GetConnection("default").DB.QueryRow(`SELECT COUNT(post_id) FROM post_categories`).Scan(&count))
	assert.Equal(t, 2, count)
	assert.NoError(t, orm.GetConnection("default").DB.QueryRow(`SELECT COUNT(id) FROM categories`).Scan(&count))
	assert.Equal(t, 2, count)

	categories, err := post.Categories()
	assert.NoError(t, err)

	assert.Equal(t, 2, len(categories))
	assert.Equal(t, int64(1), categories[0].ID)
	assert.Equal(t, int64(2), categories[1].ID)

}

func TestSave(t *testing.T) {
	t.Run("save should insert", func(t *testing.T) {
		err := setup()
		assert.NoError(t, err)
		post := &Post{
			BodyText: "1",
		}
		assert.NoError(t, orm.Save(post))
		assert.Equal(t, int64(1), post.ID)
	})

	t.Run("save should update", func(t *testing.T) {
		err := setup()
		assert.NoError(t, err)

		post := &Post{
			BodyText: "1",
		}
		assert.NoError(t, orm.Save(post))
		assert.Equal(t, int64(1), post.ID)

		post.BodyText += "2"
		assert.NoError(t, orm.Save(post))

		myPost, err := orm.Find[Post](1)
		assert.NoError(t, err)

		assert.EqualValues(t, post.BodyText, myPost.BodyText)
	})

}

func TestHasMany(t *testing.T) {
	err := setup()
	assert.NoError(t, err)

	post := &Post{
		BodyText: "first post",
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

	comments, err := orm.HasMany[Comment](post).All()
	assert.NoError(t, err)

	assert.Len(t, comments, 2)

	assert.Equal(t, post.ID, comments[0].PostID)
	assert.Equal(t, post.ID, comments[1].PostID)
}

func TestBelongsTo(t *testing.T) {
	err := setup()
	assert.NoError(t, err)

	post := &Post{
		BodyText: "first post",
	}
	assert.NoError(t, orm.Save(post))
	assert.Equal(t, int64(1), post.ID)

	comment := &Comment{
		PostID: post.ID,
		Body:   "comment 1",
	}
	assert.NoError(t, orm.Save(comment))

	post2, err := orm.BelongsTo[Post](comment).Get()
	assert.NoError(t, err)

	assert.Equal(t, post.BodyText, post2.BodyText)
}

func TestHasOne(t *testing.T) {
	err := setup()
	assert.NoError(t, err)

	post := &Post{
		BodyText: "first post",
	}
	assert.NoError(t, orm.Save(post))
	assert.Equal(t, int64(1), post.ID)

	headerPicture := &HeaderPicture{
		PostID: post.ID,
		Link:   "google",
	}
	assert.NoError(t, orm.Save(headerPicture))

	c1, err := orm.HasOne[HeaderPicture](post).Get()
	assert.NoError(t, err)

	assert.Equal(t, headerPicture.PostID, c1.PostID)
}

func TestBelongsToMany(t *testing.T) {
	err := setup()
	assert.NoError(t, err)

	post := &Post{
		BodyText: "first Post",
	}

	assert.NoError(t, orm.Save(post))
	assert.Equal(t, int64(1), post.ID)

	category := &Category{
		Title: "first category",
	}
	assert.NoError(t, orm.Save(category))
	assert.Equal(t, int64(1), category.ID)

	_, _, err = orm.ExecRaw[Category](`INSERT INTO post_categories (post_id, category_id) VALUES (?,?)`, post.ID, category.ID)
	assert.NoError(t, err)

	categories, err := orm.BelongsToMany[Category](post).All()
	assert.NoError(t, err)

	assert.Len(t, categories, 1)
}

func TestSchematic(t *testing.T) {
	err := setup()
	assert.NoError(t, err)

	orm.Schematic()
}

func TestAddProperty(t *testing.T) {
	t.Run("having pk value", func(t *testing.T) {
		err := setup()
		assert.NoError(t, err)

		post := &Post{
			BodyText: "first post",
		}

		assert.NoError(t, orm.Save(post))
		assert.EqualValues(t, 1, post.ID)

		err = orm.Add(post, &Comment{PostID: post.ID, Body: "firstComment"})
		assert.NoError(t, err)

		var comment Comment
		assert.NoError(t, orm.GetConnection("default").
			DB.
			QueryRow(`SELECT id, post_id, body FROM comments WHERE post_id=?`, post.ID).
			Scan(&comment.ID, &comment.PostID, &comment.Body))

		assert.EqualValues(t, post.ID, comment.PostID)
	})
	t.Run("not having PK value", func(t *testing.T) {
		err := setup()
		assert.NoError(t, err)

		post := &Post{
			BodyText: "first post",
		}
		assert.NoError(t, orm.Save(post))
		assert.EqualValues(t, 1, post.ID)

		err = orm.Add(post, &AuthorEmail{Email: "myemail"})
		assert.NoError(t, err)

		emails, err := orm.QueryRaw[AuthorEmail](`SELECT id, email FROM emails WHERE post_id=?`, post.ID)

		assert.NoError(t, err)
		assert.Equal(t, []AuthorEmail{{ID: 1, Email: "myemail"}}, emails)
	})
}

func TestQuery(t *testing.T) {
	t.Run("querying single row", func(t *testing.T) {
		err := setup()
		assert.NoError(t, err)

		assert.NoError(t, orm.Save(&Post{BodyText: "body 1"}))
		// post, err := orm.Query[Post]().Where("id", 1).First()
		post, err := orm.Query[Post]().WherePK(1).First().Get()
		assert.NoError(t, err)
		assert.EqualValues(t, "body 1", post.BodyText)
		assert.EqualValues(t, 1, post.ID)

	})
	t.Run("querying multiple rows", func(t *testing.T) {
		err := setup()
		assert.NoError(t, err)

		assert.NoError(t, orm.Save(&Post{BodyText: "body 1"}))
		assert.NoError(t, orm.Save(&Post{BodyText: "body 2"}))
		assert.NoError(t, orm.Save(&Post{BodyText: "body 3"}))
		posts, err := orm.Query[Post]().All()
		assert.NoError(t, err)
		assert.Len(t, posts, 3)
		assert.Equal(t, "body 1", posts[0].BodyText)
	})

	t.Run("updating a row using query interface", func(t *testing.T) {
		err := setup()
		assert.NoError(t, err)

		assert.NoError(t, orm.Save(&Post{BodyText: "body 1"}))

		affected, err := orm.Query[Post]().Where("id", 1).Set("body", "body jadid").Update()
		assert.NoError(t, err)
		assert.EqualValues(t, 1, affected)

		post, err := orm.Find[Post](1)
		assert.NoError(t, err)
		assert.Equal(t, "body jadid", post.BodyText)
	})

	t.Run("deleting a row using query interface", func(t *testing.T) {
		err := setup()
		assert.NoError(t, err)

		assert.NoError(t, orm.Save(&Post{BodyText: "body 1"}))

		affected, err := orm.Query[Post]().WherePK(1).Delete()
		assert.NoError(t, err)
		assert.EqualValues(t, 1, affected)
		count, err := orm.Query[Post]().WherePK(1).Count().Get()
		assert.NoError(t, err)
		assert.EqualValues(t, 0, count)
	})

	t.Run("count", func(t *testing.T) {
		err := setup()
		assert.NoError(t, err)

		count, err := orm.Query[Post]().WherePK(1).Count().Get()
		assert.NoError(t, err)
		assert.EqualValues(t, 0, count)
	})

	t.Run("latest", func(t *testing.T) {
		err := setup()
		assert.NoError(t, err)

		assert.NoError(t, orm.Save(&Post{BodyText: "body 1"}))
		assert.NoError(t, orm.Save(&Post{BodyText: "body 2"}))

		post, err := orm.Query[Post]().Latest().Get()
		assert.NoError(t, err)

		assert.EqualValues(t, "body 2", post.BodyText)
	})
}

func TestSetup(t *testing.T) {
	t.Run("tables are out of sync", func(t *testing.T) {
		db, err := sql.Open("sqlite3", ":memory:")
		// _, err = db.Exec(`CREATE TABLE IF NOT EXISTS posts (id INTEGER PRIMARY KEY, body text, created_at TIMESTAMP, updated_at TIMESTAMP, deleted_at TIMESTAMP)`)
		_, err = db.Exec(`CREATE TABLE IF NOT EXISTS emails (id INTEGER PRIMARY KEY, post_id INTEGER, email text)`)
		_, err = db.Exec(`CREATE TABLE IF NOT EXISTS header_pictures (id INTEGER PRIMARY KEY, post_id INTEGER, link text)`)
		_, err = db.Exec(`CREATE TABLE IF NOT EXISTS comments (id INTEGER PRIMARY KEY, post_id INTEGER, body text)`)
		_, err = db.Exec(`CREATE TABLE IF NOT EXISTS categories (id INTEGER PRIMARY KEY, title text)`)
		// _, err = db.Exec(`CREATE TABLE IF NOT EXISTS post_categories (post_id INTEGER, category_id INTEGER, PRIMARY KEY(post_id, category_id))`)

		err = orm.SetupConnections(orm.ConnectionConfig{
			Name:                "default",
			DB:                  db,
			Dialect:             orm.Dialects.SQLite3,
			Entities:            []orm.Entity{&Post{}, &Comment{}, &Category{}, &HeaderPicture{}},
			DatabaseValidations: true,
		})
		assert.Error(t, err)

	})
	t.Run("schemas are wrong", func(t *testing.T) {
		db, err := sql.Open("sqlite3", ":memory:")
		_, err = db.Exec(`CREATE TABLE IF NOT EXISTS posts (id INTEGER PRIMARY KEY, body text, created_at TIMESTAMP, updated_at TIMESTAMP, deleted_at TIMESTAMP)`)
		_, err = db.Exec(`CREATE TABLE IF NOT EXISTS emails (id INTEGER PRIMARY KEY, post_id INTEGER, email text)`)
		_, err = db.Exec(`CREATE TABLE IF NOT EXISTS header_pictures (id INTEGER PRIMARY KEY, post_id INTEGER, link text)`)
		_, err = db.Exec(`CREATE TABLE IF NOT EXISTS comments (id INTEGER PRIMARY KEY, body text)`) // missing post_id
		_, err = db.Exec(`CREATE TABLE IF NOT EXISTS categories (id INTEGER PRIMARY KEY, title text)`)
		_, err = db.Exec(`CREATE TABLE IF NOT EXISTS post_categories (post_id INTEGER, category_id INTEGER, PRIMARY KEY(post_id, category_id))`)

		err = orm.SetupConnections(orm.ConnectionConfig{
			Name:                "default",
			DB:                  db,
			Dialect:             orm.Dialects.SQLite3,
			Entities:            []orm.Entity{&Post{}, &Comment{}, &Category{}, &HeaderPicture{}},
			DatabaseValidations: true,
		})
		assert.Error(t, err)

	})
}
