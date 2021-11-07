package orm_test

import (
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golobby/orm"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSchema(t *testing.T) {
	type User struct {
		Id   int    `bind:"id" pk:"true"`
		Name string `bind:"name"`
	}
	db, _, err := sqlmock.New()

	var userSchema = orm.NewSchema(db, &User{})
	query, err := orm.NewQuery().
		Schema(userSchema).
		Where(orm.WhereHelpers.EqualID("1")).SQL()
	assert.NoError(t, err)
	assert.Equal(t, "SELECT id, name FROM users WHERE id = 1", query)
	//sm.ExpectExec(`INSERT INTO users (name) VALUES (.*)`)
	//err = userSchema.NewModel(&User{
	//	Id:   1,
	//	Name: "amirreza",
	//}).Save()
	//assert.NoError(t, err)
	//
	//assert.NoError(t, sm.ExpectationsWereMet())

	//userSchema.NewModel(&User{Id: 1}).Fill()
}
