package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/golobby/orm"
	_ "github.com/mattn/go-sqlite3"
)

type Post struct {
	ID         int64
	comments   []*Comment
	Content    string
	categories []*Category
}

func (p *Post) MD() *orm.MetaData {
	return &orm.MetaData{}
}

func (p *Post) Comments(e *orm.Entity) orm.Relation {
	return e.HasMany(p.comments, orm.HasManyConfig{})
}

func (p *Post) Categories(e *orm.Entity) orm.Relation {
	return e.ManyToMany(p.categories, orm.ManyToManyConfig{IntermediateTable: "post_categories"})
}

type Comment struct {
	ID      int64
	PostID  int64
	post    *Post
	Content string
}

func (c *Comment) Post(e *orm.Entity) orm.Relation {
	return e.BelongsTo(c.post, orm.BelongsToConfig{})
}

func (c *Comment) MD() *orm.MetaData {
	return &orm.MetaData{}
}

type Category struct {
	ID    int64
	Title string
	posts []*Post
}

func (c *Category) Posts(e *orm.Entity) orm.Relation {
	return e.ManyToMany(c.posts, orm.ManyToManyConfig{IntermediateTable: "post_categories"})
}

func (c *Category) MD() *orm.MetaData {
	return &orm.MetaData{}
}
func main() {
	_ = os.Remove("blog.db")
	dbGolobby, err := sql.Open("sqlite3", "blog.db")
	if err != nil {
		panic(err)
	}
	if err = dbGolobby.Ping(); err != nil {
		panic(err)
	}
	createPosts := `CREATE TABLE IF NOT EXISTS posts (id integer primary key, content text);`
	createComments := `CREATE TABLE IF NOT EXISTS comments (id integer primary key, post_id integer, content text);`
	createCategories := `CREATE TABLE IF NOT EXISTS categories (id integer primary key, name text);`
	createPostCategories := `CREATE TABLE IF NOT EXISTS post_categories(id integer primary key, post_id int, category_id int)`
	_, err = dbGolobby.Exec(createPosts)
	if err != nil {
		panic(err)
	}
	_, err = dbGolobby.Exec(createComments)
	if err != nil {
		panic(err)
	}
	_, err = dbGolobby.Exec(createCategories)
	if err != nil {
		panic(err)
	}
	_, err = dbGolobby.Exec(createPostCategories)
	if err != nil {
		panic(err)
	}

	// Initializing ORM
	orm.Initialize(orm.ConnectionConfig{DB: dbGolobby, Name: "test", Entities: []orm.Entity{&Post{}, &Comment{}}})

	// Saving first Post
	firstPost := &Post{
		Content: "salam donya",
	}
	err = orm.AsEntity(firstPost).Fill()
	if err != nil {
		panic(err)
	}
	fmt.Println("Post primary key is ", firstPost.ID)

	// Saving a comment
	firstComment := &Comment{
		PostID:  firstPost.ID,
		Content: "comment aval",
	}

	err = orm.AsEntity(firstComment).Save()
	if err != nil {
		panic(err)
	}

	err = orm.AsEntity(firstPost).Load(firstPost.Comments)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Loaded post %d comment -> %+v\n", firstPost.ID, firstPost.comments[0])

	err = orm.AsEntity(firstComment).Load(firstComment.Post)
	if err != nil {
		panic(err)
	}
	fmt.Printf("loaded comment %d post %+v", firstComment.PostID, firstComment.post)

	err = orm.
		AsEntity(firstPost).
		Load(firstPost.Categories)

	if err != nil {
		panic(err)
	}
	fmt.Printf("loaded post %d categories %+v", firstPost.ID, firstPost.categories)
}
