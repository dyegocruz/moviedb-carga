package catalogCharge

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"moviedb/common"
	"moviedb/configs"
	"moviedb/database"
	"moviedb/movie"
	"moviedb/person"
	"moviedb/queue"
	"moviedb/tv"
	"os"
	"time"

	"github.com/olivere/elastic"
)

func CatalogCharge() {
	go CheckAndUpdateCatalogByFile(common.MEDIA_TYPE_TV)
	CheckAndUpdateCatalogByFile(common.MEDIA_TYPE_MOVIE)
	log.Println("FINISH CatalogCharge")
}

func CatalogUpdates() {
  go movie.CheckMoviesChanges()
	tv.CheckTvChanges()	
	log.Println("FINISH CatalogUpdates")
}

const (
	INDEX_MAPPING_CATALOG_SEARCH = `{	  
    "settings": {
      "number_of_shards" : 1,
			"number_of_replicas" : 0,
      "analysis": {
        "analyzer": {
          "default": { 
            "type": "custom",
            "tokenizer": "standard",
            "filter": [
              "lowercase",
              "asciifolding"
            ]
          }
        }
      }
    },
	  "mappings":{
	    "properties":{        
				"search_field": {
					"type": "text",
          "analyzer": "default"
				},
				"locations.title": {
					"type": "text",
          "analyzer": "default", 
					"copy_to": "search_field"
				},
				"name": {
					"type": "text",
					"copy_to": "search_field"
				},
				"originalTitle": {
					"type": "text",
					"copy_to": "search_field"
				},
	      "popularity":{
	        "type":"double"
	      }
	    }
	  }
	}`
	INDEX_MAPPING_SERIES = `{
	  "settings":{
	    "number_of_shards" : 1,
			"number_of_replicas" : 0
	  },
	  "mappings":{
	    "properties":{
				"search_field": {
					"type": "text"
				},
				"title": {
					"type": "text",
					"copy_to": "search_field"
				},
				"original_title": {
					"type": "text",
					"copy_to": "search_field"
				},
				"slug": {
					"type": "text",
					"copy_to": "search_field"
				},
	      "popularity":{
	        "type":"double"
	      }
	    }
	  }
	}`
	INDEX_MAPPING_MOVIES = `{
	  "settings":{
	    "number_of_shards" : 1,
			"number_of_replicas" : 0
	  },
	  "mappings":{
	    "properties":{
				"search_field": {
					"type": "text"
				},
				"title": {
					"type": "text",
					"copy_to": "search_field"
				},
				"original_title": {
					"type": "text",
					"copy_to": "search_field"
				},
	      "popularity":{
	        "type":"double"
	      }
	    }
	  }
	}`
	INDEX_MAPPING_PERSONS = `{
	  "settings":{
	    "number_of_shards" : 1,
			"number_of_replicas" : 0
	  },
	  "mappings":{
	    "properties":{
				"search_field": {
					"type": "text"
				},
				"name": {
					"type": "text",
					"copy_to": "search_field"
				},
				"biography": {
					"type": "text",
					"copy_to": "search_field"
				},
	      "popularity":{
	        "type":"double"
	      }
	    }
	  }
	}`
	INDEX_MAPPING_SERIES_EPISODE = `{
	  "settings":{
	    "number_of_shards" : 1,
			"number_of_replicas" : 0
	  },
	  "mappings":{
	    "properties":{
	      "language":{
	        "type":"text"
	      }
	    }
	  }
	}`
)

func elascitClient(logString string) *elastic.Client {
	elasticClient, err := elastic.NewClient(
		elastic.SetURL(os.Getenv("ELASTICSEARCH")),
		elastic.SetSniff(false),
		elastic.SetBasicAuth(os.Getenv("ELASTICSEARCH_USER"), os.Getenv("ELASTICSEARCH_PASS")),
		elastic.SetErrorLog(log.New(os.Stderr, "ELASTIC ", log.LstdFlags)),
		elastic.SetInfoLog(log.New(os.Stdout, logString+": ", log.LstdFlags)),
		// elastic.SetTraceLog(log.New(os.Stdout, "QUERY: ", log.LstdFlags)),
	)
	fmt.Println("connect to es success!")
	if err != nil {
		log.Println(err)
		time.Sleep(3 * time.Second)
	} else {
		// break
	}

	return elasticClient
}

func after(executionID int64, requests []elastic.BulkableRequest, response *elastic.BulkResponse, err error) {
	if err != nil {
		log.Printf("bulk commit failed, err: %v\n", err)
	}
	// do what ever you want in case bulk commit success
	log.Printf("commit successfully, len(requests)=%d\n", len(requests))
}

func handleCatalogTv(listTvIdsIn []int, newIndexName string, bulkProcessor *elastic.BulkProcessor) {
	docs := tv.GetCatalogSearchIn(listTvIdsIn)

	catalogTvLocalizated := make(map[int]CatalogSearch, 0)
	for _, item := range docs {
		var catalog CatalogSearch
		if catalogTvLocalizated[item.Id].Id == 0 {
			catalog.Id = item.Id
			catalog.CatalogType = common.MEDIA_TYPE_TV
			catalog.ReleaseDate = item.FirstAirDate
			catalog.OriginalLanguage = item.OriginalLanguage
			catalog.OriginalTitle = item.OriginalTitle
			catalog.Popularity = item.Popularity
			catalogTvLocalizated[item.Id] = catalog
		}

		var location Location
		location.Language = item.Language
		location.Title = item.Title
		location.PosterPath = item.PosterPath

		loc := catalogTvLocalizated[item.Id]
		loc.Locations = append(loc.Locations, location)
		catalogTvLocalizated[item.Id] = loc
	}

	for _, item := range catalogTvLocalizated {
		req := elastic.NewBulkIndexRequest().
			Index(newIndexName).
			Doc(item)
		bulkProcessor.Add(req)
	}
}

func handleCatalogMovie(listMovieIdsIn []int, newIndexName string, bulkProcessor *elastic.BulkProcessor) {
	docs := movie.GetCatalogSearchIn(listMovieIdsIn)

	catalogMovieLocalizated := make(map[int]CatalogSearch, 0)
	for _, item := range docs {
		var catalog CatalogSearch
		if catalogMovieLocalizated[item.Id].Id == 0 {
			catalog.Id = item.Id
			catalog.CatalogType = common.MEDIA_TYPE_MOVIE
			catalog.ReleaseDate = item.ReleaseDate
			catalog.OriginalLanguage = item.OriginalLanguage
			catalog.OriginalTitle = item.OriginalTitle
			catalog.Popularity = item.Popularity
			catalogMovieLocalizated[item.Id] = catalog
		}

		var location Location
		location.Language = item.Language
		location.Title = item.Title
		location.PosterPath = item.PosterPath

		loc := catalogMovieLocalizated[item.Id]
		loc.Locations = append(loc.Locations, location)
		catalogMovieLocalizated[item.Id] = loc
	}

	for _, item := range catalogMovieLocalizated {
		req := elastic.NewBulkIndexRequest().
			Index(newIndexName).
			Doc(item)
		bulkProcessor.Add(req)
	}
}

func handleCatalogPerson(listPersonIdsIn []int, newIndexName string, bulkProcessor *elastic.BulkProcessor) {
	docs := person.GetCatalogSearchIn("en", listPersonIdsIn)
	for _, item := range docs {
		var catalog CatalogSearch
		catalog.Id = item.Id
		catalog.Name = item.Name
		catalog.CatalogType = common.MEDIA_TYPE_PERSON
		catalog.ProfilePath = item.ProfilePath
		catalog.Popularity = item.Popularity
		req := elastic.NewBulkIndexRequest().
			Index(newIndexName).
			Doc(catalog)
		bulkProcessor.Add(req)
	}
}

func CatalogSearchCharge() {
	workers := 5
	indexName := "catalog_search"
	elasticClient := elascitClient(indexName)
	ctx := context.Background()

	elasticAliasName := indexName
	currentTime := time.Now()
	var newIndexName = elasticAliasName + "_" + currentTime.Format("20060102150401")
	log.Println(newIndexName)

	_, err := elasticClient.CreateIndex(newIndexName).BodyString(INDEX_MAPPING_CATALOG_SEARCH).Do(ctx)
	if err != nil {
		log.Println("Falha ao criar o índice:", newIndexName)
		panic(err)
	}

	bulkProcessor, err := elastic.NewBulkProcessorService(elasticClient).
		Workers(workers).
		BulkActions(-1).
		After(after).
		Stats(true).
		Do(ctx)
	if err != nil {
		log.Println("bulkProcessor Error", err)
	}

	// CATALOG SEARCH TV
	idsTv := database.GetAllIdsByLanguage(database.COLLECTION_SERIE, "en")
	log.Println(len(idsTv))

	var iTv int64 = 0
	var intervalTv = int64(1000)
	var listTvIdsIn []int = []int{}
	for iTv = 0; iTv < int64(len(idsTv)); iTv++ {
		listTvIdsIn = append(listTvIdsIn, idsTv[iTv])
		if iTv%intervalTv == 0 {
			handleCatalogTv(listTvIdsIn, newIndexName, bulkProcessor)
			listTvIdsIn = []int{}
		}
	}
	if len(listTvIdsIn) > 0 {
		handleCatalogTv(listTvIdsIn, newIndexName, bulkProcessor)
	}

	bulkProcessor.Flush()
	bulkProcessor.Close()

	bulkProcessor, err = elastic.NewBulkProcessorService(elasticClient).
		Workers(workers).
		BulkActions(-1).
		After(after).
		Stats(true).
		Do(ctx)
	if err != nil {
		log.Println("bulkProcessor Error", err)
	}

	// CATALOG SEARCH MOVIE
	idsMovies := database.GetAllIdsByLanguage(database.COLLECTION_MOVIE, "en")
	log.Println(len(idsMovies))

	var iMovie int64 = 0
	var intervalMovies = int64(1000)
	var listMovieIdsIn []int = []int{}
	for iMovie = 0; iMovie < int64(len(idsMovies)); iMovie++ {
		listMovieIdsIn = append(listMovieIdsIn, idsMovies[iMovie])
		if iMovie%intervalMovies == 0 {
			handleCatalogMovie(listMovieIdsIn, newIndexName, bulkProcessor)
			listMovieIdsIn = []int{}
		}
	}

	if len(listMovieIdsIn) > 0 {
		handleCatalogMovie(listMovieIdsIn, newIndexName, bulkProcessor)
	}

	bulkProcessor.Flush()
	bulkProcessor.Close()

	bulkProcessor, err = elastic.NewBulkProcessorService(elasticClient).
		Workers(5).
		BulkActions(-1).
		After(after).
		Stats(true).
		Do(ctx)
	if err != nil {
		log.Println("bulkProcessor Error", err)
	}

	idsPersons := database.GetAllIdsByLanguage(database.COLLECTION_PERSON, "en")
	log.Println(len(idsPersons))

	var iPerson int64 = 0
	var intervalPerson = int64(1000)
	var listPersonIdsIn []int = []int{}
	for iPerson = 0; iPerson < int64(len(idsPersons)); iPerson++ {
		listPersonIdsIn = append(listPersonIdsIn, idsPersons[iPerson])
		if iPerson%intervalPerson == 0 {
			handleCatalogPerson(listPersonIdsIn, newIndexName, bulkProcessor)
			listPersonIdsIn = []int{}
		}
	}

	if len(listPersonIdsIn) > 0 {
		handleCatalogPerson(listPersonIdsIn, newIndexName, bulkProcessor)
	}

	bulkProcessor.Flush()
	bulkProcessor.Close()

	// BUSCA SE JÁ EXISTE ALGUM ÍNDICE NO ALIAS DE SÉRIES
	existentSerieAliases, err := IndexNamesByAlias(elasticAliasName, elasticClient)
	if err != nil {
		log.Println("Error ao buscar o index no alias: " + elasticAliasName)
	}
	log.Println(existentSerieAliases)

	// ADICIONA
	elasticClient.Alias().Add(newIndexName, elasticAliasName).Do(ctx)

	if len(existentSerieAliases) > 0 {
		oldIndex := existentSerieAliases[0]
		elasticClient.Alias().Remove(oldIndex, elasticAliasName).Do(ctx)
		elasticClient.DeleteIndex(oldIndex).Do(ctx)
	}

	// this method is called just to update the quantity of docs inserted on elastic search
	elasticClient.Count(indexName).Do(ctx)

	log.Println("Carga finalizada com sucesso!")
}

func handleElasticChargeInsertDocs(indexName string, listIdsIn []int, newIndexName string, bulkProcessor *elastic.BulkProcessor) {
	var docs []interface{}

	switch indexName {
	case "series":
		docs = tv.GetAllByIds(listIdsIn)
	case "movies":
		docs = movie.GetAllByIds(listIdsIn)
	case "persons":
		docs = person.GetAllByIds(listIdsIn)
	}

	for _, doc := range docs {
		req := elastic.NewBulkIndexRequest().
			Index(newIndexName).
			Doc(doc)
		bulkProcessor.Add(req)
	}
}

func ElasticChargeInsert(indexName string, interval int64, mapping string, workers int) {
	elasticClient := elascitClient(indexName)
	ctx := context.Background()

	collectionCount := ""

	switch indexName {
	case "series":
		collectionCount = database.COLLECTION_SERIE
	case "movies":
		collectionCount = database.COLLECTION_MOVIE
	case "persons":
		collectionCount = database.COLLECTION_PERSON
	}

	docsIds := database.GetAllIdsByLanguage(collectionCount, "en")

	elasticAliasName := indexName

	currentTime := time.Now()
	var newIndexName = elasticAliasName + "_" + currentTime.Format("20060102150401")
	log.Println(newIndexName)

	_, err := elasticClient.CreateIndex(newIndexName).BodyString(mapping).Do(context.TODO())
	if err != nil {
		log.Println("Falha ao criar o índice:", newIndexName)
		panic(err)
	}

	bulkProcessor, err := elastic.NewBulkProcessorService(elasticClient).
		Workers(workers).
		BulkActions(-1).
		After(after).
		Stats(true).
		Do(ctx)
	if err != nil {
		log.Println("bulkProcessor Error", err)
	}

	var i int64
	var listIdsIn []int = []int{}
	for i = 0; i < int64(len(docsIds)); i++ {
		listIdsIn = append(listIdsIn, docsIds[i])
		if i%interval == 0 {
			handleElasticChargeInsertDocs(indexName, listIdsIn, newIndexName, bulkProcessor)
			listIdsIn = []int{}
		}
	}

	if len(listIdsIn) > 0 {
		handleElasticChargeInsertDocs(indexName, listIdsIn, newIndexName, bulkProcessor)
	}

	// BUSCA SE JÁ EXISTE ALGUM ÍNDICE NO ALIAS DE SÉRIES
	existentSerieAliases, err := IndexNamesByAlias(elasticAliasName, elasticClient)
	if err != nil {
		log.Println("Error ao buscar o index no alias: " + elasticAliasName)
	}
	log.Println(existentSerieAliases)

	// ADICIONA
	elasticClient.Alias().Add(newIndexName, elasticAliasName).Do(ctx)

	if len(existentSerieAliases) > 0 {
		oldIndex := existentSerieAliases[0]
		elasticClient.Alias().Remove(oldIndex, elasticAliasName).Do(ctx)
		elasticClient.DeleteIndex(oldIndex).Do(ctx)
	}

	// this method is called just to update the quantity of docs inserted on elastic search
	elasticClient.Count(indexName).Do(ctx)

	log.Println("Carga finalizada com sucesso!")

	bulkProcessor.Flush()
	bulkProcessor.Close()
}

func ElasticGeneralCharge() {
  go ElasticChargeInsert("series", 10000, INDEX_MAPPING_SERIES, 3)
  go ElasticChargeInsert("movies", 10000, INDEX_MAPPING_MOVIES, 3)
  go ElasticChargeInsert("persons", 10000, INDEX_MAPPING_PERSONS, 5)
	CatalogSearchCharge()

	log.Println("FINISH ElasticGeneralCharge")
}

func GeneralCharge() {
	CatalogCharge()
	CatalogUpdates()
	SendMessageProcessCatalogConcluded()
}

func SendMessageProcessCatalogConcluded() {
	bytes, err := json.Marshal(queue.EsChargeMessage{
		UpdatesCompleted: true,
		Text:             "The update process was completed. Starting ES charge.",
		Env:              configs.GetEnv(),
	})
	if err != nil {
		panic(err)
	}

	fmt.Println(string(bytes))

	messageBody := string(bytes)
	err = queue.SendMessage(configs.GetQueueUrl(), messageBody)
	if err != nil {
		fmt.Printf("Got an error while trying to send message to queue: %v", err)
		return
	}

	fmt.Println("Message sent successfully")
}

func IndexNamesByAlias(aliasName string, elasticClient *elastic.Client) ([]string, error) {
	res, err := elasticClient.Aliases().Index("_all").Do(context.TODO())
	if err != nil {
		return nil, err
	}
	return res.IndicesByAlias(aliasName), nil
}
