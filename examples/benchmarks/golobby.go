package main

import (
	"database/sql"
	"fmt"
	"github.com/golobby/orm"
	"log"
	"time"
)

func golobby() {
	type Record struct {
		ID   int64  `orm:"name=id pk=true"`
		Name string `orm:"name=name"`
	}
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
		repo := orm.NewRepository(dbGolobby, orm.Sqlite3SQLDialect, &Record{})

		for i := 0; i < 10000; i++ {
			m := &Record{
				Name: "comrade-" + fmt.Sprint(i),
			}
			err := repo.Save(m)
			if err != nil {
				log.Println("sGolobby error : ", err.Error())
				continue
			}
			err = repo.Fill(m, false)
			if err != nil {
				log.Println("rGolobby error : ", err.Error())
				continue
			}
			m.Name = m.Name + " updated"
			err = repo.Update(m)
			if err != nil {
				log.Println("uGolobby error : ", err.Error())
				continue
			}

			err = repo.Delete(m)
			if err != nil {
				log.Println("dGolobby error : ", err.Error())
				continue
			}
		}
	}()
	fmt.Printf("Golobby finished in %d miliseconds\n", time.Now().Sub(start).Milliseconds())
}
