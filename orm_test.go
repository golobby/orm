package orm_test

import (
	"github.com/golobby/orm"
	"github.com/stretchr/testify/assert"
	"testing"
)

var PostMD *orm.EntityMetadata
var CommentMD *orm.EntityMetadata
var CategoryMD *orm.EntityMetadata

type Post struct {
}

func (p *Post) MD() *orm.BaseMetadata {
	return &orm.BaseMetadata{}
}
func (p *Post) Comments() ([]Comment, error) {
	return orm.HasMany[Comment](p, CommentMD, orm.HasManyConfig{})
}

type Comment struct {
	ID   int
	Text string
}

func (c Comment) MD() *orm.BaseMetadata {
	return &orm.BaseMetadata{}
}
func (c *Comment) Post() (Post, error) {
	return orm.BelongsTo[Post](c, PostMD, orm.BelongsToConfig{})
}

type Category struct {
}

func (c *Category) MD() *orm.BaseMetadata {
	return &orm.BaseMetadata{}
}

func setup(t *testing.T) {
	err := orm.Initialize(orm.ConnectionConfig{
		Name:             "sqlite3",
		Driver:           "sqlite3",
		ConnectionString: ":memory:",
		Entities:         []orm.Entity{&Comment{}, &Post{}, &Category{}},
		EntityMDs:        []*orm.EntityMetadata{CommentMD, PostMD, CategoryMD},
	})
	assert.NoError(t, err)

}

func TestFind(t *testing.T) {
	setup(t)
	comment, err := orm.Find[Comment](CommentMD, 1)
	assert.NoError(t, err)
	assert.Equal(t, "comment 1", comment.Text)
	assert.Equal(t, 1, comment.ID)
}

func TestSave(t *testing.T) {
	setup(t)
	err := orm.Save(&Post{})
	var p Post

	cs, err := p.Comments()
	cs[0].Post()
	assert.NoError(t, err)

	// Query
}
