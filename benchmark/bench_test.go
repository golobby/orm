package benchmark

import (
	"database/sql"
	"fmt"
	"github.com/golobby/orm"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"testing"
)

type User struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
}

var (
	gormDB *gorm.DB
)

func setupGolobby() {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		panic(err)
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY, username text)`)
	if err != nil {
		panic(err)
	}
	err = orm.SetupConnections(orm.ConnectionConfig{
		Name:                "default",
		DB:                  db,
		Dialect:             orm.Dialects.SQLite3,
		Entities:            []orm.Entity{User{}},
		DatabaseValidations: true,
	})
	if err != nil {
		panic(err)
	}

}

func setupGORM() {
	var err error
	gormDB, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic(err)
	}
	err = gormDB.AutoMigrate(&User2{})
	if err != nil {
		panic(err)
	}
}

func (u User) ConfigureEntity(e *orm.EntityConfigurator) {
	e.Table("users").PKSetter(func(o orm.Entity, pk interface{}) {
		user := o.(*User)
		user.ID = pk.(int64)
	})
}

func BenchmarkGolobby(t *testing.B) {
	setupGolobby()
	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		var user User
		user.Username = "amir" + fmt.Sprint(i)
		err := orm.InsertAll(&user)
		if err != nil {
			panic(err)
		}
	}
}

type User2 struct {
	gorm.Model
	ID       int64  `json:"id"`
	Username string `json:"username"`
}

func BenchmarkGorm(t *testing.B) {
	setupGORM()
	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		var user User2
		user.Username = "amir" + fmt.Sprint(i)
		err := gormDB.Create(&user).Error
		if err != nil {
			panic(err)
		}
	}
}
