package main

import (
	"log"
	"net/http"
	"github.com/labstack/echo"
	_ "github.com/go-sql-driver/mysql"
	"database/sql"
	"github.com/robfig/cron"
)

// func handler(w http.ResponseWriter, r *http.Request) {
// 	fmt.Fprintf(w, "Hello go web server")
// }

// func main() {
// 	http.HandleFunc("/", handler)
// 	http.ListenAndServe(":8080", nil)
// }


func main() {
	//DB
	db, err := sql.Open("mysql", "root@tcp(127.0.0.1:3306)/go_learn")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// _, err = db.Exec("drop table crons")
	// _, err = db.Exec("create table crons (id integer not null AUTO_INCREMENT, spec varchar(50), hook varchar(255), PRIMARY KEY (id))")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	//cron
	c := cron.New()
	c.Start()
	defer c.Stop()

	addCronFunc := func(cr CronRecord) {
		log.Print("add ", cr)
		c.AddFunc(cr.Spec, func() {
			log.Print("executed ", cr.Hook)
		})
	}

	initCrons, _ := getCrons(db)
	for _, cr := range(initCrons) {
		addCronFunc(cr)
	}
	
	//web api
	e := echo.New()

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!")
	})

	e.GET("/crons", func(c echo.Context) error {
		res, err := getCrons(db)
		if err != nil {
			return err
		}
		return c.JSON(http.StatusOK, res)
	})

	e.POST("crons", func(context echo.Context) error {
		cr := new(CronRecord)
		if err = context.Bind(cr); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid JSON")
		}
		_, err := cron.Parse(cr.Spec)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Invalid Spec")
		}

		_, err = db.Exec(
			`INSERT INTO crons (spec, hook) VALUES (?, ?) `,
			cr.Spec,
			cr.Hook,
		)
		if err != nil {
			log.Fatal(err)
			return err
		}
		
		addCronFunc(*cr)
		return context.JSON(http.StatusOK, cr)
	})
	e.Logger.Fatal(e.Start(":1323"))
}

type CronRecord struct {
	Spec string `json:"spec"`
	Hook string `json:"hook"`
}

func getCrons(db *sql.DB) ([]CronRecord, error) {
		rows, err := db.Query(
			`SELECT spec, hook FROM crons`,
		)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		var res []CronRecord
		for rows.Next() {
			var r CronRecord
			err = rows.Scan(&r.Spec, &r.Hook)			
			if err != nil {
				log.Fatal(err);
        return nil, err
			}
			res = append(res, r)
		}
	return res, nil
}
