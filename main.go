package main

import (
	"log"
	"moviedb/carga"
	"moviedb/configs"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron"
)

// func init() {
// 	env := os.Getenv("GO_ENV")

// 	if env == "production" {
// 		gin.SetMode(gin.ReleaseMode)
// 	}

// 	if env == "" {
// 		env = "development"
// 	}
// 	log.Println("=> ENV: " + env)

// 	err := godotenv.Load(env + ".env")
// 	if err != nil {
// 		log.Fatal("Error loading .env file")
// 	}

// 	database.CheckCreateCollections()

// }

func init() {
	if configs.GetEnv() == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
}

func main() {
	// carga.GeneralCharge()
	// log.Println("PROCESS COMPLETE")

	c := cron.New()
	c.AddFunc("@daily", func() {
		log.Println("[Job] General Charge")
		carga.GeneralCharge()
		log.Println("PROCESS COMPLETE")
	})
	log.Println("Start Job")
	c.Start()

	g := gin.Default()

	g.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"appName": "App to make a Charge data", "env": os.Getenv("GO_ENV")})
	})
	g.Run(":1323")
}
