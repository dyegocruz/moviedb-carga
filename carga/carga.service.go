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
	"strconv"
	"time"

	"github.com/olivere/elastic"
)

func CatalogCharge() {

	go CheckAndUpdateCatalogByFile(common.MEDIA_TYPE_TV)
	go CheckAndUpdateCatalogByFile(common.MEDIA_TYPE_MOVIE)
	CheckAndUpdateCatalogByFile(common.MEDIA_TYPE_PERSON)

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
	go movie.CheckMoviesChanges()
	go person.CheckPersonChanges()
	tv.CheckTvChanges()
}

func ElasticChargeMovies(elasticClient *elastic.Client, interval int64) {
	moviesCount := database.GetCountAllByColletcion(database.COLLECTION_MOVIE)
	log.Println("Total de filmes: ", moviesCount)

	// CONFIGURAÇÃO DO MAPPING DO NOVO INDEX
	mapping := `{
    "settings":{
      "number_of_shards":1,
      "number_of_replicas":0
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
	ctx := context.TODO()

	elasticMovieAliasName := "movies"

	currentMovieTime := time.Now()
	var newMovieIndexName = elasticMovieAliasName + "_" + currentMovieTime.Format("20060102150401")
	log.Println(newMovieIndexName)

	_, err := elasticClient.CreateIndex(newMovieIndexName).BodyString(mapping).Do(ctx)
	if err != nil {
		log.Println("Falha ao criar o índice:", newMovieIndexName)
		panic(err)
	}

	var bulkRequest = elasticClient.Bulk()
	var m int64
	for m = 0; m < moviesCount; m++ {

		if m%interval == 0 {
			// log.Println(m)
			movies := movie.GetAll(m, interval)

			for _, movie := range movies {
				// log.Println(m)
				req := elastic.NewBulkIndexRequest().
					Index(newMovieIndexName).
					// Type(elasticIndexType).
					Id(strconv.Itoa(movie.Id) + "-" + movie.Language).
					Doc(movie)

				bulkRequest = bulkRequest.Add(req)
			}

			_, err := bulkRequest.Do(ctx)
			if err != nil {
				fmt.Println(err)
			}
			bulkRequest = elasticClient.Bulk()
		}
	}

	// BUSCA SE JÁ EXISTE ALGUM ÍNDICE NO ALIAS DOS FILMES
	existentMovieAliases, err := IndexNamesByAlias(elasticMovieAliasName, elasticClient)
	log.Println(existentMovieAliases)
	if err != nil {
		log.Println("Error ao buscar o index no alias: " + elasticMovieAliasName)
	}

	// ADICIONA
	elasticClient.Alias().Add(newMovieIndexName, elasticMovieAliasName).Do(context.TODO())

	if len(existentMovieAliases) > 0 {
		oldIndex := existentMovieAliases[0]
		elasticClient.Alias().Remove(oldIndex, elasticMovieAliasName).Do(context.TODO())
		elasticClient.DeleteIndex(oldIndex).Do(context.TODO())
	}
	log.Println("Carga finalizada com sucesso!")
	log.Println("Filmes carregados length: ", moviesCount)
}

func ElasticChargeTv(elasticClient *elastic.Client, interval int64) {
	// CONFIGURAÇÃO DO MAPPING DO NOVO INDEX
	mapping := `{
    "settings":{
      "number_of_shards":1,
      "number_of_replicas":0
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
	ctx := context.TODO()

	// ==========> SÉRIEs
	seriesCount := database.GetCountAllByColletcion(database.COLLECTION_SERIE)
	log.Println("Total de séries: ", seriesCount)

	elasticSerieAliasName := "series"

	currentSerieTime := time.Now()
	var newSerieIndexName = elasticSerieAliasName + "_" + currentSerieTime.Format("20060102150401")
	log.Println(newSerieIndexName)

	_, err := elasticClient.CreateIndex(newSerieIndexName).BodyString(mapping).Do(ctx)
	if err != nil {
		// Handle error
		// panic(err)
		log.Println("Falha ao criar o índice:", newSerieIndexName)
		panic(err)
	}

	var bulkRequest = elasticClient.Bulk()
	var s int64
	for s = 0; s < seriesCount; s++ {

		if s%interval == 0 {
			series := tv.GetAll(s, interval)

			for _, serie := range series {
				// log.Println(m)
				req := elastic.NewBulkIndexRequest().
					Index(newSerieIndexName).
					// Type(elasticIndexType).
					Id(strconv.Itoa(serie.Id) + "-" + serie.Language).
					Doc(serie)

				bulkRequest = bulkRequest.Add(req)
			}

			_, err := bulkRequest.Do(ctx)
			if err != nil {
				fmt.Println(err)
			}

			bulkRequest = elasticClient.Bulk()
		}
	}

	// BUSCA SE JÁ EXISTE ALGUM ÍNDICE NO ALIAS DE SÉRIES
	existentSerieAliases, err := IndexNamesByAlias(elasticSerieAliasName, elasticClient)
	if err != nil {
		log.Println("Error ao buscar o index no alias: " + elasticSerieAliasName)
	}
	log.Println(existentSerieAliases)

	// ADICIONA
	elasticClient.Alias().Add(newSerieIndexName, elasticSerieAliasName).Do(context.TODO())

	if len(existentSerieAliases) > 0 {
		oldIndex := existentSerieAliases[0]
		elasticClient.Alias().Remove(oldIndex, elasticSerieAliasName).Do(context.TODO())
		elasticClient.DeleteIndex(oldIndex).Do(context.TODO())
	}
	log.Println("Carga finalizada com sucesso!")
	log.Println("Séries carregadas length: ", seriesCount)
}

func ElasticChargeTvEpisodes(elasticClient *elastic.Client, interval int64) {

	// CONFIGURAÇÃO DO MAPPING DO NOVO INDEX
	mapping := `{
    "settings":{
      "number_of_shards":1,
      "number_of_replicas":0
    },
    "mappings":{
      "properties":{
        "language":{
          "type":"text"
        }
      }
    }
  }`
	ctx := context.TODO()

	// ==========> EPISODES
	seriesEpisodesCount := database.GetCountAllByColletcion(database.COLLECTION_SERIE_EPISODE)
	log.Println("Total de episódios: ", seriesEpisodesCount)

	elasticSerieEpisodeAliasName := "series-episodes"

	currentSerieTime := time.Now()
	var newSerieEpisodeIndexName = elasticSerieEpisodeAliasName + "_" + currentSerieTime.Format("20060102150401")
	log.Println(newSerieEpisodeIndexName)

	_, err := elasticClient.CreateIndex(newSerieEpisodeIndexName).BodyString(mapping).Do(ctx)
	if err != nil {
		// Handle error
		// panic(err)
		log.Println("Falha ao criar o índice:", newSerieEpisodeIndexName)
		panic(err)
	}

	var bulkRequest = elasticClient.Bulk()
	var s int64
	for s = 0; s < seriesEpisodesCount; s++ {

		if s%interval == 0 {
			episodes := tv.GetAllEpisodes(s, interval)

			for _, episode := range episodes {
				// log.Println(m)
				req := elastic.NewBulkIndexRequest().
					Index(newSerieEpisodeIndexName).
					// Type(elasticIndexType).
					Id(strconv.Itoa(episode.Id) + "-" + episode.Language).
					Doc(episode)

				bulkRequest = bulkRequest.Add(req)
			}

			_, err := bulkRequest.Do(ctx)
			if err != nil {
				fmt.Println(err)
			}
			bulkRequest = elasticClient.Bulk()
		}
	}

	// BUSCA SE JÁ EXISTE ALGUM ÍNDICE NO ALIAS DE SÉRIES
	existentSerieAliases, err := IndexNamesByAlias(elasticSerieEpisodeAliasName, elasticClient)
	if err != nil {
		log.Println("Error ao buscar o index no alias: " + elasticSerieEpisodeAliasName)
	}
	log.Println(existentSerieAliases)

	// ADICIONA
	elasticClient.Alias().Add(newSerieEpisodeIndexName, elasticSerieEpisodeAliasName).Do(context.TODO())

	if len(existentSerieAliases) > 0 {
		oldIndex := existentSerieAliases[0]
		elasticClient.Alias().Remove(oldIndex, elasticSerieEpisodeAliasName).Do(context.TODO())
		elasticClient.DeleteIndex(oldIndex).Do(context.TODO())
	}
	log.Println("Carga finalizada com sucesso!")
	log.Println("Episódios carregadas length: ", seriesEpisodesCount)
}

func ElasticChargePerson(elasticClient *elastic.Client, interval int64) {

	// CONFIGURAÇÃO DO MAPPING DO NOVO INDEX
	mapping := `{
    "settings":{
      "number_of_shards":1,
      "number_of_replicas":0
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
	ctx := context.TODO()

	// ==========> PESSOAS
	personsCount := database.GetCountAllByColletcion(database.COLLECTION_PERSON)
	log.Println("Total de pessoas: ", personsCount)

	elasticPersonAliasName := "persons"

	currentPersonTime := time.Now()
	var newPersonIndexName = elasticPersonAliasName + "_" + currentPersonTime.Format("20060102150401")
	log.Println(newPersonIndexName)

	_, err := elasticClient.CreateIndex(newPersonIndexName).BodyString(mapping).Do(ctx)
	if err != nil {
		log.Println("Falha ao criar o índice:", newPersonIndexName)
		panic(err)
	}

	var bulkRequest = elasticClient.Bulk()
	var p int64
	for p = 0; p < personsCount; p++ {

		if p%interval == 0 {
			persons := person.GetAll(p, interval)

			for _, person := range persons {
				// log.Println(m)
				req := elastic.NewBulkIndexRequest().
					Index(newPersonIndexName).
					// Type(elasticIndexType).
					Id(strconv.Itoa(person.Id) + "-" + person.Language).
					Doc(person)

				bulkRequest = bulkRequest.Add(req)
			}

			_, err := bulkRequest.Do(ctx)
			if err != nil {
				fmt.Println(err)
			}
			bulkRequest.Reset()
			// bulkRequest = elasticClient.Bulk()
		}
	}

	// BUSCA SE JÁ EXISTE ALGUM ÍNDICE NO ALIAS DE PESSOAS
	existentPersonAliases, err := IndexNamesByAlias(elasticPersonAliasName, elasticClient)
	log.Println(existentPersonAliases)
	if err != nil {
		log.Println("Error ao buscar o index no alias: " + elasticPersonAliasName)
	}

	// ADICIONA
	elasticClient.Alias().Add(newPersonIndexName, elasticPersonAliasName).Do(context.TODO())

	if len(existentPersonAliases) > 0 {
		oldIndex := existentPersonAliases[0]
		elasticClient.Alias().Remove(oldIndex, elasticPersonAliasName).Do(context.TODO())
		elasticClient.DeleteIndex(oldIndex).Do(context.TODO())
	}
	log.Println("Carga finalizada com sucesso!")
	log.Println("Pessoas carregadas length: ", personsCount)
}

func elascitClient(logString string) *elastic.Client {
	elasticClient, err := elastic.NewSimpleClient(
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

func ElasticGeneralCharge() {
	go ElasticChargeTv(elascitClient("TV"), 70)
	go ElasticChargePerson(elascitClient("PERSONS"), 4000)
	go ElasticChargeTvEpisodes(elascitClient("TV_EPISODES"), 3000)
	ElasticChargeMovies(elascitClient("MOVIES"), 60)
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
