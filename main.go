package main

import (
	"log"
	"moviedb/carga"
	"moviedb/database"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func init() {
	env := os.Getenv("GO_ENV")

	if env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	if env == "" {
		env = "development"
	}
	log.Println("=> ENV: " + env)

	err := godotenv.Load(env + ".env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	database.CheckCreateCollections()

}

func main() {
	// docsTotal := tv.GetCountAll()
	// log.Println(docsTotal)

	// persons := person.GetAllTest(10000)
	// log.Println(len(persons))

	// persons := person.GeneratePersonCatalogCheck(common.LANGUAGE_EN)
	// log.Println(len(persons))

	// var i int64 = 0
	// listIds := make([]int, 0)
	// for _, personCatalog := range persons {

	// 	listIds = append(listIds, personCatalog.Id)

	// 	if i%5000 == 0 {

	// 		log.Println(len(person.GetByListId(listIds)))

	// 		listIds = make([]int, 0)
	// 	}

	// 	i++
	// }

	// var interval int64 = 10000
	// var i int64
	// series := make([]tv.Serie, 0)
	// for i = 0; i < docsTotal; i++ {
	// 	if i%interval == 0 {
	// 		log.Println(i)
	// 		series = append(series, tv.GetAll(i, interval)...)
	// 	}
	// }
	// log.Println(len(series))
	// 	// tv.GetAll(i, 1)

	// 	// log.Println(tv.GetAll(i, 1)[0].Id)
	// }
	// series := tv.GetAllTest(100)
	// log.Println("FINISH", len(series))
	// movies := movie.GetAllTest(10000)
	// log.Println("FINISH", len(movies))

	// tv.GetAll(0, 100)
	carga.GeneralCharge()
	log.Println("PROCESS CONCLUDED")
	// log.Println(runtime.NumCPU())

	// log.Println(len(movie.GetAll(0, 10)))

	// c := cron.New()
	// c.AddFunc("@daily", func() {
	// 	log.Println("[Job] General Charge")
	// 	carga.GeneralCharge()
	// 	log.Println("PROCESS CONCLUDED")
	// })
	// log.Println("Start Job")
	// c.Start()

	// g := gin.Default()

	// g.GET("/", func(c *gin.Context) {
	// 	c.JSON(http.StatusOK, gin.H{"appName": "App to make a Charge data", "env": os.Getenv("GO_ENV")})
	// })
	// g.Run(":1323")
}
