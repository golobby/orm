package orm_test

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golobby/orm"
	"github.com/stretchr/testify/assert"
)

type HeadMaster struct {
	orm.BaseModel
	ID       int64
	Name     string
	Students []*Student
}

type Student struct {
	orm.BaseModel
	HeadMaster *HeadMaster
	ID         int64
	Name       string
	Courses    []*Course
}

type Course struct {
	orm.BaseModel
	ID       int64
	Name     string
	Students []*Student
}

func TestMyWishes(t *testing.T) {
	db, mockDB, err := sqlmock.New()
	assert.NoError(t, err)

	studentRepo := orm.NewRepository(db, orm.Sqlite3SQLDialect, &Student{})
	courseRepo := orm.NewRepository(db, orm.Sqlite3SQLDialect, &Course{})

	me := &Student{ID: 1}
	
	courses := studentRepo.
		Entity(me).
		HasMany(orm.RelationMetadata{
		Table:   tableNameFrom(&Course{}),
		Type:    "",
		Lookup:  "",
		Columns: nil,
	}).([]*Course)
	
	studentRepo.
		Entity(me).
		BelongsTo(me.HeadMaster)
}
