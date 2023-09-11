package carga

import (
	"bufio"
	"encoding/json"
	"log"
	"moviedb/common"
	"moviedb/movie"
	"moviedb/person"
	"moviedb/tmdb"
	"moviedb/tv"
	"moviedb/util"
	"os"
	"time"
)

func CheckAndUpdateCatalogByFile(mediaType string) {
	t := time.Now()
	dateFile := t.Format("01_02_2006")
	mediaFile := ""
	var catalogGenerate map[int]common.CatalogCheck

	switch mediaType {
	case common.MEDIA_TYPE_MOVIE:
		mediaFile = "movie_ids_"
		catalogGenerate = movie.GenerateMovieCatalogCheck(common.LANGUAGE_EN)
	case common.MEDIA_TYPE_TV:
		mediaFile = "tv_series_ids_"
		catalogGenerate = tv.GenerateTvCatalogCheck(common.LANGUAGE_EN)
	case common.MEDIA_TYPE_PERSON:
		mediaFile = "person_ids_"
		catalogGenerate = person.GeneratePersonCatalogCheck(common.LANGUAGE_EN)
	}

	fileName := mediaFile + dateFile

	log.Println("====================>INIT " + mediaType)
	util.DownloadExportFile("http://files.tmdb.org/p/exports", fileName)
	util.Unzip(fileName)

	fileCatalog, err := os.Open(fileName + ".json")
	if err != nil {
		log.Fatal(err)
	}
	defer fileCatalog.Close()

	scannerFile := bufio.NewScanner(fileCatalog)

	for scannerFile.Scan() {
		var elementRead tmdb.TmdbDailyFile
		json.Unmarshal([]byte(scannerFile.Text()), &elementRead)
		if catalogGenerate[elementRead.Id].Id == 0 {
			log.Println("INSERT: "+fileName, elementRead.Id)

			switch mediaType {
			case common.MEDIA_TYPE_MOVIE:
				movie.PopulateMovieByIdAndLanguage(elementRead.Id, common.LANGUAGE_EN, "Y")
				go movie.PopulateMovieByIdAndLanguage(elementRead.Id, common.LANGUAGE_PTBR, "Y")
			case common.MEDIA_TYPE_TV:
				tv.PopulateSerieByIdAndLanguage(elementRead.Id, common.LANGUAGE_EN)
				go tv.PopulateSerieByIdAndLanguage(elementRead.Id, common.LANGUAGE_PTBR)
			case common.MEDIA_TYPE_PERSON:
				person.PopulatePersonByIdAndLanguage(elementRead.Id, common.LANGUAGE_EN, "N")
				go person.PopulatePersonByIdAndLanguage(elementRead.Id, common.LANGUAGE_PTBR, "N")
			}
		}
	}

	util.RemoveFile(fileName + ".json")
	log.Println("FINISH " + mediaFile)
}
