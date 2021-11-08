package orm_test

import (
	"database/sql"
	"github.com/golobby/orm"
	"testing"
)

func TestExampleRepositories(t *testing.T) {
	type User struct {
		Id   int    `bind:"id" pk:"true"`
		Name string `bind:"name"`
		Age  int    `bind:"age"`
	}
	// any sql database connection
	db, _ := sql.Open("", "")

	// create the repository using database connection and an instance of the type representing the table in database.
	userRepository := orm.NewRepository(db, &User{})
	firstUser := &User{
		Name: "Amirreza",
	}
	// Saves an entity to database.
	err := userRepository.Save(firstUser)

	secondUser := &User{
		Id: 1,
	}
	// Fill the gaps of struct from database.
	err = userRepository.Fill(secondUser)

	// Update the entity in database.
	secondUser.Age = 11
	err = userRepository.Update(secondUser)

	// Delete the entity from database.
	err = userRepository.Delete(secondUser)

	// Custom query with binding to a custom struct.
	type onlyName struct {
		Name string `bind:"name"`
	}
	var names []onlyName
	err = userRepository.Query().
		Where(orm.WhereHelpers.Like("name", "%Amir%")).
		Limit(10).
		Select("name").
		Bind(&names)
}
