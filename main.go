package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"moviedb/database"
	"moviedb/movie"
	"moviedb/person"
	"moviedb/tv"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/olivere/elastic"
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
	movie.Populate("en", "")
	movie.Populate("pt-BR", "")

	// FILTRA APENAS ANIMAÇÕES
	movie.Populate("en", "16")
	movie.Populate("pt-BR", "16")

	tv.Populate("en", "")
	tv.Populate("pt-BR", "")

	// FILTRA APENAS ANIMAÇÕES
	tv.Populate("en", "16")
	tv.Populate("pt-BR", "16")

	person.Populate("en")
	person.Populate("pt-BR")

	movies := movie.GetAll()
	log.Println(movies)

	var (
		elasticClient *elastic.Client
		err           error
	)
	elasticIndexType := "_doc"

	elasticClient, err = elastic.NewSimpleClient(
		elastic.SetURL(os.Getenv("ELASTICSEARCH")),
		elastic.SetSniff(false),
		elastic.SetBasicAuth(os.Getenv("ELASTICSEARCH_USER"), os.Getenv("ELASTICSEARCH_PASS")),
		elastic.SetErrorLog(log.New(os.Stderr, "ELASTIC ", log.LstdFlags)),
		elastic.SetInfoLog(log.New(os.Stdout, "LOG: ", log.LstdFlags)),
		// elastic.SetTraceLog(log.New(os.Stdout, "QUERY", log.LstdFlags)),
	)
	fmt.Println("connect to es success!")
	if err != nil {
		log.Println(err)
		time.Sleep(3 * time.Second)
	} else {
		// break
	}

	// CONFIGURAÇÃO DO MAPPING DO NOVO INDEX
	mapping := `{
		"settings":{
			"number_of_shards":1,
			"number_of_replicas":0
		},
		"mappings":{
			"properties":{
				"tags":{
					"type":"keyword"
				},
				"suggest_field":{
					"type":"completion"
				}
			}
		}
	}`
	ctx := context.TODO()

	elasticMovieAliasName := "movies"

	currentMovieTime := time.Now()
	var newMovieIndexName = elasticMovieAliasName + "_" + currentMovieTime.Format("20060102150401")
	log.Println(newMovieIndexName)

	createMovieIndex, err := elasticClient.CreateIndex(newMovieIndexName).BodyString(mapping).Do(ctx)
	if err != nil {
		// Handle error
		// panic(err)
		log.Println("Falha ao criar o índice:", newMovieIndexName)
		panic(err)
	}
	if !createMovieIndex.Acknowledged {
		// Not acknowledged
	}

	for _, movie := range movies {
		// CONVERTE O STRUCT DO MOVIE PARA UMA STRING JSON
		movieJSONByte, err := json.Marshal(movie)
		if err != nil {
			fmt.Println(err)
			return
		}

		// Iterate over the docs and index them one-by-one
		_, err = elasticClient.Index().
			Index(newMovieIndexName).
			Type(elasticIndexType).            // unique doctype now deprecated
			BodyString(string(movieJSONByte)). // Usando o JSON, pois o método BodyJson estava bagunçando o valor enviado
			// Omit this if you want dynamically generated _id
			// Id(strconv.Itoa(id)). // Convert int to string
			Id(strconv.Itoa(movie.Id) + "-" + movie.Language). // DEFININDO O _ID DO PRESTADRO NO ELASTIC COMO SENDO O COD_PRESTADOR DO SABIUS
			Do(ctx)
		if err != nil {
			log.Println("ERROR", err)
		}
	}

	// BUSCA SE JÁ EXISTE ALGUM ÍNDICE NO ALIAS DO GUIA MÉDICO
	existentMovieAliases, err := IndexNamesByAlias(elasticMovieAliasName, elasticClient)
	log.Println(existentMovieAliases)

	// ADICIONA
	elasticClient.Alias().Add(newMovieIndexName, elasticMovieAliasName).Do(context.TODO())

	if len(existentMovieAliases) > 0 {
		oldIndex := existentMovieAliases[0]
		elasticClient.Alias().Remove(oldIndex, elasticMovieAliasName).Do(context.TODO())
		elasticClient.DeleteIndex(oldIndex).Do(context.TODO())
	}
	log.Println("Carga finalizada com sucesso!")
	log.Println("Filmes carregados length: ", len(movies))

	// ==========> SÉRIEs
	series := tv.GetAll()
	log.Println(series)

	elasticSerieAliasName := "series"

	currentSerieTime := time.Now()
	var newSerieIndexName = elasticSerieAliasName + "_" + currentSerieTime.Format("20060102150401")
	log.Println(newSerieIndexName)

	createSerieIndex, err := elasticClient.CreateIndex(newSerieIndexName).BodyString(mapping).Do(ctx)
	if err != nil {
		// Handle error
		// panic(err)
		log.Println("Falha ao criar o índice:", newSerieIndexName)
		panic(err)
	}
	if !createSerieIndex.Acknowledged {
		// Not acknowledged
	}

	for _, serie := range series {
		// CONVERTE O STRUCT DO serie PARA UMA STRING JSON
		serieJSONByte, err := json.Marshal(serie)
		if err != nil {
			fmt.Println(err)
			return
		}

		// Iterate over the docs and index them one-by-one
		_, err = elasticClient.Index().
			Index(newSerieIndexName).
			Type(elasticIndexType).            // unique doctype now deprecated
			BodyString(string(serieJSONByte)). // Usando o JSON, pois o método BodyJson estava bagunçando o valor enviado
			// Omit this if you want dynamically generated _id
			// Id(strconv.Itoa(id)). // Convert int to string
			Id(strconv.Itoa(serie.Id) + "-" + serie.Language). // DEFININDO O _ID DO PRESTADRO NO ELASTIC COMO SENDO O COD_PRESTADOR DO SABIUS
			Do(ctx)
		if err != nil {
			log.Println("ERROR", err)
		}
	}

	// BUSCA SE JÁ EXISTE ALGUM ÍNDICE NO ALIAS DO GUIA MÉDICO
	existentSerieAliases, err := IndexNamesByAlias(elasticSerieAliasName, elasticClient)
	log.Println(existentSerieAliases)

	// ADICIONA
	elasticClient.Alias().Add(newSerieIndexName, elasticSerieAliasName).Do(context.TODO())

	if len(existentSerieAliases) > 0 {
		oldIndex := existentSerieAliases[0]
		elasticClient.Alias().Remove(oldIndex, elasticSerieAliasName).Do(context.TODO())
		elasticClient.DeleteIndex(oldIndex).Do(context.TODO())
	}
	log.Println("Carga finalizada com sucesso!")
	log.Println("Séries carregados length: ", len(series))

	// ==========> PESSOAS
	persons := person.GetAll()
	log.Println(persons)

	elasticPersonAliasName := "persons"

	currentPersonTime := time.Now()
	var newPersonIndexName = elasticPersonAliasName + "_" + currentPersonTime.Format("20060102150401")
	log.Println(newPersonIndexName)

	createPersonIndex, err := elasticClient.CreateIndex(newPersonIndexName).BodyString(mapping).Do(ctx)
	if err != nil {
		// Handle error
		// panic(err)
		log.Println("Falha ao criar o índice:", newPersonIndexName)
		panic(err)
	}
	if !createPersonIndex.Acknowledged {
		// Not acknowledged
	}

	for _, person := range persons {
		// CONVERTE O STRUCT DO person PARA UMA STRING JSON
		personJSONByte, err := json.Marshal(person)
		if err != nil {
			fmt.Println(err)
			return
		}

		// Iterate over the docs and index them one-by-one
		_, err = elasticClient.Index().
			Index(newPersonIndexName).
			Type(elasticIndexType).             // unique doctype now deprecated
			BodyString(string(personJSONByte)). // Usando o JSON, pois o método BodyJson estava bagunçando o valor enviado
			// Omit this if you want dynamically generated _id
			// Id(strconv.Itoa(id)). // Convert int to string
			Id(strconv.Itoa(person.Id) + "-" + person.Language). // DEFININDO O _ID DO PRESTADRO NO ELASTIC COMO SENDO O COD_PRESTADOR DO SABIUS
			Do(ctx)
		if err != nil {
			log.Println("ERROR", err)
		}
	}

	// BUSCA SE JÁ EXISTE ALGUM ÍNDICE NO ALIAS DO GUIA MÉDICO
	existentPersonAliases, err := IndexNamesByAlias(elasticPersonAliasName, elasticClient)
	log.Println(existentPersonAliases)

	// ADICIONA
	elasticClient.Alias().Add(newPersonIndexName, elasticPersonAliasName).Do(context.TODO())

	if len(existentPersonAliases) > 0 {
		oldIndex := existentPersonAliases[0]
		elasticClient.Alias().Remove(oldIndex, elasticPersonAliasName).Do(context.TODO())
		elasticClient.DeleteIndex(oldIndex).Do(context.TODO())
	}
	log.Println("Carga finalizada com sucesso!")
	log.Println("Séries carregados length: ", len(persons))

	// r := gin.Default()
	// r.GET("/ping", func(c *gin.Context) {
	// 	c.JSON(200, gin.H{
	// 		"message": "pong",
	// 	})
	// })
	// r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}

func IndexNamesByAlias(aliasName string, elasticClient *elastic.Client) ([]string, error) {
	res, err := elasticClient.Aliases().Index("_all").Do(context.TODO())
	if err != nil {
		return nil, err
	}
	return res.IndicesByAlias(aliasName), nil
}
