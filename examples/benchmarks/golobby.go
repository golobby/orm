package main

import (
	"database/sql"
	"fmt"
	"github.com/golobby/orm"
	"log"
	"time"
)

type Record struct {
	ID   int64  `orm:"name=id pk=true"`
	Name string `orm:"name=name"`
}

func (r *Record) MD() *orm.MetaData {
	return &orm.MetaData{}
}
func golobby() {

	dbGolobby, err := sql.Open("sqlite3", "golobby.db")
	if err != nil {
		panic(err)
	}

	if err = dbGolobby.Ping(); err != nil {
		panic(err)
	}
	createTable := `CREATE TABLE IF NOT EXISTS records (id integer primary key, name text);`
	_, err = dbGolobby.Exec(createTable)
	if err != nil {
		panic(err)
	}
	start := time.Now()

	// golobby
	func() {
		err := orm.Initialize(orm.ConnectionConfig{
			Name:     "test",
			DB:       dbGolobby,
			Entities: []orm.Entity{&Record{}},
		})
		if err != nil {
			panic(err)
		}
		for i := 0; i < 10000; i++ {
			m := &Record{
				Name: "comrade-" + fmt.Sprint(i),
			}
			entity := orm.AsEntity(m)
			err := entity.Save()
			if err != nil {
				log.Println("sGolobby error : ", err.Error())
				continue
			}
			err = entity.Fill()
			if err != nil {
				log.Println("rGolobby error : ", err.Error())
				continue
			}
			m.Name = m.Name + " updated"
			err = entity.Update()
			if err != nil {
				log.Println("uGolobby error : ", err.Error())
				continue
			}

			err = entity.Delete()
			if err != nil {
				log.Println("dGolobby error : ", err.Error())
				continue
			}
		}
	}()
	fmt.Printf("Golobby finished in %d miliseconds\n", time.Now().Sub(start).Milliseconds())
}
