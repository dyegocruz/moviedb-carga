package catalogCharge

import (
	"bufio"
	"encoding/json"
	"log"
	"moviedb/common"
	"moviedb/movie"
	"moviedb/person"
	"moviedb/queue"
	"moviedb/tmdb"
	"moviedb/tv"
	"moviedb/util"
	"os"
	"time"
)

func CheckAndUpdateCatalogByFile(mediaType string) {
	t := time.Now()
	dateFile := t.AddDate(0, 0, -1).Format("01_02_2006")
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

  log.Println("====>catalogGenerate", len(catalogGenerate), catalogGenerate[198116].Id)
  
  scannerFile := bufio.NewScanner(fileCatalog)

  // Initialize RabbitMQ connection
	rmq, err := queue.NewRabbitMQ()
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %s", err)
	}
	defer rmq.Close()

	for scannerFile.Scan() {
		var elementRead tmdb.TmdbDailyFile
		json.Unmarshal([]byte(scannerFile.Text()), &elementRead)
		if catalogGenerate[elementRead.Id].Id == 0 {

      // Publish a message
      err = rmq.PublishJSON(queue.QueueCatalogProcess, queue.CatalogProcessMessage{Id: elementRead.Id, MediaType: mediaType})
      if err != nil {
        log.Fatalf("Failed to publish a message: %s", err)
      }

    log.Println("Message published successfully!")
		}
	}

  util.RemoveFile(fileName + ".json")
	log.Println("====================>FINISH " + mediaType)
}
