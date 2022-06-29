package carga

import (
	"context"
	"log"
	"moviedb/movie"
	"moviedb/person"
	"moviedb/tv"

	"github.com/olivere/elastic"
)

func CargaGeral() {
	movie.PopulateMovies("en", "")
	movie.PopulateMovies("pt-BR", "")

	// // // FILTRA APENAS ANIMAÇÕES
	movie.PopulateMovies("en", "16")
	movie.PopulateMovies("pt-BR", "16")

	tv.PopulateSeries("en", "")
	tv.PopulateSeries("pt-BR", "")

	// // // // FILTRA APENAS ANIMAÇÕES
	tv.PopulateSeries("en", "16")
	tv.PopulateSeries("pt-BR", "16")

	person.PopulatePersons("en")
	person.PopulatePersons("pt-BR")

	// personsWithoutCreditsEn := person.GetPersonsWithoutCredits("en")
	// log.Println("PERSONS EN WITHOUT personsWithoutCreditsEn: ", len(personsWithoutCreditsEn))

	// for _, personUpdate := range personsWithoutCreditsEn {
	// 	person.PopulatePersonByLanguage(personUpdate, "en")
	// }

	// personsWithoutCreditsBr := person.GetPersonsWithoutCredits("pt-BR")
	// log.Println("PERSONS BR WITHOUT personsWithoutCreditsBr: ", len(personsWithoutCreditsBr))

	// for _, personUpdate := range personsWithoutCreditsBr {
	// 	person.PopulatePersonByLanguage(personUpdate, "pt-BR")
	// }

	personsEn := person.GetPersonsWithCredits("en")
	log.Println("PERSONS EN CAST CHECK: ", len(personsEn))

	for _, personEn := range personsEn {
		log.Println("personEn.Credits.Cast: ", len(personEn.Credits.Cast))
		for _, cast := range personEn.Credits.Cast {
			if cast.MediaType == "movie" {
				movieFindEn := movie.GetMovieByIdAndLanguage(cast.ID, "en")
				log.Println("movieFindEn.Id: ", movieFindEn.Id)
				if movieFindEn.Id == 0 {
					movieInsert := movie.GetMovieDetailsOnApiDb(cast.ID, "en")

					movie.PopulateMovieByLanguage(movieInsert, "en")
				}
			}

			if cast.MediaType == "serie" {
				serieFindEn := tv.GetSerieByIdAndLanguage(cast.ID, "en")
				if serieFindEn.Id == 0 {
					serieInsert := tv.GetSerieDetailsOnApiDb(cast.ID, "en")
					tv.PopulateSerieByLanguage(serieInsert, "en")
				}
			}
		}
	}

	personsBr := person.GetPersonsWithCredits("pt-BR")
	log.Println("PERSONS EN CAST CHECK: ", len(personsBr))

	for _, personEn := range personsBr {
		log.Println("personEn.Credits.Cast: ", len(personEn.Credits.Cast))
		for _, cast := range personEn.Credits.Cast {
			if cast.MediaType == "movie" {
				movieFindEn := movie.GetMovieByIdAndLanguage(cast.ID, "pt-BR")
				log.Println("movieFindEn.Id: ", movieFindEn.Id)
				if movieFindEn.Id == 0 {
					movieInsert := movie.GetMovieDetailsOnApiDb(cast.ID, "pt-BR")

					movie.PopulateMovieByLanguage(movieInsert, "pt-BR")
				}
			}

			if cast.MediaType == "serie" {
				serieFindEn := tv.GetSerieByIdAndLanguage(cast.ID, "pt-BR")
				if serieFindEn.Id == 0 {
					serieInsert := tv.GetSerieDetailsOnApiDb(cast.ID, "pt-BR")
					tv.PopulateSerieByLanguage(serieInsert, "pt-BR")
				}
			}
		}
	}

	// moviesCount := movie.GetCountAll()
	// log.Println("Total de filmes: ", moviesCount)

	// var (
	// 	elasticClient *elastic.Client
	// 	err           error
	// )
	// // elasticIndexType := "_doc"

	// elasticClient, err = elastic.NewSimpleClient(
	// 	elastic.SetURL(os.Getenv("ELASTICSEARCH")),
	// 	elastic.SetSniff(false),
	// 	elastic.SetBasicAuth(os.Getenv("ELASTICSEARCH_USER"), os.Getenv("ELASTICSEARCH_PASS")),
	// 	elastic.SetErrorLog(log.New(os.Stderr, "ELASTIC ", log.LstdFlags)),
	// 	elastic.SetInfoLog(log.New(os.Stdout, "LOG: ", log.LstdFlags)),
	// 	// elastic.SetTraceLog(log.New(os.Stdout, "QUERY: ", log.LstdFlags)),
	// )
	// fmt.Println("connect to es success!")
	// if err != nil {
	// 	log.Println(err)
	// 	time.Sleep(3 * time.Second)
	// } else {
	// 	// break
	// }

	// // CONFIGURAÇÃO DO MAPPING DO NOVO INDEX
	// mapping := `{
	// 	"settings":{
	// 		"number_of_shards":1,
	// 		"number_of_replicas":0
	// 	},
	// 	"mappings":{
	// 		"properties":{
	// 			"tags":{
	// 				"type":"keyword"
	// 			},
	// 			"suggest_field":{
	// 				"type":"completion"
	// 			}
	// 		}
	// 	}
	// }`
	// ctx := context.TODO()

	// elasticMovieAliasName := "movies"

	// currentMovieTime := time.Now()
	// var newMovieIndexName = elasticMovieAliasName + "_" + currentMovieTime.Format("20060102150401")
	// log.Println(newMovieIndexName)

	// createMovieIndex, err := elasticClient.CreateIndex(newMovieIndexName).BodyString(mapping).Do(ctx)
	// if err != nil {
	// 	log.Println("Falha ao criar o índice:", newMovieIndexName)
	// 	panic(err)
	// }
	// if !createMovieIndex.Acknowledged {
	// 	// Not acknowledged
	// }

	// var bulkRequest = elasticClient.Bulk()
	// var m int64
	// for m = 0; m < moviesCount; m++ {

	// 	if m%1000 == 0 {
	// 		log.Println(m)
	// 		movies := movie.GetAll(m, 1000)

	// 		for _, movie := range movies {
	// 			// log.Println(m)
	// 			req := elastic.NewBulkIndexRequest().
	// 				Index(newMovieIndexName).
	// 				// Type(elasticIndexType).
	// 				Id(strconv.Itoa(movie.Id) + "-" + movie.Language).
	// 				Doc(movie)

	// 			bulkRequest = bulkRequest.Add(req)
	// 		}

	// 		bulkResponse, err := bulkRequest.Do(ctx)
	// 		if err != nil {
	// 			fmt.Println(err)
	// 		}
	// 		if bulkResponse != nil {

	// 		}
	// 		bulkRequest = elasticClient.Bulk()
	// 	}
	// }

	// // BUSCA SE JÁ EXISTE ALGUM ÍNDICE NO ALIAS DOS FILMES
	// existentMovieAliases, err := IndexNamesByAlias(elasticMovieAliasName, elasticClient)
	// log.Println(existentMovieAliases)

	// // ADICIONA
	// elasticClient.Alias().Add(newMovieIndexName, elasticMovieAliasName).Do(context.TODO())

	// if len(existentMovieAliases) > 0 {
	// 	oldIndex := existentMovieAliases[0]
	// 	elasticClient.Alias().Remove(oldIndex, elasticMovieAliasName).Do(context.TODO())
	// 	elasticClient.DeleteIndex(oldIndex).Do(context.TODO())
	// }
	// log.Println("Carga finalizada com sucesso!")
	// log.Println("Filmes carregados length: ", moviesCount)

	// // ==========> SÉRIEs
	// seriesCount := tv.GetCountAll()
	// log.Println("Total de séries: ", seriesCount)

	// elasticSerieAliasName := "series"

	// currentSerieTime := time.Now()
	// var newSerieIndexName = elasticSerieAliasName + "_" + currentSerieTime.Format("20060102150401")
	// log.Println(newSerieIndexName)

	// createSerieIndex, err := elasticClient.CreateIndex(newSerieIndexName).BodyString(mapping).Do(ctx)
	// if err != nil {
	// 	// Handle error
	// 	// panic(err)
	// 	log.Println("Falha ao criar o índice:", newSerieIndexName)
	// 	panic(err)
	// }
	// if !createSerieIndex.Acknowledged {
	// 	// Not acknowledged
	// }

	// bulkRequest = elasticClient.Bulk()
	// var s int64
	// for s = 0; s < seriesCount; s++ {

	// 	if s%200 == 0 {
	// 		series := tv.GetAll(s, 200)

	// 		for _, serie := range series {
	// 			// log.Println(m)
	// 			req := elastic.NewBulkIndexRequest().
	// 				Index(newSerieIndexName).
	// 				// Type(elasticIndexType).
	// 				Id(strconv.Itoa(serie.Id) + "-" + serie.Language).
	// 				Doc(serie)

	// 			bulkRequest = bulkRequest.Add(req)
	// 		}

	// 		bulkResponse, err := bulkRequest.Do(ctx)
	// 		if err != nil {
	// 			fmt.Println(err)
	// 		}
	// 		if bulkResponse != nil {

	// 		}
	// 		bulkRequest = elasticClient.Bulk()
	// 	}
	// }

	// // BUSCA SE JÁ EXISTE ALGUM ÍNDICE NO ALIAS DE SÉRIES
	// existentSerieAliases, err := IndexNamesByAlias(elasticSerieAliasName, elasticClient)
	// log.Println(existentSerieAliases)

	// // ADICIONA
	// elasticClient.Alias().Add(newSerieIndexName, elasticSerieAliasName).Do(context.TODO())

	// if len(existentSerieAliases) > 0 {
	// 	oldIndex := existentSerieAliases[0]
	// 	elasticClient.Alias().Remove(oldIndex, elasticSerieAliasName).Do(context.TODO())
	// 	elasticClient.DeleteIndex(oldIndex).Do(context.TODO())
	// }
	// log.Println("Carga finalizada com sucesso!")
	// log.Println("Séries carregadas length: ", seriesCount)

	// // // ==========> PESSOAS
	// personsCount := person.GetCountAll()
	// log.Println("Total de pessoas: ", personsCount)

	// elasticPersonAliasName := "persons"

	// currentPersonTime := time.Now()
	// var newPersonIndexName = elasticPersonAliasName + "_" + currentPersonTime.Format("20060102150401")
	// log.Println(newPersonIndexName)

	// createPersonIndex, err := elasticClient.CreateIndex(newPersonIndexName).BodyString(mapping).Do(ctx)
	// if err != nil {
	// 	// Handle error
	// 	// panic(err)
	// 	log.Println("Falha ao criar o índice:", newPersonIndexName)
	// 	panic(err)
	// }
	// if !createPersonIndex.Acknowledged {
	// 	// Not acknowledged
	// }

	// bulkRequest = elasticClient.Bulk()
	// var p int64
	// for p = 0; p < personsCount; p++ {

	// 	if p%1000 == 0 {
	// 		persons := person.GetAll(p, 1000)

	// 		for _, person := range persons {
	// 			// log.Println(m)
	// 			req := elastic.NewBulkIndexRequest().
	// 				Index(newPersonIndexName).
	// 				// Type(elasticIndexType).
	// 				Id(strconv.Itoa(person.Id) + "-" + person.Language).
	// 				Doc(person)

	// 			bulkRequest = bulkRequest.Add(req)
	// 		}

	// 		bulkResponse, err := bulkRequest.Do(ctx)
	// 		if err != nil {
	// 			fmt.Println(err)
	// 		}
	// 		if bulkResponse != nil {

	// 		}
	// 		bulkRequest.Reset()
	// 		// bulkRequest = elasticClient.Bulk()
	// 	}
	// }

	// // BUSCA SE JÁ EXISTE ALGUM ÍNDICE NO ALIAS DE PESSOAS
	// existentPersonAliases, err := IndexNamesByAlias(elasticPersonAliasName, elasticClient)
	// log.Println(existentPersonAliases)

	// // ADICIONA
	// elasticClient.Alias().Add(newPersonIndexName, elasticPersonAliasName).Do(context.TODO())

	// if len(existentPersonAliases) > 0 {
	// 	oldIndex := existentPersonAliases[0]
	// 	elasticClient.Alias().Remove(oldIndex, elasticPersonAliasName).Do(context.TODO())
	// 	elasticClient.DeleteIndex(oldIndex).Do(context.TODO())
	// }
	// log.Println("Carga finalizada com sucesso!")
	// log.Println("Pessoas carregadas length: ", personsCount)
}

func IndexNamesByAlias(aliasName string, elasticClient *elastic.Client) ([]string, error) {
	res, err := elasticClient.Aliases().Index("_all").Do(context.TODO())
	if err != nil {
		return nil, err
	}
	return res.IndicesByAlias(aliasName), nil
}
