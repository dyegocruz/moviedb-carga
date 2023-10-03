package main

import (
	"log"
	"moviedb/carga"
	"moviedb/configs"
	"moviedb/database"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/robfig/cron"
)

func init() {
	if configs.GetEnv() == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	database.CheckCreateCollections()
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
