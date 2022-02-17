package main

import (
	"database/sql"
	"fmt"
	"github.com/golobby/orm"
	_ "github.com/mattn/go-sqlite3"
	"os"
)

type Post struct {
	ID         int64
	Comments   []*Comment
	Content    string
	Categories []*Category
}

type Comment struct {
	ID      int64
	PostID  int64
	Post    Post
	Content string
}

type Category struct {
	ID    int64
	Title string
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
	postRepository := orm.Initialize(dbGolobby, orm.dialects.SQLite3, &Post{})
	commentRepository := orm.Initialize(dbGolobby, orm.dialects.SQLite3, &Comment{})
	firstPost := &Post{
		Content: "salam donya",
	}
	err = postRepository.Entity(firstPost).Save()
	if err != nil {
		panic(err)
	}
	fmt.Println("Post primary key is ", firstPost.ID)
	firstComment := &Comment{
		PostID:  firstPost.ID,
		Content: "comment aval",
	}
	err = commentRepository.Entity(firstComment).Save()
	if err != nil {
		panic(err)
	}
	err = postRepository.Entity(firstPost).HasMany(&firstPost.Comments)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Loaded post %d comment -> %+v\n", firstPost.ID, firstPost.Comments[0])
	var newPost Post
	err = commentRepository.Entity(firstComment).BelongsTo(&newPost)
	if err != nil {
		panic(err)
	}
	fmt.Printf("loaded comment %d post %+v", firstComment.PostID, firstPost)

	err = postRepository.
		Entity(firstPost).
		ManyToMany(&firstPost.Categories, orm.ManyToManyConfigurators.IntermediateTable("post_categories"))
	if err != nil {
		panic(err)
	}
	fmt.Printf("loaded post %d categories %+v", firstPost.ID, firstPost.Categories)
}
