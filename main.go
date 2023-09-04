package main

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"moviedb/common"
	"moviedb/database"
	"moviedb/movie"
	"moviedb/person"
	"moviedb/tv"
	"net/http"
	"os"
	"time"

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

func downloadExportFile(name string) {
	url := fmt.Sprintf("http://files.tmdb.org/p/exports/%s.json.gz", name)
	resp, _ := http.Get(url)
	defer resp.Body.Close()
	filename := fmt.Sprintf("%s.json.gz", name)
	out, _ := os.Create(filename)
	defer out.Close()
	io.Copy(out, resp.Body)
}

func unzip(name string) {
	// Open compressed file
	gzipFile, err := os.Open(name + ".json.gz")
	if err != nil {
		log.Fatal(err)
	}

	// Create a gzip reader on top of the file reader
	// Again, it could be any type reader though
	gzipReader, err := gzip.NewReader(gzipFile)
	if err != nil {
		log.Fatal(err)
	}
	defer gzipReader.Close()

	// Uncompress to a writer. We'll use a file writer
	outfileWriter, err := os.Create(name + ".json")
	if err != nil {
		log.Fatal(err)
	}
	defer outfileWriter.Close()

	// Copy contents of gzipped file to output file
	_, err = io.Copy(outfileWriter, gzipReader)
	if err != nil {
		log.Fatal(err)
	}

	RemoveFile(name + ".json.gz")
}

func RemoveFile(name string) {
	e := os.Remove(name)
	if e != nil {
		log.Fatal(e)
	}
}

func main() {
	// env := os.Getenv("GO_ENV")

	t := time.Now()
	dateFile := t.Format("01_02_2006")
	movieFile := "movie_ids_" + dateFile
	tvFile := "tv_series_ids_" + dateFile
	personFile := "person_ids_" + dateFile

	log.Println("INIT MOVIES")
	downloadExportFile(movieFile)
	unzip(movieFile)

	fileMovie, err := os.Open(movieFile + ".json")
	if err != nil {
		log.Fatal(err)
	}
	defer fileMovie.Close()

	scannerMovies := bufio.NewScanner(fileMovie)

	var movieCatalog = movie.GenerateMovieCatalogCheck(common.LANGUAGE_EN)

	for scannerMovies.Scan() {
		var movieRead movie.Movie
		json.Unmarshal([]byte(scannerMovies.Text()), &movieRead)
		if movieCatalog[movieRead.Id].Id == 0 {
			log.Println("INSERT MOVIE: ", movieRead.Id)

			movie.PopulateMovieByIdAndLanguage(movieRead.Id, common.LANGUAGE_EN, "Y")
			go movie.PopulateMovieByIdAndLanguage(movieRead.Id, common.LANGUAGE_PTBR, "Y")

		}

	}

	RemoveFile(movieFile + ".json")
	log.Println("FINISH MOVIES")

	log.Println("------------------------------------------------------------------------------------------------------------------------------------")

	log.Println("INIT SERIES")
	downloadExportFile(tvFile)
	unzip(tvFile)
	fileTv, err := os.Open(tvFile + ".json")

	if err != nil {
		log.Fatal(err)
	}

	defer fileTv.Close()

	scannerTv := bufio.NewScanner(fileTv)

	var tvCatalog = tv.GenerateTvCatalogCheck(common.LANGUAGE_EN)

	for scannerTv.Scan() {
		var tvRead tv.Serie
		json.Unmarshal([]byte(scannerTv.Text()), &tvRead)

		if tvCatalog[tvRead.Id].Id == 0 {
			tv.PopulateSerieByIdAndLanguage(tvRead.Id, common.LANGUAGE_EN)
			go tv.PopulateSerieByIdAndLanguage(tvRead.Id, common.LANGUAGE_PTBR)
		}
	}
	RemoveFile(tvFile + ".json")
	log.Println("FINISH SERIES")

	log.Println("------------------------------------------------------------------------------------------------------------------------------------")

	log.Println("INIT PERSONS")
	downloadExportFile(personFile)
	unzip(personFile)
	filePerson, err := os.Open(personFile + ".json")

	if err != nil {
		log.Fatal(err)
	}

	defer filePerson.Close()

	scannerPerson := bufio.NewScanner(filePerson)

	var personCatalog = person.GeneratePersonCatalogCheck(common.LANGUAGE_EN)

	for scannerPerson.Scan() {

		var personRead person.Person
		json.Unmarshal([]byte(scannerPerson.Text()), &personRead)

		if personCatalog[personRead.Id].Id == 0 {
			person.PopulatePersonByIdAndLanguage(personRead.Id, common.LANGUAGE_EN)
			go person.PopulatePersonByIdAndLanguage(personRead.Id, common.LANGUAGE_PTBR)
		}
	}
	RemoveFile(personFile + ".json")
	log.Println("FINISH PERSONS")

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
