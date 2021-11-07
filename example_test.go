package sql_test

import (
	"github.com/golobby/sql/builder"
	"github.com/stretchr/testify/assert"
	"testing"
)

type User struct {
	Id   int    `bind:"id" pk:"true"`
	Name string `bind:"name"`
}

var userSchema = builder.NewSchema(nil, &User{})

func TestImportant(t *testing.T) {
	query, err := builder.
		NewQuery().
		Schema(userSchema).
		Where(builder.WhereHelpers.EqualID("1")).SQL()
	assert.NoError(t, err)
	assert.Equal(t, "SELECT Id, name FROM users WHERE id = 1", query)

	//userSchema.NewModel(&User{
	//	Id:   1,
	//	Name: "amirreza",
	//}).Save()

	//userSchema.NewModel(&User{Id: 1}).Fill()
}
