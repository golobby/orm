package orm_test

import "github.com/golobby/orm"

type User struct {
	Name string
}

func (u *User) EntityConfig() orm.EntityConfig {
	return orm.EntityConfig{
		Table: "users",
	}
}

func main() {
	orm.Initialize()
	var user User
	orm.Entity(&user).Fill()
}
