package orm_test

import (
	"testing"

	"github.com/golobby/orm"
	"github.com/stretchr/testify/assert"
)

type AuthorEmail struct {
	ID    int64
	Email string
}

func (a AuthorEmail) ConfigureEntity(e *orm.EntityConfigurator) {
	e.Table("emails").BelongsTo(&Post{}, orm.BelongsToConfig{})
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
	ID   int64
	Body string
	orm.Timestamps
}

func (p Post) ConfigureEntity(e *orm.EntityConfigurator) {
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

func (p *Post) Comments() ([]Comment, error) {
	return orm.HasMany[Comment](p).All()
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
	return orm.BelongsTo[Post](c).One()
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

func setup(t *testing.T) {
	err := orm.Initialize(orm.ConnectionConfig{
		Driver:           "sqlite3",
		ConnectionString: ":memory:",
	})
	//orm.Schematic()
	_, err = orm.GetConnection("default").Connection.Exec(`CREATE TABLE IF NOT EXISTS posts (id INTEGER PRIMARY KEY, body text, created_at TIMESTAMP, updated_at TIMESTAMP, deleted_at TIMESTAMP)`)
	_, err = orm.GetConnection("default").Connection.Exec(`CREATE TABLE IF NOT EXISTS emails (id INTEGER PRIMARY KEY, post_id INTEGER, email text)`)
	_, err = orm.GetConnection("default").Connection.Exec(`CREATE TABLE IF NOT EXISTS header_pictures (id INTEGER PRIMARY KEY, post_id INTEGER, link text)`)
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
func TestInsertAll(t *testing.T) {
	setup(t)

	post1 := &Post{
		Body: "Body1",
	}
	post2 := &Post{
		Body: "Body2",
	}

	post3 := &Post{
		Body: "Body3",
	}

	err := orm.Insert(post1, post2, post3)
	assert.NoError(t, err)
	var counter int
	assert.NoError(t, orm.GetConnection("default").Connection.QueryRow(`SELECT count(id) FROM posts`).Scan(&counter))
	assert.Equal(t, 3, counter)

}
func TestUpdateORM(t *testing.T) {
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

func TestDeleteORM(t *testing.T) {
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
	assert.NoError(t, orm.GetConnection("default").Connection.QueryRow(`SELECT COUNT(id) FROM comments`).Scan(&count))
	assert.Equal(t, 2, count)

	comment, err := orm.Find[Comment](1)
	assert.NoError(t, err)

	assert.Equal(t, int64(1), comment.PostID)

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

		assert.EqualValues(t, post.Body, myPost.Body)
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

	comments, err := orm.HasMany[Comment](post).All()
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

	post2, err := orm.BelongsTo[Post](comment).One()
	assert.NoError(t, err)

	assert.Equal(t, post.Body, post2.Body)
}

func TestHasOne(t *testing.T) {
	setup(t)
	post := &Post{
		Body: "first post",
	}
	assert.NoError(t, orm.Save(post))
	assert.Equal(t, int64(1), post.ID)

	headerPicture := &HeaderPicture{
		PostID: post.ID,
		Link:   "google",
	}
	assert.NoError(t, orm.Save(headerPicture))

	c1, err := orm.HasOne[HeaderPicture](post).One()
	assert.NoError(t, err)

	assert.Equal(t, headerPicture.PostID, c1.PostID)
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

	categories, err := orm.BelongsToMany[Category](post).All()
	assert.NoError(t, err)

	assert.Len(t, categories, 1)
}

func TestSchematic(t *testing.T) {
	setup(t)
	orm.Schematic()
}

//func TestQuery(t *testing.T) {
//	t.Run(`test QueryBuilder using qb`, func(t *testing.T) {
//		setup(t)
//
//		post := &Post{
//			Body: "first Post",
//		}
//
//		assert.NoError(t, orm.Save(post))
//		assert.Equal(t, int64(1), post.ID)
//
//		posts, err := orm.Query[Post](orm.NewQueryBuilder[Post]().Where("id", 1))
//		assert.NoError(t, err)
//		assert.Equal(t, []Post{*post}, posts)
//	})
//	t.Run(`test QueryBuilder raw`, func(t *testing.T) {
//		setup(t)
//
//		post := &Post{
//			Body: "first Post",
//		}
//
//		assert.NoError(t, orm.Save(post))
//		assert.Equal(t, int64(1), post.ID)
//
//		posts, err := orm.QueryRaw[Post](`SELECT * FROM posts WHERE id = ?`, 1)
//		assert.NoError(t, err)
//		assert.Equal(t, []Post{*post}, posts)
//	})
//}

func TestExec(t *testing.T) {

	t.Run("test exec Update", func(t *testing.T) {
		setup(t)
		assert.NoError(t, orm.Save(&Post{
			Body: "first post",
		}))
		id, affected, err := orm.Exec[Post](orm.NewQueryBuilder[Post]().Set("body", "first post body updated").Where("id", 1))
		assert.NoError(t, err)
		assert.EqualValues(t, 1, id)
		assert.EqualValues(t, 1, affected)
	})

	t.Run("test delete &Delete", func(t *testing.T) {
		setup(t)

		assert.NoError(t, orm.Save(&Post{
			Body: "first post",
		}))
		res, err := orm.Query[Post]().Where("id", 1).Delete()
		assert.NoError(t, err)
		affected, err := res.RowsAffected()
		assert.NoError(t, err)
		assert.EqualValues(t, 1, affected)
	})
	t.Run("test exec raw", func(t *testing.T) {
		setup(t)
		id, affected, err := orm.ExecRaw[Post](`INSERT INTO posts (id,body) VALUES (1, ?)`, "amirreza")
		assert.NoError(t, err)
		assert.EqualValues(t, 1, id)
		assert.EqualValues(t, 1, affected)
	})
}

func TestAddProperty(t *testing.T) {
	t.Run("having pk value", func(t *testing.T) {
		setup(t)

		post := &Post{
			Body: "first post",
		}

		assert.NoError(t, orm.Save(post))
		assert.EqualValues(t, 1, post.ID)

		err := orm.Add(post, &Comment{PostID: post.ID, Body: "firstComment"})
		assert.NoError(t, err)

		var comment Comment
		assert.NoError(t, orm.GetConnection("default").
			Connection.
			QueryRow(`SELECT id, post_id, body FROM comments WHERE post_id=?`, post.ID).
			Scan(&comment.ID, &comment.PostID, &comment.Body))

		assert.EqualValues(t, post.ID, comment.PostID)
	})
	t.Run("not having PK value", func(t *testing.T) {
		setup(t)
		post := &Post{
			Body: "first post",
		}
		assert.NoError(t, orm.Save(post))
		assert.EqualValues(t, 1, post.ID)

		err := orm.Add(post, &AuthorEmail{Email: "myemail"})
		assert.NoError(t, err)

		emails, err := orm.QueryRaw[AuthorEmail](`SELECT id, email FROM emails WHERE post_id=?`, post.ID)

		assert.NoError(t, err)
		assert.Equal(t, []AuthorEmail{{ID: 1, Email: "myemail"}}, emails)
	})
}

func TestQuery(t *testing.T) {
	t.Run("querying single row", func(t *testing.T) {
		setup(t)
		assert.NoError(t, orm.Save(&Post{Body: "body 1"}))
		//post, err := orm.Query[Post]().Where("id", 1).First()
		post, err := orm.Query[Post]().WherePK(1).First()
		assert.NoError(t, err)
		assert.EqualValues(t, "body 1", post.Body)
		assert.EqualValues(t, 1, post.ID)

	})
	t.Run("querying multiple rows", func(t *testing.T) {
		setup(t)
		assert.NoError(t, orm.Save(&Post{Body: "body 1"}))
		assert.NoError(t, orm.Save(&Post{Body: "body 2"}))
		assert.NoError(t, orm.Save(&Post{Body: "body 3"}))
		posts, err := orm.Query[Post]().All()
		assert.NoError(t, err)
		assert.Len(t, posts, 3)
		assert.Equal(t, "body 1", posts[0].Body)
	})

	t.Run("updating a row using query interface", func(t *testing.T) {
		setup(t)
		assert.NoError(t, orm.Save(&Post{Body: "body 1"}))

		res, err := orm.Query[Post]().Where("id", 1).Update(orm.KV{
			"body": "body jadid",
		})
		assert.NoError(t, err)

		affected, err := res.RowsAffected()
		assert.NoError(t, err)
		assert.EqualValues(t, 1, affected)

		post, err := orm.Find[Post](1)
		assert.NoError(t, err)
		assert.Equal(t, "body jadid", post.Body)
	})

	t.Run("deleting a row using query interface", func(s *testing.T) {
		setup(t)
		assert.NoError(t, orm.Save(&Post{Body: "body 1"}))

		_, err := orm.Query[Post]().WherePK(1).Delete()
		assert.NoError(s, err)
		count, err := orm.Query[Post]().WherePK(1).Count()
		assert.NoError(s, err)
		assert.EqualValues(s, 0, count)
	})
}
