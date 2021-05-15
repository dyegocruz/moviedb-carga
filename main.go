package main

import (
	"log"
	"moviedb/carga"
	"moviedb/database"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/robfig/cron"
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

	// carga.CargaGeral()
	// parametro.GetByTipo("CARGA_TMDB_CONFIG")

	c := cron.New()
	// // c.AddFunc("*/1 * * * *", func() {
	c.AddFunc("@daily", func() {
		log.Println("[Job] CargaGeral")
		carga.CargaGeral()
	})
	log.Println("Start Job")
	c.Start()

	g := gin.Default()

	g.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"appName": "Aplicação/Job de carga do Guia Médico", "env": os.Getenv("NODE_ENV")})
	})
	g.Run(":1323")
}
