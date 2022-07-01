package main

import (
	"bufio"
	"encoding/json"
	"log"
	"moviedb/database"
	"moviedb/movie"
	"moviedb/person"
	"moviedb/tv"
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
	env := os.Getenv("GO_ENV")

	var movieFile = "/home/dyego/dev/moviedb-carga/movie_ids_06_30_2022.json"
	var tvFile = "/home/dyego/dev/moviedb-carga/movie_ids_06_30_2022.json"
	var personFile = "/home/dyego/dev/moviedb-carga/movie_ids_06_30_2022.json"

	if env == "production" {
		movieFile = "/root/data_load/movie_ids_06_30_2022.json"
		tvFile = "/root/data_load/tv_series_ids_06_30_2022.json"
		personFile = "/root/data_load/person_ids_06_30_2022.json"
	}

	var languageEn = "en"
	var languageBr = "pt-BR"

	log.Println("INIT MOVIES")
	fileMovie, err := os.Open(movieFile)

	if err != nil {
		log.Fatal(err)
	}

	defer fileMovie.Close()

	scannerMovies := bufio.NewScanner(fileMovie)

	for scannerMovies.Scan() {

		var movieRead movie.Movie
		json.Unmarshal([]byte(scannerMovies.Text()), &movieRead)

		movieFindEn := movie.GetMovieByIdAndLanguage(movieRead.Id, languageEn)
		if movieFindEn.Id == 0 {
			movieInsert := movie.GetMovieDetailsOnApiDb(movieRead.Id, languageEn)
			movie.PopulateMovieByLanguage(movieInsert, languageEn)
		}

		movieFindBr := movie.GetMovieByIdAndLanguage(movieRead.Id, languageBr)
		if movieFindBr.Id == 0 {
			movieInsert := movie.GetMovieDetailsOnApiDb(movieRead.Id, languageBr)
			movie.PopulateMovieByLanguage(movieInsert, languageBr)
		}
	}
	log.Println("FINISH MOVIES")

	log.Println("INIT SERIES")
	fileTv, err := os.Open(tvFile)

	if err != nil {
		log.Fatal(err)
	}

	defer fileTv.Close()

	scannerTv := bufio.NewScanner(fileTv)

	for scannerTv.Scan() {

		var tvRead tv.Serie
		json.Unmarshal([]byte(scannerTv.Text()), &tvRead)

		tvFindEn := tv.GetSerieByIdAndLanguage(tvRead.Id, languageEn)
		if tvFindEn.Id == 0 {
			tvInsert := tv.GetSerieDetailsOnApiDb(tvRead.Id, languageEn)
			tv.PopulateSerieByLanguage(tvInsert, languageEn)
		}

		tvFindBr := tv.GetSerieByIdAndLanguage(tvRead.Id, languageBr)
		if tvFindBr.Id == 0 {
			tvInsert := tv.GetSerieDetailsOnApiDb(tvRead.Id, languageBr)
			tv.PopulateSerieByLanguage(tvInsert, languageBr)
		}
	}
	log.Println("FINISH SERIES")

	log.Println("INIT PERSONS")

	filePerson, err := os.Open(personFile)

	if err != nil {
		log.Fatal(err)
	}

	defer filePerson.Close()

	scannerPerson := bufio.NewScanner(filePerson)

	for scannerPerson.Scan() {

		var personRead person.Person
		json.Unmarshal([]byte(scannerPerson.Text()), &personRead)

		personFindEn := person.GetPersonByIdAndLanguage(personRead.Id, languageEn)
		if personFindEn.Id == 0 {
			personInsert := person.GetPersonDetailsOnApiDb(personRead.Id, languageEn)
			person.PopulatePersonByLanguage(personInsert, languageEn)
		}

		personFindBr := person.GetPersonByIdAndLanguage(personRead.Id, languageBr)
		if personFindBr.Id == 0 {
			personInsert := person.GetPersonDetailsOnApiDb(personRead.Id, languageBr)
			person.PopulatePersonByLanguage(personInsert, languageBr)
		}

	}
	log.Println("FINISH PERSONS")

	// carga.CargaGeral()

	// c := cron.New()
	// // // c.AddFunc("*/1 * * * *", func() {
	// c.AddFunc("@daily", func() {
	// 	log.Println("[Job] CargaGeral")
	// 	carga.CargaGeral()
	// })
	// log.Println("Start Job")
	// c.Start()

	// g := gin.Default()

	// g.GET("/", func(c *gin.Context) {
	// 	c.JSON(http.StatusOK, gin.H{"appName": "Aplicação/Job de carga do Guia Médico", "env": os.Getenv("NODE_ENV")})
	// })
	// g.Run(":1323")
}
