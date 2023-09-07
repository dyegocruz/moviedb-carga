package main

import (
	"log"
	"moviedb/carga"
	"moviedb/common"
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
	go carga.CheckAndUpdateCatalogByFile(common.MEDIA_TYPE_TV)
	go carga.CheckAndUpdateCatalogByFile(common.MEDIA_TYPE_MOVIE)
	carga.CheckAndUpdateCatalogByFile(common.MEDIA_TYPE_PERSON)
	// env := os.Getenv("GO_ENV")

	// t := time.Now()
	// dateFile := t.Format("01_02_2006")
	// movieFile := "movie_ids_" + dateFile
	// tvFile := "tv_series_ids_" + dateFile
	// personFile := "person_ids_" + dateFile

	// log.Println("INIT MOVIES")
	// downloadExportFile(movieFile)
	// unzip(movieFile)

	// fileMovie, err := os.Open(movieFile + ".json")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// defer fileMovie.Close()

	// scannerMovies := bufio.NewScanner(fileMovie)

	// var movieCatalog = movie.GenerateMovieCatalogCheck(common.LANGUAGE_EN)

	// for scannerMovies.Scan() {
	// 	var elementRead TmdbDailyFile
	// 	json.Unmarshal([]byte(scannerMovies.Text()), &elementRead)
	// 	if movieCatalog[elementRead.Id].Id == 0 {
	// 		log.Println("INSERT MOVIE: ", elementRead.Id)

	// 		movie.PopulateMovieByIdAndLanguage(elementRead.Id, common.LANGUAGE_EN, "Y")
	// 		go movie.PopulateMovieByIdAndLanguage(elementRead.Id, common.LANGUAGE_PTBR, "Y")

	// 	}

	// }

	// RemoveFile(movieFile + ".json")
	// log.Println("FINISH MOVIES")

	// log.Println("------------------------------------------------------------------------------------------------------------------------------------")

	// log.Println("INIT SERIES")
	// downloadExportFile(tvFile)
	// unzip(tvFile)
	// fileTv, err := os.Open(tvFile + ".json")

	// if err != nil {
	// 	log.Fatal(err)
	// }

	// defer fileTv.Close()

	// scannerTv := bufio.NewScanner(fileTv)

	// var tvCatalog = tv.GenerateTvCatalogCheck(common.LANGUAGE_EN)

	// for scannerTv.Scan() {
	// 	var tvRead tv.Serie
	// 	json.Unmarshal([]byte(scannerTv.Text()), &tvRead)

	// 	if tvCatalog[tvRead.Id].Id == 0 {
	// 		tv.PopulateSerieByIdAndLanguage(tvRead.Id, common.LANGUAGE_EN)
	// 		go tv.PopulateSerieByIdAndLanguage(tvRead.Id, common.LANGUAGE_PTBR)
	// 	}
	// }
	// RemoveFile(tvFile + ".json")
	// log.Println("FINISH SERIES")

	// log.Println("------------------------------------------------------------------------------------------------------------------------------------")

	// log.Println("INIT PERSONS")
	// downloadExportFile(personFile)
	// unzip(personFile)
	// filePerson, err := os.Open(personFile + ".json")

	// if err != nil {
	// 	log.Fatal(err)
	// }

	// defer filePerson.Close()

	// scannerPerson := bufio.NewScanner(filePerson)

	// var personCatalog = person.GeneratePersonCatalogCheck(common.LANGUAGE_EN)

	// for scannerPerson.Scan() {

	// 	var personRead person.Person
	// 	json.Unmarshal([]byte(scannerPerson.Text()), &personRead)

	// 	if personCatalog[personRead.Id].Id == 0 {
	// 		person.PopulatePersonByIdAndLanguage(personRead.Id, common.LANGUAGE_EN)
	// 		go person.PopulatePersonByIdAndLanguage(personRead.Id, common.LANGUAGE_PTBR)
	// 	}
	// }
	// RemoveFile(personFile + ".json")
	// log.Println("FINISH PERSONS")

	// carga.GeneralCharge()
	// log.Println("PROCESS CONCLUDED")

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
