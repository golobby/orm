package orm_test

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golobby/orm"
	"github.com/stretchr/testify/assert"
)

func TestExampleRepositoriesNoRel(t *testing.T) {
	type User struct {
		Id   int64  `sqlname:"id" pk:"true"`
		Name string `sqlname:"name"`
		Age  int    `sqlname:"age"`
	}
	// any sql database connection
	db, mockDB, err := sqlmock.New()
	assert.NoError(t, err)
	// create the repository using database connection and an instance of the type representing the table in database.
	userRepository := orm.NewRepository(db, orm.PostgreSQLDialect, &User{})
	firstUser := &User{
		Name: "Amirreza",
	}
	mockDB.ExpectExec(`INSERT INTO users`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	// Saves an entity to database.
	err = userRepository.Save(firstUser)
	assert.NoError(t, err)
	assert.NoError(t, mockDB.ExpectationsWereMet())
	assert.Equal(t, int64(1), firstUser.Id)

	mockDB.ExpectQuery(`SELECT id, name, age FROM users`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "age"}).AddRow(2, "amirreza", 19))
	secondUser := &User{
		Id: 1,
	}
	// Fill the gaps of struct from database.
	err = userRepository.Fill(secondUser)
	assert.NoError(t, err)
	assert.NoError(t, mockDB.ExpectationsWereMet())
	assert.Equal(t, "amirreza", secondUser.Name)
	assert.Equal(t, 19, secondUser.Age)

	//// Update the entity in database.
	secondUser.Age = 11
	mockDB.ExpectExec(`UPDATE users`).WithArgs(2, 2, "amirreza", 11).WillReturnResult(sqlmock.NewResult(1, 1))
	err = userRepository.Update(secondUser)

	assert.NoError(t, mockDB.ExpectationsWereMet())
	assert.NoError(t, err)

	//// Delete the entity from database.
	mockDB.ExpectExec(`DELETE FROM users`).WithArgs(2).WillReturnResult(sqlmock.NewResult(0, 1))
	err = userRepository.Delete(secondUser)
	assert.NoError(t, err)
}

type Address struct {
	Content string `orm:"name=content"`
}

func (a Address) Table() string {
	return "addresses"
}
func TestExampleRepositoriesWithRelationHasOne(t *testing.T) {
	type User struct {
		Id      int64   `orm:"name=id pk=true"`
		Name    string  `orm:"name=name"`
		Age     int     `orm:"name=age"`
		Address Address `orm:"in_rel=true with=addresses has=one left=id right=user_id"`
	}
	db, mockDB, err := sqlmock.New()
	assert.NoError(t, err)
	// create the repository using database connection and an instance of the type representing the table in database.
	userRepository := orm.NewRepository(db, orm.PostgreSQLDialect, &User{})
	firstUser := &User{
		Id: 1,
	}
	mockDB.ExpectQuery(`SELECT users.id, users.name, users.age, addresses.content`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"users.id", "users.name", "users.age", "addresses.content"}).AddRow(1, "amirreza", 23, "ahvaz"))
	err = userRepository.FillWithRelations(firstUser)
	assert.NoError(t, err)
	assert.NoError(t, mockDB.ExpectationsWereMet())
}
