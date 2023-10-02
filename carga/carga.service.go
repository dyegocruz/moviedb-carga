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
}

func CatalogUpdates() {
	go tv.CheckTvChanges()
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
	ctx := context.Background()

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

	// var bulkRequest *elastic.BulkService

	switch indexName {
	case "series":

		bulkProcessor, err := elastic.NewBulkProcessorService(elasticClient).
			Workers(3).
			BulkActions(-1).
			After(after).
			Stats(true).
			Do(ctx)
		if err != nil {
			log.Println("bulkProcessor Error", err)
		}

		var i int64
		for i = 0; i < docsCount; i++ {

			if i%interval == 0 {
				docs := tv.GetAll(i, interval)
				log.Println(i, len(docs))
				for _, doc := range docs {
					req := elastic.NewBulkIndexRequest().
						Index(newIndexName).
						Doc(doc)
					bulkProcessor.Add(req)
				}
			}
		}

		bulkProcessor.Flush()
		bulkProcessor.Close()
	case "movies":

		bulkProcessor, err := elastic.NewBulkProcessorService(elasticClient).
			Workers(3).
			BulkActions(-1).
			After(after).
			Stats(true).
			Do(ctx)

		if err != nil {
			log.Println("bulkProcessor Error", err)
		}

		var i int64
		for i = 0; i < docsCount; i++ {

			if i%interval == 0 {
				docs := movie.GetAll(i, interval)
				log.Println(i, len(docs))
				for _, doc := range docs {
					req := elastic.NewBulkIndexRequest().
						Index(newIndexName).
						Doc(doc)
					bulkProcessor.Add(req)
				}
			}
		}
		bulkProcessor.Flush()
		bulkProcessor.Close()
	case "persons":

		bulkProcessor, err := elastic.NewBulkProcessorService(elasticClient).
			Workers(5).
			BulkActions(-1).
			After(after).
			Stats(true).
			Do(ctx)

		if err != nil {
			log.Println("bulkProcessor Error", err)
		}

		var i int64
		for i = 0; i < docsCount; i++ {

			if i%interval == 0 {
				docs := person.GetAll(i, interval)
				log.Println(i, len(docs))
				for _, doc := range docs {
					req := elastic.NewBulkIndexRequest().
						Index(newIndexName).
						Doc(doc)
					bulkProcessor.Add(req)
				}
			}
		}

		bulkProcessor.Flush()
		bulkProcessor.Close()
		// case "series-episodes":
		// 	bulkRequest = elasticClient.Bulk().Index(newIndexName)
		// 	tvEpisodesCatalog := tv.GenerateTvEpisodesCatalogCheck(common.LANGUAGE_EN)

		// 	var i int64 = 0
		// 	idsGet := make([]int, 0)
		// 	for _, catalog := range tvEpisodesCatalog {

		// 		idsGet = append(idsGet, catalog.Id)

		// 		if i%interval == 0 {
		// 			docs := tv.GetEpisodesByListId(idsGet)

		// 			for _, doc := range docs {
		// 				req := elastic.NewBulkIndexRequest().
		// 					Doc(doc)
		// 				bulkRequest = bulkRequest.Add(req)
		// 			}

		// 			_, err := bulkRequest.Do(context.TODO())
		// 			if err != nil {
		// 				fmt.Println("==================>", err)
		// 				panic("STOP PERSONS")
		// 			}

		// 			idsGet = make([]int, 0)
		// 			bulkRequest = elasticClient.Bulk().Index(newIndexName)

		// 		}
		// 		i++
		// 	}
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

	// bulkProcessor.Flush()
	// bulkProcessor.Close()
}

func ElasticGeneralCharge() {
	go ElasticChargeInsert("series", 10000, INDEX_MAPPING_SERIES, 100)
	go ElasticChargeInsert("movies", 10000, INDEX_MAPPING_MOVIES, 100)
	ElasticChargeInsert("persons", 10000, INDEX_MAPPING_PERSONS, 100)
}

func GeneralCharge() {
	CatalogCharge()
	CatalogUpdates()
	ElasticGeneralCharge()
}

func IndexNamesByAlias(aliasName string, elasticClient *elastic.Client) ([]string, error) {
	res, err := elasticClient.Aliases().Index("_all").Do(context.TODO())
	if err != nil {
		return nil, err
	}
	return res.IndicesByAlias(aliasName), nil
}
