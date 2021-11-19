package main

import (
	"database/sql"
	"fmt"
	"github.com/golobby/orm"
	_ "github.com/mattn/go-sqlite3"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
	"sync"
	"time"
)

func main() {
	fmt.Println("Benchmarking Golang ORMs ...")
	wg := &sync.WaitGroup{}
	wg.Add(2)

	type Record struct {
		ID   int    `orm:"name=id pk=true"`
		Name string `orm:"name=name"`
	}
	dbGorm, err := gorm.Open(sqlite.Open("gorm.db"), &gorm.Config{})
	if err != nil {
		panic(err)
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
	err = dbGorm.Exec(createTable).Error
	if err != nil {
		panic(err)
	}

	start := time.Now()
	// golobby
	func() {
		repo := orm.NewRepository(dbGolobby, orm.MySQLDialect, &Record{})

		for i := 0; i < 10000; i++ {
			m := &Record{
				Name: "comrade-" + fmt.Sprint(i),
			}
			err := repo.Save(m)
			if err != nil {
				log.Println("sGolobby error : ", err.Error())
				continue
			}
			err = repo.Fill(m)
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
	//gorm
	start = time.Now()
	func() {
		for i := 0; i < 10000; i++ {
			m := &Record{
				Name: "comrade-" + fmt.Sprint(i),
			}
			err := dbGorm.Create(m).Error
			if err != nil {
				log.Println("sGorm error : ", err.Error())
				continue
			}
			err = dbGorm.First(m, m.ID).Error
			if err != nil {
				log.Println("rGorm error : ", err.Error())
				continue
			}
			m.Name = m.Name + " updated"
			err = dbGorm.Save(m).Error
			if err != nil {
				log.Println("uGorm error : ", err.Error())
				continue
			}

			err = dbGorm.Delete(m).Error
			if err != nil {
				log.Println("dGorm error : ", err.Error())
				continue
			}
		}
	}()
	fmt.Printf("Gorm finished in %d miliseconds\n", time.Now().Sub(start).Milliseconds())
}
