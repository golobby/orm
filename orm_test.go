package orm_test

import (
	"github.com/golobby/orm"
	"github.com/stretchr/testify/assert"
	"testing"
)

type Post struct {
}

func (p Post) Schema() *orm.Schema {
	return &orm.Schema{}
}

func (p *Post) Comments() ([]Comment, error) {
	return orm.HasMany[Comment](p, orm.HasManyConfig{})
}

type Comment struct {
	ID   int
	Text string
}

func (c Comment) Schema() *orm.Schema {
	return &orm.Schema{}
}

func (c *Comment) Post() (Post, error) {
	return orm.BelongsTo[Post](c, orm.BelongsToConfig{})
}

type Category struct {
}

func (c Category) Schema() *orm.Schema {
	return &orm.Schema{}
}

func (c Category) Posts() ([]Post, error) {
	return orm.ManyToMany[Post](c, orm.ManyToManyConfig{})
}

func setup(t *testing.T) {
	err := orm.Initialize(orm.ConnectionConfig{
		Name:             "sqlite3",
		Driver:           "sqlite3",
		ConnectionString: ":memory:",
		Entities:         []orm.Entity{&Comment{}, &Post{}, &Category{}},
	})
	assert.NoError(t, err)

}

func TestFind(t *testing.T) {
	setup(t)
	comment, err := orm.Find[Comment](1)
	assert.NoError(t, err)
	assert.Equal(t, "comment 1", comment.Text)
	assert.Equal(t, 1, comment.ID)
}

func TestSave(t *testing.T) {
	setup(t)
	err := orm.Insert(&Post{})
	var p Post

	cs, err := p.Comments()

	assert.NoError(t, err)

	// Query
	orm.Find[Post](1)
}
