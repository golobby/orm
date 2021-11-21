package orm_test

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/golobby/orm"
	"github.com/stretchr/testify/assert"
)

func TestExampleRepositoriesNoRel(t *testing.T) {
	type User struct {
		Id   int64
		Name string
		Age  int
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
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	// Saves an entity to database.
	err = userRepository.Save(firstUser)
	assert.NoError(t, err)
	assert.NoError(t, mockDB.ExpectationsWereMet())
	assert.Equal(t, int64(1), firstUser.Id)

	mockDB.ExpectQuery(`SELECT users.id, users.name, users.age FROM users`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"users.id", "users.name", "users.age"}).AddRow(2, "amirreza", 19))
	secondUser := &User{
		Id: 1,
	}
	// Fill the gaps of struct from database.
	err = userRepository.Fill(secondUser)
	assert.NoError(t, err)
	assert.NoError(t, mockDB.ExpectationsWereMet())
	assert.Equal(t, "amirreza", secondUser.Name)
	assert.Equal(t, 19, secondUser.Age)

	// Update the entity in database.
	secondUser.Age = 11
	mockDB.ExpectExec(`UPDATE users`).WithArgs(2, 2, "amirreza", 11).WillReturnResult(sqlmock.NewResult(1, 1))
	err = userRepository.Update(secondUser)

	assert.NoError(t, mockDB.ExpectationsWereMet())
	assert.NoError(t, err)

	// Delete the entity from database.
	mockDB.ExpectExec(`DELETE FROM users`).WithArgs(2).WillReturnResult(sqlmock.NewResult(0, 1))
	err = userRepository.Delete(secondUser)
	assert.NoError(t, err)
}

type AddressContent struct {
	AddressID string
	Content   string
}

type Address struct {
	UserID         string
	AddressContent AddressContent `orm:"in_rel=true with=address_content has=one left=id right=address_id"`
}

func TestEntity_HasMany(t *testing.T) {
	type Address struct {
		ID      int64
		Content string
	}
	type User struct {
		ID      int64
		Name    string
		Age     int
		Address Address `orm:"in_rel=true has=one left=id right=user_id"`
	}
	db, mockDB, err := sqlmock.New()
	assert.NoError(t, err)
	// create the repository using database connection and an instance of the type representing the table in database.
	userRepository := orm.NewRepository(db, orm.PostgreSQLDialect, &User{})
	firstUser := &User{
		ID: 1,
	}
	var addresses []*Address
	mockDB.ExpectQuery(`SELECT addresses.id, addresses.content FROM addresses`).
		WithArgs(1).
		WillReturnRows(sqlmock.NewRows([]string{"addresses.id", "addresses.content"}).
			AddRow(1, "ahvaz"))

	err = userRepository.Entity(firstUser).HasMany(&addresses)
	assert.NoError(t, err)
	assert.Len(t, addresses, 1)
}