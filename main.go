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

func reverseArray(arrayElement []string) []string {
	arrayElementSize := len(arrayElement)
	revArr := make([]string, arrayElementSize)
	j := 0
	for i := arrayElementSize - 1; i >= 0; i-- {
		revArr[j] = arrayElement[i]
		j++
	}

	return revArr
}

func main() {
	env := os.Getenv("GO_ENV")

	// var movieFile = "./movie_ids_06_30_2022.json"
	// var tvFile = "./tv_series_ids_06_30_2022.json"
	// var personFile = "./person_ids_06_30_2022.json"

	if env == "production" {
		// movieFile = "./movie_ids_06_30_2022.json"
		// tvFile = "./tv_series_ids_06_30_2022.json"
		// personFile = "./person_ids_06_30_2022.json"
	}

	// var languageEn = "en"
	// var languageBr = "pt-BR"

	// log.Println("INIT MOVIES")
	// fileMovie, err := os.Open(movieFile)

	// if err != nil {
	// 	log.Fatal(err)
	// }

	// defer fileMovie.Close()

	// scannerMovies := bufio.NewScanner(fileMovie)

	// for scannerMovies.Scan() {

	// 	var movieRead movie.Movie
	// 	json.Unmarshal([]byte(scannerMovies.Text()), &movieRead)

	// 	movieFindEn := movie.GetMovieByIdAndLanguage(movieRead.Id, languageEn)
	// 	if movieFindEn.Id == 0 {
	// 		movieInsert := movie.GetMovieDetailsOnApiDb(movieRead.Id, languageEn)
	// 		movie.PopulateMovieByLanguage(movieInsert, languageEn)
	// 	}

	// 	movieFindBr := movie.GetMovieByIdAndLanguage(movieRead.Id, languageBr)
	// 	if movieFindBr.Id == 0 {
	// 		movieInsert := movie.GetMovieDetailsOnApiDb(movieRead.Id, languageBr)
	// 		movie.PopulateMovieByLanguage(movieInsert, languageBr)
	// 	}
	// }
	// log.Println("FINISH MOVIES")

	// log.Println("INIT SERIES")
	// fileTv, err := os.Open(tvFile)

	// if err != nil {
	// 	log.Fatal(err)
	// }

	// defer fileTv.Close()

	// scannerTv := bufio.NewScanner(fileTv)

	// tvScannerArray := make([]string, 0)
	// for scannerTv.Scan() {
	// 	tvScannerArray = append(tvScannerArray, scannerTv.Text())
	// }
	// tvScannerArray = reverseArray(tvScannerArray)

	// for _, tvScanner := range tvScannerArray {

	// 	var tvRead tv.Serie
	// 	json.Unmarshal([]byte(tvScanner), &tvRead)

	// 	// tvFindEn := tv.GetSerieByIdAndLanguage(tvRead.Id, common.LANGUAGE_EN)
	// 	// if tvFindEn.Id == 0 {
	// 	tvInsert := tv.GetSerieDetailsOnTMDBApi(tvRead.Id, common.LANGUAGE_EN)
	// 	tv.PopulateSerieByLanguage(tvInsert, common.LANGUAGE_EN)
	// 	// }

	// 	// tvFindBr := tv.GetSerieByIdAndLanguage(tvRead.Id, common.LANGUAGE_PTBR)
	// 	// if tvFindBr.Id == 0 {
	// 	tvBrInsert := tv.GetSerieDetailsOnTMDBApi(tvRead.Id, common.LANGUAGE_PTBR)
	// 	tv.PopulateSerieByLanguage(tvBrInsert, common.LANGUAGE_PTBR)
	// 	// }
	// }
	// log.Println("FINISH SERIES")

	// log.Println("INIT PERSONS")

	// filePerson, err := os.Open(personFile)

	// if err != nil {
	// 	log.Fatal(err)
	// }

	// defer filePerson.Close()

	// scannerPerson := bufio.NewScanner(filePerson)

	// for scannerPerson.Scan() {

	// 	var personRead person.Person
	// 	json.Unmarshal([]byte(scannerPerson.Text()), &personRead)

	// 	personFindBr := person.GetPersonByIdAndLanguage(personRead.Id, languageBr)
	// 	if personFindBr.Id == 0 {
	// 		personInsert := person.GetPersonDetailsOnApiDb(personRead.Id, languageBr)
	// 		person.PopulatePersonByLanguage(personInsert, languageBr)

	// 		personFindEn := person.GetPersonByIdAndLanguage(personRead.Id, languageEn)
	// 		if personFindEn.Id == 0 {
	// 			personInsert := person.GetPersonDetailsOnApiDb(personRead.Id, languageEn)
	// 			person.PopulatePersonByLanguage(personInsert, languageEn)
	// 		} else {
	// 			log.Println("PERSON EN ALREADY INSERTED: ", personRead.Id)
	// 		}
	// 	} else {
	// 		log.Println("PERSON PT-BR ALREADY INSERTED: ", personRead.Id)
	// 	}

	// }
	// log.Println("FINISH PERSONS")

	carga.GeneralCharge()

	// movie.CheckMoviesChanges()

	// person.CheckPersonChanges()

	// currentMovieTime := time.Now()
	// log.Println(currentMovieTime.Format("2006-01-02"))

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
