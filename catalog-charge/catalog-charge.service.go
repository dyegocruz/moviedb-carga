package catalogCharge

import (
	"bufio"
	"context"
	"encoding/json"
	"log"
	"os"
	"time"

	"moviedb/common"
	"moviedb/movie"
	"moviedb/person"
	"moviedb/queue"
	"moviedb/services"
	"moviedb/tmdb"
	"moviedb/tv"
	"moviedb/util"

	"github.com/olivere/elastic"
)

type Service struct {
	config  services.Config
	mongo   *services.MongoService
	elastic *services.ElasticService
	movie   *movie.Service
	person  *person.Service
	tv      *tv.Service
}

func NewService(config services.Config, mongo *services.MongoService, elasticService *services.ElasticService, movieService *movie.Service, personService *person.Service, tvService *tv.Service) *Service {
	if config == nil {
		config = services.DefaultConfig()
	}

	return &Service{config: config, mongo: mongo, elastic: elasticService, movie: movieService, person: personService, tv: tvService}
}

func (s *Service) CatalogCharge() {
	go s.CheckAndUpdateCatalogByFile(common.MEDIA_TYPE_TV)
	s.CheckAndUpdateCatalogByFile(common.MEDIA_TYPE_MOVIE)
	log.Println("FINISH CatalogCharge")
}

func (s *Service) CatalogUpdates() {
	go s.movie.CheckMoviesChanges()
	s.tv.CheckTvChanges()
	log.Println("FINISH CatalogUpdates")
}

func (s *Service) CheckAndUpdateCatalogByFile(mediaType string) {
	t := time.Now()
	dateFile := t.AddDate(0, 0, -1).Format("01_02_2006")
	mediaFile := ""
	var catalogGenerate map[int]common.CatalogCheck

	switch mediaType {
	case common.MEDIA_TYPE_MOVIE:
		mediaFile = "movie_ids_"
		catalogGenerate = s.movie.GenerateMovieCatalogCheck(common.LANGUAGE_EN)
	case common.MEDIA_TYPE_TV:
		mediaFile = "tv_series_ids_"
		catalogGenerate = s.tv.GenerateTvCatalogCheck(common.LANGUAGE_EN)
	case common.MEDIA_TYPE_PERSON:
		mediaFile = "person_ids_"
		catalogGenerate = s.person.GeneratePersonCatalogCheck(common.LANGUAGE_EN)
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
	rmq, err := services.NewRabbitMQService(s.config)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %s", err)
	}
	defer rmq.Close()

	dailyFileIdsSet := make(map[int]bool)

	for scannerFile.Scan() {
		var elementRead tmdb.TmdbDailyFile
		json.Unmarshal([]byte(scannerFile.Text()), &elementRead)
		dailyFileIdsSet[elementRead.Id] = true

		if catalogGenerate[elementRead.Id].Id == 0 {
			message := queue.CatalogProcessMessage{Id: elementRead.Id, MediaType: mediaType}
			if err := rmq.PublishJSON(queue.QueueCatalogProcess, message); err != nil {
				log.Fatalf("Failed to publish a message: %s", err)
			}

			log.Println("Message published successfully for Id and mediaType: ", message.Id, mediaType)
		}
	}

	for id := range catalogGenerate {
		if !dailyFileIdsSet[id] {
			if mediaType == common.MEDIA_TYPE_MOVIE {
				s.movie.DeleteMovie(id)
				log.Println("Movie removed from catalog: ", id)
			}

			if mediaType == common.MEDIA_TYPE_TV {
				s.tv.DeleteSerie(id)
				s.tv.DeleteSerieEpisodes(id)
				log.Println("TV and episodes removed from catalog: ", id)
			}
		}
	}

	util.RemoveFile(fileName + ".json")
	log.Println("====================>FINISH " + mediaType)
}

func (s *Service) handleCatalogTv(listTvIdsIn []int, newIndexName string, bulkProcessor *elastic.BulkProcessor) {
	docs := s.tv.GetCatalogSearchIn(listTvIdsIn)

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
		req := elastic.NewBulkIndexRequest().Index(newIndexName).Doc(item)
		bulkProcessor.Add(req)
	}
}

func (s *Service) handleCatalogMovie(listMovieIdsIn []int, newIndexName string, bulkProcessor *elastic.BulkProcessor) {
	docs := s.movie.GetCatalogSearchIn(listMovieIdsIn)

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
		req := elastic.NewBulkIndexRequest().Index(newIndexName).Doc(item)
		bulkProcessor.Add(req)
	}
}

func (s *Service) handleCatalogPerson(listPersonIdsIn []int, newIndexName string, bulkProcessor *elastic.BulkProcessor) {
	docs := s.person.GetCatalogSearchIn("en", listPersonIdsIn)
	for _, item := range docs {
		var catalog CatalogSearch
		catalog.Id = item.Id
		catalog.Name = item.Name
		catalog.CatalogType = common.MEDIA_TYPE_PERSON
		catalog.ProfilePath = item.ProfilePath
		catalog.Popularity = item.Popularity
		req := elastic.NewBulkIndexRequest().Index(newIndexName).Doc(catalog)
		bulkProcessor.Add(req)
	}
}

func (s *Service) CatalogSearchCharge() {
	workers := 5
	indexName := "catalog_search"
	elasticClient := s.elastic.Client()
	ctx := context.Background()

	elasticAliasName := indexName
	currentTime := time.Now()
	newIndexName := elasticAliasName + "_" + currentTime.Format("20060102150401")
	log.Println(newIndexName)

	if _, err := elasticClient.CreateIndex(newIndexName).BodyString(INDEX_MAPPING_CATALOG_SEARCH).Do(ctx); err != nil {
		log.Println("Falha ao criar o índice:", newIndexName)
		panic(err)
	}

	bulkProcessor, err := elastic.NewBulkProcessorService(elasticClient).Workers(workers).BulkActions(-1).After(after).Stats(true).Do(ctx)
	if err != nil {
		log.Println("bulkProcessor Error", err)
	}

	idsTv := s.mongo.GetAllIdsByLanguage(services.CollectionSerie, "en")
	log.Println(len(idsTv))

	var listTvIdsIn []int
	for i := 0; i < len(idsTv); i++ {
		listTvIdsIn = append(listTvIdsIn, idsTv[i])
		if i%1000 == 0 {
			s.handleCatalogTv(listTvIdsIn, newIndexName, bulkProcessor)
			listTvIdsIn = []int{}
		}
	}
	if len(listTvIdsIn) > 0 {
		s.handleCatalogTv(listTvIdsIn, newIndexName, bulkProcessor)
	}

	bulkProcessor.Flush()
	bulkProcessor.Close()

	bulkProcessor, err = elastic.NewBulkProcessorService(elasticClient).Workers(workers).BulkActions(-1).After(after).Stats(true).Do(ctx)
	if err != nil {
		log.Println("bulkProcessor Error", err)
	}

	idsMovies := s.mongo.GetAllIdsByLanguage(services.CollectionMovie, "en")
	log.Println(len(idsMovies))

	var listMovieIdsIn []int
	for i := 0; i < len(idsMovies); i++ {
		listMovieIdsIn = append(listMovieIdsIn, idsMovies[i])
		if i%1000 == 0 {
			s.handleCatalogMovie(listMovieIdsIn, newIndexName, bulkProcessor)
			listMovieIdsIn = []int{}
		}
	}
	if len(listMovieIdsIn) > 0 {
		s.handleCatalogMovie(listMovieIdsIn, newIndexName, bulkProcessor)
	}

	bulkProcessor.Flush()
	bulkProcessor.Close()

	bulkProcessor, err = elastic.NewBulkProcessorService(elasticClient).Workers(5).BulkActions(-1).After(after).Stats(true).Do(ctx)
	if err != nil {
		log.Println("bulkProcessor Error", err)
	}

	idsPersons := s.mongo.GetAllIdsByLanguage(services.CollectionPerson, "en")
	log.Println(len(idsPersons))

	var listPersonIdsIn []int
	for i := 0; i < len(idsPersons); i++ {
		listPersonIdsIn = append(listPersonIdsIn, idsPersons[i])
		if i%1000 == 0 {
			s.handleCatalogPerson(listPersonIdsIn, newIndexName, bulkProcessor)
			listPersonIdsIn = []int{}
		}
	}
	if len(listPersonIdsIn) > 0 {
		s.handleCatalogPerson(listPersonIdsIn, newIndexName, bulkProcessor)
	}

	bulkProcessor.Flush()
	bulkProcessor.Close()

	existentSerieAliases, err := s.IndexNamesByAlias(elasticAliasName, elasticClient)
	if err != nil {
		log.Println("Error ao buscar o index no alias: " + elasticAliasName)
	}
	log.Println(existentSerieAliases)

	elasticClient.Alias().Add(newIndexName, elasticAliasName).Do(ctx)

	if len(existentSerieAliases) > 0 {
		oldIndex := existentSerieAliases[0]
		elasticClient.Alias().Remove(oldIndex, elasticAliasName).Do(ctx)
		elasticClient.DeleteIndex(oldIndex).Do(ctx)
	}

	elasticClient.Count(indexName).Do(ctx)
	log.Println("Carga finalizada com sucesso!")
}

func (s *Service) handleElasticChargeInsertDocs(indexName string, listIdsIn []int, newIndexName string, bulkProcessor *elastic.BulkProcessor) {
	var docs []interface{}

	switch indexName {
	case "series":
		docs = s.tv.GetAllByIds(listIdsIn)
	case "movies":
		docs = s.movie.GetAllByIds(listIdsIn)
	case "persons":
		docs = s.person.GetAllByIds(listIdsIn)
	}

	for _, doc := range docs {
		req := elastic.NewBulkIndexRequest().Index(newIndexName).Doc(doc)
		bulkProcessor.Add(req)
	}
}

func (s *Service) ElasticChargeInsert(indexName string, interval int64, mapping string, workers int) {
	elasticClient := s.elastic.Client()
	ctx := context.Background()

	collectionCount := ""
	switch indexName {
	case "series":
		collectionCount = services.CollectionSerie
	case "movies":
		collectionCount = services.CollectionMovie
	case "persons":
		collectionCount = services.CollectionPerson
	}

	docsIds := s.mongo.GetAllIdsByLanguage(collectionCount, "en")
	elasticAliasName := indexName
	currentTime := time.Now()
	newIndexName := elasticAliasName + "_" + currentTime.Format("20060102150401")
	log.Println(newIndexName)

	if _, err := elasticClient.CreateIndex(newIndexName).BodyString(mapping).Do(context.TODO()); err != nil {
		log.Println("Falha ao criar o índice:", newIndexName)
		panic(err)
	}

	bulkProcessor, err := elastic.NewBulkProcessorService(elasticClient).Workers(workers).BulkActions(-1).After(after).Stats(true).Do(ctx)
	if err != nil {
		log.Println("bulkProcessor Error", err)
	}

	var listIdsIn []int
	for i := 0; i < len(docsIds); i++ {
		listIdsIn = append(listIdsIn, docsIds[i])
		if int64(i)%interval == 0 {
			s.handleElasticChargeInsertDocs(indexName, listIdsIn, newIndexName, bulkProcessor)
			listIdsIn = []int{}
		}
	}

	if len(listIdsIn) > 0 {
		s.handleElasticChargeInsertDocs(indexName, listIdsIn, newIndexName, bulkProcessor)
	}

	existentSerieAliases, err := s.IndexNamesByAlias(elasticAliasName, elasticClient)
	if err != nil {
		log.Println("Error ao buscar o index no alias: " + elasticAliasName)
	}
	log.Println(existentSerieAliases)

	elasticClient.Alias().Add(newIndexName, elasticAliasName).Do(ctx)

	if len(existentSerieAliases) > 0 {
		oldIndex := existentSerieAliases[0]
		elasticClient.Alias().Remove(oldIndex, elasticAliasName).Do(ctx)
		elasticClient.DeleteIndex(oldIndex).Do(ctx)
	}

	elasticClient.Count(indexName).Do(ctx)
	log.Println("Carga finalizada com sucesso!")

	bulkProcessor.Flush()
	bulkProcessor.Close()
}

func (s *Service) ElasticGeneralCharge() {
	s.CatalogSearchCharge()
	log.Println("FINISH ElasticGeneralCharge")
}

func (s *Service) GeneralCatalogHandler() {
	s.CatalogCharge()
	s.CatalogUpdates()
}

func (s *Service) IndexNamesByAlias(aliasName string, elasticClient *elastic.Client) ([]string, error) {
	res, err := elasticClient.Aliases().Index("_all").Do(context.TODO())
	if err != nil {
		return nil, err
	}

	return res.IndicesByAlias(aliasName), nil
}

func after(executionID int64, requests []elastic.BulkableRequest, response *elastic.BulkResponse, err error) {
	if err != nil {
		log.Printf("bulk commit failed, err: %v\n", err)
	}
	log.Printf("commit successfully, len(requests)=%d\n", len(requests))
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
)
