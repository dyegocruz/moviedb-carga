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

type bulkAdder interface {
	Add(r elastic.BulkableRequest)
}

type Service struct {
	config  services.Config
	mongo   *services.MongoService
	elastic *services.ElasticService
	movie   *movie.Service
	person  *person.Service
	tv      *tv.Service
}

func NewService(
	config services.Config,
	mongo *services.MongoService,
	elasticService *services.ElasticService,
	movieService *movie.Service,
	personService *person.Service,
	tvService *tv.Service,
) *Service {
	if config == nil {
		config = services.DefaultConfig()
	}

	return &Service{config: config,
		mongo: mongo, elastic: elasticService, movie: movieService, person: personService, tv: tvService}
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
	mediaFile := mediaFilePrefix(mediaType)
	var catalogGenerate map[int]bool

	switch mediaType {
	case common.MEDIA_TYPE_MOVIE:
		catalogGenerate = s.movie.GenerateMovieCatalogCheck(common.LANGUAGE_PTBR)
	case common.MEDIA_TYPE_TV:
		catalogGenerate = s.tv.GenerateTvCatalogCheck(common.LANGUAGE_PTBR)
	case common.MEDIA_TYPE_PERSON:
		catalogGenerate = s.person.GeneratePersonCatalogCheck(common.LANGUAGE_PTBR)
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

	for scannerFile.Scan() {
		var elementRead tmdb.TmdbDailyFile
		json.Unmarshal([]byte(scannerFile.Text()), &elementRead)

		if _, isInCatalog := catalogGenerate[elementRead.Id]; isInCatalog {
			catalogGenerate[elementRead.Id] = true
			continue
		}

		message := queue.CatalogProcessMessage{Id: elementRead.Id, MediaType: mediaType}
		if err := rmq.PublishJSON(queue.QueueCatalogProcess, message); err != nil {
			log.Fatalf("Failed to publish a message: %s", err)
		}
		log.Println("Message published successfully for Id and mediaType: ", message.Id, mediaType)
	}

	for id, isInCatalog := range catalogGenerate {

		if !isInCatalog {
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

func (s *Service) handleCatalogTv(listTvIdsIn []int, newIndexName string, bulkProcessor bulkAdder) {
	docs := s.tv.GetCatalogSearchIn(listTvIdsIn)
	for _, item := range buildCatalogTvLocalized(docs) {
		req := elastic.NewBulkIndexRequest().Index(newIndexName).Doc(item)
		bulkProcessor.Add(req)
	}
}

func (s *Service) handleCatalogMovie(listMovieIdsIn []int, newIndexName string, bulkProcessor bulkAdder) {
	docs := s.movie.GetCatalogSearchIn(listMovieIdsIn)
	for _, item := range buildCatalogMovieLocalized(docs) {
		req := elastic.NewBulkIndexRequest().Index(newIndexName).Doc(item)
		bulkProcessor.Add(req)
	}
}

func (s *Service) handleCatalogPerson(listPersonIdsIn []int, newIndexName string, bulkProcessor bulkAdder) {
	docs := s.person.GetCatalogSearchIn(common.LANGUAGE_PTBR, listPersonIdsIn)
	for _, catalog := range buildCatalogPersonDocs(docs) {
		req := elastic.NewBulkIndexRequest().Index(newIndexName).Doc(catalog)
		bulkProcessor.Add(req)
	}
}

func buildCatalogTvLocalized(docs []tv.Serie) []CatalogSearch {
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

		loc := catalogTvLocalizated[item.Id]

		for _, localization := range item.Localizations {
			var location Location
			location.Language = localization.Locale
			location.Title = localization.Title
			if localization.PosterPath != "" {
				location.PosterPath = localization.PosterPath
			} else {
				location.PosterPath = item.PosterPath
			}

			loc.Locations = append(loc.Locations, location)
		}

		catalogTvLocalizated[item.Id] = loc
	}

	catalogList := make([]CatalogSearch, 0, len(catalogTvLocalizated))
	for _, item := range catalogTvLocalizated {
		catalogList = append(catalogList, item)
	}

	return catalogList
}

func buildCatalogMovieLocalized(docs []movie.Movie) []CatalogSearch {
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

		loc := catalogMovieLocalizated[item.Id]

		if len(item.Localizations) > 0 {
			for _, localization := range item.Localizations {
				var location Location
				location.Language = localization.Locale
				location.Title = localization.Title
				if localization.PosterPath != "" {
					location.PosterPath = localization.PosterPath
				} else {
					location.PosterPath = item.PosterPath
				}

				loc.Locations = append(loc.Locations, location)
			}
		} else {
			var location Location
			location.Language = item.Language
			location.Title = item.Title
			location.PosterPath = item.PosterPath
			loc.Locations = append(loc.Locations, location)
		}

		catalogMovieLocalizated[item.Id] = loc
	}

	catalogList := make([]CatalogSearch, 0, len(catalogMovieLocalizated))
	for _, item := range catalogMovieLocalizated {
		catalogList = append(catalogList, item)
	}

	return catalogList
}

func buildCatalogPersonDocs(docs []person.Person) []CatalogSearch {
	catalogList := make([]CatalogSearch, 0, len(docs))
	for _, person := range docs {

		personCatalog := CatalogSearch{
			Id:          person.Id,
			Name:        person.Name,
			CatalogType: common.MEDIA_TYPE_PERSON,
			ProfilePath: person.ProfilePath,
			Popularity:  person.Popularity,
		}

		if len(person.Localizations) > 0 {
			personCatalog.Name = person.Localizations[0].Name
			for _, localization := range person.Localizations {
				var location Location
				location.Language = localization.Locale
				location.Title = localization.Name
				personCatalog.Locations = append(personCatalog.Locations, location)
			}
		} else {
			personCatalog.Name = person.Name
		}

		catalogList = append(catalogList, personCatalog)
	}

	return catalogList
}

func (s *Service) CatalogSearchCharge() {
	workers := 5
	indexName := "catalog_search"
	newIndexName := indexName + "_" + time.Now().Format("20060102150401")
	log.Println(newIndexName)

	elasticClient := s.elastic.Client()
	ctx := context.Background()
	if _, err := elasticClient.CreateIndex(newIndexName).BodyString(INDEX_MAPPING_CATALOG_SEARCH).Do(ctx); err != nil {
		log.Println("Falha ao criar o índice:", newIndexName)
		panic(err)
	}

	bulkProcessor, err := newBulkProcessor(elasticClient, ctx, workers)
	if err != nil {
		log.Println("bulkProcessor TV Error", err)
	} else {
		ids := s.mongo.GetAllIdsByLanguage(services.CollectionSerie, common.LANGUAGE_PTBR)
		log.Println(len(ids))
		processIDsInBatches(ids, 1000, func(batch []int) {
			s.handleCatalogTv(batch, newIndexName, bulkProcessor)
		})
		bulkProcessor.Flush()
		bulkProcessor.Close()
	}

	bulkProcessor, err = newBulkProcessor(elasticClient, ctx, workers)
	if err != nil {
		log.Println("bulkProcessor Movie Error", err)
	} else {
		ids := s.mongo.GetAllIdsByLanguage(services.CollectionMovie, common.LANGUAGE_PTBR)
		log.Println(len(ids))
		processIDsInBatches(ids, 1000, func(batch []int) {
			s.handleCatalogMovie(batch, newIndexName, bulkProcessor)
		})
		bulkProcessor.Flush()
		bulkProcessor.Close()
	}

	bulkProcessor, err = newBulkProcessor(elasticClient, ctx, workers)
	if err != nil {
		log.Println("bulkProcessor Person Error", err)
	} else {
		ids := s.mongo.GetAllIdsByLanguage(services.CollectionPerson, common.LANGUAGE_PTBR)
		log.Println(len(ids))
		processIDsInBatches(ids, 1000, func(batch []int) {
			s.handleCatalogPerson(batch, newIndexName, bulkProcessor)
		})
		bulkProcessor.Flush()
		bulkProcessor.Close()
	}

	existingAliases, err := s.IndexNamesByAlias(indexName, elasticClient)
	if err != nil {
		log.Println("Error ao buscar o index no alias: " + indexName)
	}
	log.Println(existingAliases)

	rotateCatalogAlias(elasticClient, ctx, newIndexName, indexName, existingAliases)

	log.Println("Carga finalizada com sucesso!")
}

func newBulkProcessor(client *elastic.Client, ctx context.Context, workers int) (*elastic.BulkProcessor, error) {
	return elastic.NewBulkProcessorService(client).
		Workers(workers).
		BulkActions(1000).
		After(after).
		Do(ctx)
}

func rotateCatalogAlias(client *elastic.Client, ctx context.Context, newIndexName string, aliasName string, existingAliases []string) {
	_, _ = client.Alias().Add(newIndexName, aliasName).Do(ctx)
	if oldIndex, ok := oldIndexToRetire(existingAliases); ok {
		_, _ = client.Alias().Remove(oldIndex, aliasName).Do(ctx)
		_, _ = client.DeleteIndex(oldIndex).Do(ctx)
	}
	_, _ = client.Count(aliasName).Do(ctx)
}

func (s *Service) handleElasticChargeInsertDocs(indexName string, listIdsIn []int, newIndexName string, bulkProcessor bulkAdder) {
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
	collectionCount := collectionByIndexName(indexName)
	docsIDs := s.mongo.GetAllIdsByLanguage(collectionCount, common.LANGUAGE_PTBR)
	newIndexName := indexName + "_" + time.Now().Format("20060102150401")
	log.Println(newIndexName)

	elasticClient := s.elastic.Client()
	ctx := context.Background()
	if _, err := elasticClient.CreateIndex(newIndexName).BodyString(mapping).Do(ctx); err != nil {
		log.Println("Falha ao criar o índice:", newIndexName)
		panic(err)
	}

	bulkProcessor, err := newBulkProcessor(elasticClient, ctx, workers)
	if err != nil {
		log.Println("bulkProcessor Error", err)
	} else {
		processIDsInBatchesWithInterval(docsIDs, interval, func(batch []int) {
			s.handleElasticChargeInsertDocs(indexName, batch, newIndexName, bulkProcessor)
		})
		bulkProcessor.Flush()
		bulkProcessor.Close()
	}

	existingAliases, err := s.IndexNamesByAlias(indexName, elasticClient)
	if err != nil {
		log.Println("Error ao buscar o index no alias: " + indexName)
	}
	log.Println(existingAliases)

	rotateCatalogAlias(elasticClient, ctx, newIndexName, indexName, existingAliases)
	log.Println("Carga finalizada com sucesso!")
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

func shouldPublishCatalogMessage(id int, catalogGenerate map[int]common.CatalogCheck) bool {
	return catalogGenerate[id].Id == 0
}

func idsMissingFromDaily(catalogGenerate map[int]common.CatalogCheck, dailyFileIdsSet map[int]bool) []int {
	missing := make([]int, 0)
	for id := range catalogGenerate {
		if !dailyFileIdsSet[id] {
			missing = append(missing, id)
		}
	}

	return missing
}

func collectionByIndexName(indexName string) string {
	switch indexName {
	case "series":
		return services.CollectionSerie
	case "movies":
		return services.CollectionMovie
	case "persons":
		return services.CollectionPerson
	default:
		return ""
	}
}

func mediaFilePrefix(mediaType string) string {
	switch mediaType {
	case common.MEDIA_TYPE_MOVIE:
		return "movie_ids_"
	case common.MEDIA_TYPE_TV:
		return "tv_series_ids_"
	case common.MEDIA_TYPE_PERSON:
		return "person_ids_"
	default:
		return ""
	}
}

func oldIndexToRetire(existing []string) (string, bool) {
	if len(existing) == 0 {
		return "", false
	}
	return existing[0], true
}

func processIDsInBatches(ids []int, interval int64, handleBatch func(batch []int)) {
	processIDsInBatchesWithInterval(ids, interval, handleBatch)
}

func processIDsInBatchesWithInterval(ids []int, interval int64, handleBatch func(batch []int)) {
	var batch []int
	for _, id := range ids {
		batch = append(batch, id)
		if interval > 0 && int64(len(batch)) >= interval {
			handleBatch(batch)
			batch = nil
		}
	}

	if len(batch) > 0 {
		handleBatch(batch)
	}
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
