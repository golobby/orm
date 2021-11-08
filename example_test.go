package orm_test

import (
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golobby/orm"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestExampleRepositories(t *testing.T) {
	type User struct {
		Id   int64  `bind:"id" pk:"true"`
		Name string `bind:"name"`
		Age  int    `bind:"age"`
	}
	// any sql database connection
	db, mockDB, err := sqlmock.New()
	assert.NoError(t, err)
	// create the repository using database connection and an instance of the type representing the table in database.
	userRepository := orm.NewRepository(db, &User{})
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
	//secondUser.Age = 11
	//err = userRepository.Update(secondUser)
	//assert.NoError(t, err)
	//
	//// Delete the entity from database.
	//err = userRepository.Delete(secondUser)
	//assert.NoError(t, err)
	//
	//// Custom query with binding to a custom struct.
	//type onlyName struct {
	//	Name string `bind:"name"`
	//}
	//var names []onlyName
	//err = userRepository.Query().
	//	Where(orm.WhereHelpers.Like("name", "%Amir%")).
	//	Limit(10).
	//	Select("name").
	//	Bind(&names)
	//assert.NoError(t, err)

}
