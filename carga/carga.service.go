package carga

import (
	"context"
	"fmt"
	"log"
	"moviedb/common"
	"moviedb/database"
	"moviedb/movie"
	"moviedb/person"
	"moviedb/tv"
	"os"
	"time"

	"github.com/olivere/elastic"
)

func CatalogCharge() {

	go CheckAndUpdateCatalogByFile(common.MEDIA_TYPE_TV)
	CheckAndUpdateCatalogByFile(common.MEDIA_TYPE_MOVIE)
	// CheckAndUpdateCatalogByFile(common.MEDIA_TYPE_PERSON)

	// go movie.PopulateMovies(common.LANGUAGE_EN, "")

	// // FILTER JUST ANIMATIONS
	// go movie.PopulateMovies(common.LANGUAGE_EN, "16")

	// go tv.PopulateSeries(common.LANGUAGE_EN, "")

	// // FILTER JUST ANIMATIONS
	// go tv.PopulateSeries(common.LANGUAGE_EN, "16")

	// go person.PopulatePersons(common.LANGUAGE_EN)

}

func CatalogUpdates() {
	// Checking changes by data type
	go tv.CheckTvChanges()
	// go person.CheckPersonChanges()
	movie.CheckMoviesChanges()
}

const (
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

func ElasticChargeInsert(indexName string, interval int64, mapping string, bulkActions int) {
	elasticClient := elascitClient(indexName)
	// ctx := context.TODO()

	collectionCount := ""

	switch indexName {
	case "series":
		collectionCount = database.COLLECTION_SERIE
	case "movies":
		collectionCount = database.COLLECTION_MOVIE
	case "persons":
		collectionCount = database.COLLECTION_PERSON
	case "series-episodes":
		collectionCount = database.COLLECTION_SERIE_EPISODE
	}

	// ==========> Elements docs
	docsCount := database.GetCountAllByColletcion(collectionCount)
	log.Println("Total de docs: ", docsCount)

	elasticAliasName := indexName

	currentTime := time.Now()
	var newIndexName = elasticAliasName + "_" + currentTime.Format("20060102150401")
	log.Println(newIndexName)

	_, err := elasticClient.CreateIndex(newIndexName).BodyString(mapping).Do(context.TODO())
	if err != nil {
		log.Println("Falha ao criar o índice:", newIndexName)
		panic(err)
	}

	// bulkProcessor, err := elastic.NewBulkProcessorService(elasticClient).
	// 	// Workers(runtime.NumCPU()).
	// 	Workers(5).
	// 	// BulkActions(-1).
	// 	BulkActions(bulkActions).
	// 	// BulkActions(int(interval) * 2).
	// 	// BulkSize(20 << 20).
	// 	// FlushInterval(1 * time.Second).
	// 	After(after).
	// 	Stats(true).
	// 	Do(ctx)

	if err != nil {
		log.Println("bulkProcessor Error", err)
	}
	var bulkRequest *elastic.BulkService

	switch indexName {
	case "series":
		bulkRequest = elasticClient.Bulk().Index(newIndexName)
		tvCatalog := tv.GenerateTvCatalogCheck(common.LANGUAGE_EN)

		var i int64 = 0
		idsGet := make([]int, 0)
		for _, catalog := range tvCatalog {

			idsGet = append(idsGet, catalog.Id)

			if i%interval == 0 {
				docs := tv.GetByListId(idsGet)

				for _, doc := range docs {
					req := elastic.NewBulkIndexRequest().
						// Index(newIndexName).
						Doc(doc)
					bulkRequest = bulkRequest.Add(req)
				}

				_, err := bulkRequest.Do(context.TODO())
				if err != nil {
					fmt.Println("==================>", err)
					panic("STOP")
				}

				idsGet = make([]int, 0)
				bulkRequest = elasticClient.Bulk().Index(newIndexName)

			}
			i++
		}
	case "movies":
		bulkRequest = elasticClient.Bulk().Index(newIndexName)
		moiveCatalog := movie.GenerateMovieCatalogCheck(common.LANGUAGE_EN)

		var i int64 = 0
		idsGet := make([]int, 0)
		for _, catalog := range moiveCatalog {

			idsGet = append(idsGet, catalog.Id)

			if i%interval == 0 {
				docs := movie.GetByListId(idsGet)

				for _, doc := range docs {
					req := elastic.NewBulkIndexRequest().
						Doc(doc)
					bulkRequest = bulkRequest.Add(req)
				}

				_, err := bulkRequest.Do(context.TODO())
				if err != nil {
					fmt.Println("==================>", err)
					panic("STOP MOVIES")
				}

				idsGet = make([]int, 0)
				bulkRequest = elasticClient.Bulk().Index(newIndexName)

			}
			i++
		}
	case "persons":
		bulkRequest = elasticClient.Bulk().Index(newIndexName)
		personCatalog := person.GeneratePersonCatalogCheck(common.LANGUAGE_EN)

		var i int64 = 0
		idsGet := make([]int, 0)
		for _, catalog := range personCatalog {

			idsGet = append(idsGet, catalog.Id)

			if i%interval == 0 {
				docs := person.GetByListId(idsGet)

				for _, doc := range docs {
					req := elastic.NewBulkIndexRequest().
						Doc(doc)
					bulkRequest = bulkRequest.Add(req)
				}

				_, err := bulkRequest.Do(context.TODO())
				if err != nil {
					fmt.Println("==================>", err)
					panic("STOP PERSONS")
				}

				idsGet = make([]int, 0)
				bulkRequest = elasticClient.Bulk().Index(newIndexName)

			}
			i++
		}
	case "series-episodes":
		bulkRequest = elasticClient.Bulk().Index(newIndexName)
		tvEpisodesCatalog := tv.GenerateTvEpisodesCatalogCheck(common.LANGUAGE_EN)

		var i int64 = 0
		idsGet := make([]int, 0)
		for _, catalog := range tvEpisodesCatalog {

			idsGet = append(idsGet, catalog.Id)

			if i%interval == 0 {
				docs := tv.GetEpisodesByListId(idsGet)

				for _, doc := range docs {
					req := elastic.NewBulkIndexRequest().
						Doc(doc)
					bulkRequest = bulkRequest.Add(req)
				}

				_, err := bulkRequest.Do(context.TODO())
				if err != nil {
					fmt.Println("==================>", err)
					panic("STOP PERSONS")
				}

				idsGet = make([]int, 0)
				bulkRequest = elasticClient.Bulk().Index(newIndexName)

			}
			i++
		}
	}

	// var i int64
	// series := make([]tv.Serie, 0)
	// movies := make([]movie.Movie, 0)
	// persons := make([]person.Person, 0)
	// episodes := make([]tv.Episode, 0)
	// for i = 0; i < docsCount; i++ {
	// 	if i%interval == 0 {
	// 		log.Println(indexName+": ", i)
	// 		switch indexName {
	// 		case "series":
	// 			// series = append(series, tv.GetAll(i, interval)...)
	// 			docs := tv.GetAll(i, interval)
	// 			for _, doc := range docs {
	// 				req := elastic.NewBulkIndexRequest().
	// 					Index(newIndexName).
	// 					Doc(doc)
	// 				bulkProcessor.Add(req)
	// 			}
	// 		case "movies":
	// 			docs := movie.GetAll(i, interval)
	// 			// movies = append(movies, movie.GetAll(i, interval)...)
	// 			for _, doc := range docs {
	// 				req := elastic.NewBulkIndexRequest().
	// 					Index(newIndexName).
	// 					Doc(doc)
	// 				bulkProcessor.Add(req)
	// 			}
	// 		case "persons":
	// 			docs := person.GetAll(i, interval)
	// 			// persons = append(persons, person.GetAll(i, interval)...)
	// 			for _, doc := range docs {
	// 				req := elastic.NewBulkIndexRequest().
	// 					Index(newIndexName).
	// 					Doc(doc)
	// 				bulkProcessor.Add(req)
	// 			}
	// 		case "series-episodes":
	// 			// episodes = append(episodes, tv.GetAllEpisodes(i, interval)...)
	// 			docs := tv.GetAllEpisodes(i, interval)
	// 			for _, doc := range docs {
	// 				req := elastic.NewBulkIndexRequest().
	// 					Index(newIndexName).
	// 					Doc(doc)
	// 				bulkProcessor.Add(req)
	// 			}
	// 		}
	// 	}
	// }

	// for _, serie := range series {
	// 	req := elastic.NewBulkIndexRequest().
	// 		Index(newIndexName).
	// 		Doc(serie)
	// 	bulkProcessor.Add(req)
	// }

	// for _, movie := range movies {
	// 	req := elastic.NewBulkIndexRequest().
	// 		Index(newIndexName).
	// 		Doc(movie)
	// 	bulkProcessor.Add(req)
	// }

	// for _, person := range persons {
	// 	req := elastic.NewBulkIndexRequest().
	// 		Index(newIndexName).
	// 		Doc(person)
	// 	bulkProcessor.Add(req)
	// }

	// for _, episode := range episodes {
	// 	req := elastic.NewBulkIndexRequest().
	// 		Index(newIndexName).
	// 		Doc(episode)
	// 	bulkProcessor.Add(req)
	// }

	// BUSCA SE JÁ EXISTE ALGUM ÍNDICE NO ALIAS DE SÉRIES
	existentSerieAliases, err := IndexNamesByAlias(elasticAliasName, elasticClient)
	if err != nil {
		log.Println("Error ao buscar o index no alias: " + elasticAliasName)
	}
	log.Println(existentSerieAliases)

	// ADICIONA
	elasticClient.Alias().Add(newIndexName, elasticAliasName).Do(context.TODO())

	if len(existentSerieAliases) > 0 {
		oldIndex := existentSerieAliases[0]
		elasticClient.Alias().Remove(oldIndex, elasticAliasName).Do(context.TODO())
		elasticClient.DeleteIndex(oldIndex).Do(context.TODO())
	}

	// this method is called just to update the quantity of docs inserted on elastic search
	elasticClient.Count(indexName).Do(context.TODO())

	log.Println("Carga finalizada com sucesso!")

	// bulkProcessor.Flush()
	// bulkProcessor.Close()
}

func ElasticGeneralCharge() {
	go ElasticChargeInsert("series", 200, INDEX_MAPPING_SERIES, 100)
	go ElasticChargeInsert("movies", 200, INDEX_MAPPING_MOVIES, 100)
	ElasticChargeInsert("persons", 5000, INDEX_MAPPING_PERSONS, 100)
	// ElasticChargeInsert("series-episodes", 2000, INDEX_MAPPING_SERIES_EPISODE, 100)
}

func GeneralCharge() {
	// CatalogCharge()
	// CatalogUpdates()
	ElasticGeneralCharge()
}

func IndexNamesByAlias(aliasName string, elasticClient *elastic.Client) ([]string, error) {
	res, err := elasticClient.Aliases().Index("_all").Do(context.TODO())
	if err != nil {
		return nil, err
	}
	return res.IndicesByAlias(aliasName), nil
}
