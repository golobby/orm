package orm_test

import "github.com/golobby/orm"

type Address string

type User struct {
	orm.BaseEntity
	Name string
}

func (u *User) EntityConfig() orm.EntityConfig {
	return orm.EntityConfig{}
}

func (u *User) Parent(output interface{}) orm.ExecutableQuery {
	return u.HasOne(output, orm.HasOneConfig{})
}

func (u *User) Addresses(addresses interface{}) orm.ExecutableQuery {
	return u.HasMany(addresses, orm.HasManyConfig{})
}

func main() {
	orm.Initialize()
	var user User
	//var parent User
	var addresses []Address
	orm.Select().Where().Fetch(user)
	orm.Select().Where().All(users)
	err := orm.
		AsEntity(&user).
		Load(user.Addresses).
		Scan(&addresses)
	if err != nil {
		panic(err)
	}
}
