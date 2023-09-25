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

	// tvCatalog := tv.GenerateTvCatalogCheck(common.LANGUAGE_EN)
	// log.Println(len(tvCatalog))

	// var i int64 = 0
	// var interval int64 = 1000
	// idsGet := make([]int, 0)
	// for _, catalog := range tvCatalog {
	// 	// log.Println(catalog.Id)

	// 	idsGet = append(idsGet, catalog.Id)

	// 	if i%interval == 0 {
	// 		log.Println(len(idsGet))

	// 		batch := tv.GetByListId(idsGet)
	// 		log.Println(len(batch))
	// 		idsGet = make([]int, 0)
	// 	}

	// 	i++
	// }

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
