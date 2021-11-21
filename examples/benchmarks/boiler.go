package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/golobby/orm/examples/benchmarks/models"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

func boiler() {
	dbBoiler, err := sql.Open("sqlite3", "boiler.db")
	if err != nil {
	    panic(err)
	}
	if err = dbBoiler.Ping(); err != nil {
		panic(err)
	}

	start := time.Now()
	func() {
		for i := 0; i < 10000; i++ {
			r := &models.Record{
				Name: null.StringFrom("comrade"+fmt.Sprint(i)),
			}
			err = r.Insert(context.Background(), dbBoiler, boil.Infer())
			if err != nil {
				log.Println("iboiler error: ", err.Error())
			    continue
			}
			err = models.Records(models.RecordWhere.ID.EQ(r.ID)).Bind(context.Background(), dbBoiler, r)
			if err != nil {
				log.Println("sboiler error: ", err.Error())
				continue
			}
			r.Name = null.StringFrom(r.Name.String + "-updated")
			_, err = r.Update(context.Background(), dbBoiler, boil.Infer())
			if err != nil {
				log.Println("uboiler error: ", err.Error())
				continue
			}

			_, err = r.Delete(context.Background(), dbBoiler)
			if err != nil {
				log.Println("dboiler error: ", err.Error())
				continue
			}
		}
	}()
	fmt.Printf("Boiler finished in %d miliseconds\n", time.Now().Sub(start).Milliseconds())

}
