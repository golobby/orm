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

func (u *User) Parent(output *User) orm.Relation {
	return u.HasOne(output, orm.HasOneConfig{})
}

func (u *User) Addresses(addresses []orm.IsEntity) orm.Relation {
	return u.HasMany(addresses, orm.HasManyConfig{})
}

func main() {
	orm.Initialize()
	var user User
	var parent User
	var addresses []Address
	orm.AsEntity(&user).Load(user.Addresses).Scan(addresses)
	orm.AsEntity(user).Load(user.Addresses(addresses))
}
