package main

import (
	"database/sql"
	"fmt"

	"github.com/golobby/orm"
	_ "github.com/mattn/go-sqlite3"
)

type Post struct {
	ID       int64
	Comments []*Comment
	Content  string
}

type Comment struct {
	ID int64
	PostID int64
	Post Post
	Content string
}

func main() {
	dbGolobby, err := sql.Open("sqlite3", "blog.db")
	if err != nil {
		panic(err)
	}
	if err = dbGolobby.Ping(); err != nil {
		panic(err)
	}
	createPosts := `CREATE TABLE IF NOT EXISTS posts (id integer primary key, content text);`
	createComments := `CREATE TABLE IF NOT EXISTS comments (id integer primary key, post_id integer, content text);`
	_, err = dbGolobby.Exec(createPosts)
	if err != nil {
		panic(err)
	}
	_, err = dbGolobby.Exec(createComments)
	if err != nil {
		panic(err)
	}
	postRepository := orm.NewRepository(dbGolobby, orm.Dialects.SQLite3, &Post{})
	commentRepository := orm.NewRepository(dbGolobby, orm.Dialects.SQLite3, &Comment{})
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
}