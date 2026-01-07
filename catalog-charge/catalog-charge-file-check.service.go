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

	scannerFile := bufio.NewScanner(fileCatalog)

	// Initialize RabbitMQ connection
	rmq, err := queue.NewRabbitMQ()
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %s", err)
	}
	defer rmq.Close()

	dailyFileIdsSet := make(map[int]bool)

	for scannerFile.Scan() {
		var elementRead tmdb.TmdbDailyFile
		json.Unmarshal([]byte(scannerFile.Text()), &elementRead)
		dailyFileIdsSet[elementRead.Id] = true

		// Check if the ID exists in the generated catalog
		if catalogGenerate[elementRead.Id].Id == 0 {
			// Publish a message
			message := queue.CatalogProcessMessage{Id: elementRead.Id, MediaType: mediaType}
			err = rmq.PublishJSON(queue.QueueCatalogProcess, message)
			if err != nil {
				log.Fatalf("Failed to publish a message: %s", err)
			}

			log.Println("Message published successfully for Id and mediaType: ", message.Id, mediaType)
		}
	}

	for id := range catalogGenerate {
		if !dailyFileIdsSet[id] {
			// ID not found in daily file IDs, remove it from database
			if mediaType == common.MEDIA_TYPE_MOVIE {
				movie.DeleteMovie(id)
				log.Println("Movie removed from catalog: ", id)
			}

			if mediaType == common.MEDIA_TYPE_TV {
				tv.DeleteSerie(id)
				tv.DeleteSerieEpisodes(id)
				log.Println("TV and episodes removed from catalog: ", id)
			}
		}
	}

	util.RemoveFile(fileName + ".json")
	log.Println("====================>FINISH " + mediaType)
}
