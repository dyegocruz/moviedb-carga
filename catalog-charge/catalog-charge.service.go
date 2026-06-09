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

type rabbitPublisher interface {
	Close()
	PublishJSON(queueName string, data interface{}) error
}

type Service struct {
	config                        services.Config
	mongo                         *services.MongoService
	elastic                       *services.ElasticService
	movie                         *movie.Service
	person                        *person.Service
	tv                            *tv.Service
	checkAndUpdateCatalogByFileFn func(mediaType string)
	checkMoviesChangesFn          func()
	checkTvChangesFn              func()
	catalogSearchChargeFn         func()
	catalogChargeFn               func()
	catalogUpdatesFn              func()
	getTvCatalogSearchInFn        func(ids []int) []tv.Serie
	getMovieCatalogSearchInFn     func(ids []int) []movie.Movie
	getPersonCatalogSearchInFn    func(language string, ids []int) []person.Person
	getTvAllByIdsFn               func(ids []int) []interface{}
	getMovieAllByIdsFn            func(ids []int) []interface{}
	getPersonAllByIdsFn           func(ids []int) []interface{}
	nowFn                         func() time.Time
	downloadExportFileFn          func(url string, fileName string)
	unzipFn                       func(fileName string)
	openFileFn                    func(name string) (*os.File, error)
	newRabbitMQServiceFn          func(config services.Config) (rabbitPublisher, error)
	removeFileFn                  func(name string)
	generateMovieCatalogCheckFn   func(language string) map[int]common.CatalogCheck
	generateTvCatalogCheckFn      func(language string) map[int]common.CatalogCheck
	generatePersonCatalogCheckFn  func(language string) map[int]common.CatalogCheck
	deleteMovieFn                 func(id int)
	deleteSerieFn                 func(id int)
	deleteSerieEpisodesFn         func(id int)
	getAllIdsByLanguageFn         func(collection string, language string) []int
	catalogSearchChargeExecutorFn func()
	catalogSearchChargeRunFn      func(workers int, indexName string, newIndexName string) error
	elasticChargeInsertExecutorFn func(indexName string, interval int64, mapping string, workers int)
	elasticChargeInsertRunFn      func(indexName string, interval int64, workers int, mapping string, newIndexName string, docsIDs []int) error
}

func NewService(config services.Config, mongo *services.MongoService, elasticService *services.ElasticService, movieService *movie.Service, personService *person.Service, tvService *tv.Service) *Service {
	if config == nil {
		config = services.DefaultConfig()
	}

	return &Service{config: config, mongo: mongo, elastic: elasticService, movie: movieService, person: personService, tv: tvService}
}

func (s *Service) CatalogCharge() {
	runCheckAndUpdate := s.CheckAndUpdateCatalogByFile
	if s.checkAndUpdateCatalogByFileFn != nil {
		runCheckAndUpdate = s.checkAndUpdateCatalogByFileFn
	}

	go runCheckAndUpdate(common.MEDIA_TYPE_TV)
	runCheckAndUpdate(common.MEDIA_TYPE_MOVIE)
	log.Println("FINISH CatalogCharge")
}

func (s *Service) CatalogUpdates() {
	runMoviesChanges := func() {
		s.movie.CheckMoviesChanges()
	}
	runTvChanges := func() {
		s.tv.CheckTvChanges()
	}

	if s.checkMoviesChangesFn != nil {
		runMoviesChanges = s.checkMoviesChangesFn
	}
	if s.checkTvChangesFn != nil {
		runTvChanges = s.checkTvChangesFn
	}

	go runMoviesChanges()
	runTvChanges()
	log.Println("FINISH CatalogUpdates")
}

func (s *Service) CheckAndUpdateCatalogByFile(mediaType string) {
	now := s.nowFn
	if now == nil {
		now = time.Now
	}
	download := s.downloadExportFileFn
	if download == nil {
		download = util.DownloadExportFile
	}
	unzip := s.unzipFn
	if unzip == nil {
		unzip = util.Unzip
	}
	openFile := s.openFileFn
	if openFile == nil {
		openFile = os.Open
	}
	newRabbit := s.newRabbitMQServiceFn
	if newRabbit == nil {
		newRabbit = func(config services.Config) (rabbitPublisher, error) {
			return services.NewRabbitMQService(config)
		}
	}
	removeFile := s.removeFileFn
	if removeFile == nil {
		removeFile = util.RemoveFile
	}

	genMovieCatalogCheck := func(language string) map[int]common.CatalogCheck {
		return s.movie.GenerateMovieCatalogCheck(language)
	}
	if s.generateMovieCatalogCheckFn != nil {
		genMovieCatalogCheck = s.generateMovieCatalogCheckFn
	}
	genTvCatalogCheck := func(language string) map[int]common.CatalogCheck {
		return s.tv.GenerateTvCatalogCheck(language)
	}
	if s.generateTvCatalogCheckFn != nil {
		genTvCatalogCheck = s.generateTvCatalogCheckFn
	}
	genPersonCatalogCheck := func(language string) map[int]common.CatalogCheck {
		return s.person.GeneratePersonCatalogCheck(language)
	}
	if s.generatePersonCatalogCheckFn != nil {
		genPersonCatalogCheck = s.generatePersonCatalogCheckFn
	}
	deleteMovie := func(id int) { s.movie.DeleteMovie(id) }
	if s.deleteMovieFn != nil {
		deleteMovie = s.deleteMovieFn
	}
	deleteSerie := func(id int) { s.tv.DeleteSerie(id) }
	if s.deleteSerieFn != nil {
		deleteSerie = s.deleteSerieFn
	}
	deleteSerieEpisodes := func(id int) { s.tv.DeleteSerieEpisodes(id) }
	if s.deleteSerieEpisodesFn != nil {
		deleteSerieEpisodes = s.deleteSerieEpisodesFn
	}

	t := now()
	dateFile := t.AddDate(0, 0, -1).Format("01_02_2006")
	mediaFile := mediaFilePrefix(mediaType)
	var catalogGenerate map[int]common.CatalogCheck

	switch mediaType {
	case common.MEDIA_TYPE_MOVIE:
		catalogGenerate = genMovieCatalogCheck(common.LANGUAGE_EN)
	case common.MEDIA_TYPE_TV:
		catalogGenerate = genTvCatalogCheck(common.LANGUAGE_EN)
	case common.MEDIA_TYPE_PERSON:
		catalogGenerate = genPersonCatalogCheck(common.LANGUAGE_EN)
	}

	fileName := mediaFile + dateFile

	log.Println("====================>INIT " + mediaType)
	download("http://files.tmdb.org/p/exports", fileName)
	unzip(fileName)

	fileCatalog, err := openFile(fileName + ".json")
	if err != nil {
		log.Fatal(err)
	}
	defer fileCatalog.Close()

	scannerFile := bufio.NewScanner(fileCatalog)
	rmq, err := newRabbit(s.config)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %s", err)
	}
	defer rmq.Close()

	dailyFileIdsSet := make(map[int]bool)

	for scannerFile.Scan() {
		var elementRead tmdb.TmdbDailyFile
		json.Unmarshal([]byte(scannerFile.Text()), &elementRead)
		dailyFileIdsSet[elementRead.Id] = true

		if shouldPublishCatalogMessage(elementRead.Id, catalogGenerate) {
			message := queue.CatalogProcessMessage{Id: elementRead.Id, MediaType: mediaType}
			if err := rmq.PublishJSON(queue.QueueCatalogProcess, message); err != nil {
				log.Fatalf("Failed to publish a message: %s", err)
			}

			log.Println("Message published successfully for Id and mediaType: ", message.Id, mediaType)
		}
	}

	for _, id := range idsMissingFromDaily(catalogGenerate, dailyFileIdsSet) {
		if mediaType == common.MEDIA_TYPE_MOVIE {
			deleteMovie(id)
			log.Println("Movie removed from catalog: ", id)
		}

		if mediaType == common.MEDIA_TYPE_TV {
			deleteSerie(id)
			deleteSerieEpisodes(id)
			log.Println("TV and episodes removed from catalog: ", id)
		}
	}

	removeFile(fileName + ".json")
	log.Println("====================>FINISH " + mediaType)
}

func (s *Service) handleCatalogTv(listTvIdsIn []int, newIndexName string, bulkProcessor bulkAdder) {
	getter := func(ids []int) []tv.Serie { return s.tv.GetCatalogSearchIn(ids) }
	if s.getTvCatalogSearchInFn != nil {
		getter = s.getTvCatalogSearchInFn
	}

	docs := getter(listTvIdsIn)
	for _, item := range buildCatalogTvLocalized(docs) {
		req := elastic.NewBulkIndexRequest().Index(newIndexName).Doc(item)
		bulkProcessor.Add(req)
	}
}

func (s *Service) handleCatalogMovie(listMovieIdsIn []int, newIndexName string, bulkProcessor bulkAdder) {
	getter := func(ids []int) []movie.Movie { return s.movie.GetCatalogSearchIn(ids) }
	if s.getMovieCatalogSearchInFn != nil {
		getter = s.getMovieCatalogSearchInFn
	}

	docs := getter(listMovieIdsIn)
	for _, item := range buildCatalogMovieLocalized(docs) {
		req := elastic.NewBulkIndexRequest().Index(newIndexName).Doc(item)
		bulkProcessor.Add(req)
	}
}

func (s *Service) handleCatalogPerson(listPersonIdsIn []int, newIndexName string, bulkProcessor bulkAdder) {
	getter := func(language string, ids []int) []person.Person { return s.person.GetCatalogSearchIn(language, ids) }
	if s.getPersonCatalogSearchInFn != nil {
		getter = s.getPersonCatalogSearchInFn
	}

	docs := getter("en", listPersonIdsIn)
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

		var location Location
		location.Language = item.Language
		location.Title = item.Title
		location.PosterPath = item.PosterPath

		loc := catalogTvLocalizated[item.Id]
		loc.Locations = append(loc.Locations, location)
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

		var location Location
		location.Language = item.Language
		location.Title = item.Title
		location.PosterPath = item.PosterPath

		loc := catalogMovieLocalizated[item.Id]
		loc.Locations = append(loc.Locations, location)
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
	for _, item := range docs {
		catalogList = append(catalogList, CatalogSearch{
			Id:          item.Id,
			Name:        item.Name,
			CatalogType: common.MEDIA_TYPE_PERSON,
			ProfilePath: item.ProfilePath,
			Popularity:  item.Popularity,
		})
	}

	return catalogList
}

func (s *Service) CatalogSearchCharge() {
	if s.catalogSearchChargeExecutorFn != nil {
		s.catalogSearchChargeExecutorFn()
		return
	}

	workers := 5
	indexName := "catalog_search"
	elasticAliasName := indexName
	now := s.nowFn
	if now == nil {
		now = time.Now
	}
	currentTime := now()
	newIndexName := elasticAliasName + "_" + currentTime.Format("20060102150401")
	log.Println(newIndexName)

	run := s.catalogSearchChargeRunFn
	if run == nil {
		run = func(workers int, indexName string, newIndexName string) error {
			elasticClient := s.elastic.Client()
			ctx := context.Background()

			return s.executeCatalogSearchCharge(
				workers,
				indexName,
				newIndexName,
				func(name string) error {
					_, err := elasticClient.CreateIndex(name).BodyString(INDEX_MAPPING_CATALOG_SEARCH).Do(ctx)
					return err
				},
				func(workerCount int) (bulkAdder, func(), error) {
					bulkProcessor, err := elastic.NewBulkProcessorService(elasticClient).Workers(workerCount).BulkActions(-1).After(after).Stats(true).Do(ctx)
					if err != nil {
						return nil, nil, err
					}
					return bulkProcessor, func() {
						bulkProcessor.Flush()
						bulkProcessor.Close()
					}, nil
				},
				func(collection string) []int { return s.mongo.GetAllIdsByLanguage(collection, "en") },
				func(ids []int, idx string, bulk bulkAdder) { s.handleCatalogTv(ids, idx, bulk) },
				func(ids []int, idx string, bulk bulkAdder) { s.handleCatalogMovie(ids, idx, bulk) },
				func(ids []int, idx string, bulk bulkAdder) { s.handleCatalogPerson(ids, idx, bulk) },
				func(alias string) ([]string, error) { return s.IndexNamesByAlias(alias, elasticClient) },
				func(index, alias string) { elasticClient.Alias().Add(index, alias).Do(ctx) },
				func(index, alias string) { elasticClient.Alias().Remove(index, alias).Do(ctx) },
				func(index string) { elasticClient.DeleteIndex(index).Do(ctx) },
				func(index string) { elasticClient.Count(index).Do(ctx) },
			)
		}
	}

	err := run(workers, indexName, newIndexName)
	if err != nil {
		log.Println("Falha ao criar o índice:", newIndexName)
		panic(err)
	}

	log.Println("Carga finalizada com sucesso!")
}

func (s *Service) executeCatalogSearchCharge(
	workers int,
	indexName string,
	newIndexName string,
	createIndex func(name string) error,
	newBulk func(workerCount int) (bulkAdder, func(), error),
	getIDs func(collection string) []int,
	handleTv func(ids []int, newIndexName string, bulk bulkAdder),
	handleMovie func(ids []int, newIndexName string, bulk bulkAdder),
	handlePerson func(ids []int, newIndexName string, bulk bulkAdder),
	indexNamesByAlias func(alias string) ([]string, error),
	addAlias func(index, alias string),
	removeAlias func(index, alias string),
	deleteIndex func(index string),
	countIndex func(index string),
) error {
	if err := createIndex(newIndexName); err != nil {
		return err
	}

	bulkProcessor, closeBulk, err := newBulk(workers)
	if err != nil {
		log.Println("bulkProcessor Error", err)
	} else {
		idsTv := getIDs(services.CollectionSerie)
		log.Println(len(idsTv))
		processIDsInBatches(idsTv, 1000, func(batch []int) {
			handleTv(batch, newIndexName, bulkProcessor)
		})
		closeBulk()
	}

	bulkProcessor, closeBulk, err = newBulk(workers)
	if err != nil {
		log.Println("bulkProcessor Error", err)
	} else {
		idsMovies := getIDs(services.CollectionMovie)
		log.Println(len(idsMovies))
		processIDsInBatches(idsMovies, 1000, func(batch []int) {
			handleMovie(batch, newIndexName, bulkProcessor)
		})
		closeBulk()
	}

	bulkProcessor, closeBulk, err = newBulk(workers)
	if err != nil {
		log.Println("bulkProcessor Error", err)
	} else {
		idsPersons := getIDs(services.CollectionPerson)
		log.Println(len(idsPersons))
		processIDsInBatches(idsPersons, 1000, func(batch []int) {
			handlePerson(batch, newIndexName, bulkProcessor)
		})
		closeBulk()
	}

	existentSerieAliases, err := indexNamesByAlias(indexName)
	if err != nil {
		log.Println("Error ao buscar o index no alias: " + indexName)
	}
	log.Println(existentSerieAliases)

	rotateAliasAndCleanup(newIndexName, indexName, indexName, existentSerieAliases, addAlias, removeAlias, deleteIndex, countIndex)
	return nil
}

func (s *Service) handleElasticChargeInsertDocs(indexName string, listIdsIn []int, newIndexName string, bulkProcessor bulkAdder) {
	var docs []interface{}

	switch indexName {
	case "series":
		if s.getTvAllByIdsFn != nil {
			docs = s.getTvAllByIdsFn(listIdsIn)
		} else {
			docs = s.tv.GetAllByIds(listIdsIn)
		}
	case "movies":
		if s.getMovieAllByIdsFn != nil {
			docs = s.getMovieAllByIdsFn(listIdsIn)
		} else {
			docs = s.movie.GetAllByIds(listIdsIn)
		}
	case "persons":
		if s.getPersonAllByIdsFn != nil {
			docs = s.getPersonAllByIdsFn(listIdsIn)
		} else {
			docs = s.person.GetAllByIds(listIdsIn)
		}
	}

	for _, doc := range docs {
		req := elastic.NewBulkIndexRequest().Index(newIndexName).Doc(doc)
		bulkProcessor.Add(req)
	}
}

func (s *Service) ElasticChargeInsert(indexName string, interval int64, mapping string, workers int) {
	if s.elasticChargeInsertExecutorFn != nil {
		s.elasticChargeInsertExecutorFn(indexName, interval, mapping, workers)
		return
	}

	collectionCount := collectionByIndexName(indexName)
	getAllIdsByLanguage := s.getAllIdsByLanguageFn
	if getAllIdsByLanguage == nil {
		getAllIdsByLanguage = func(collection string, language string) []int {
			return s.mongo.GetAllIdsByLanguage(collection, language)
		}
	}

	docsIds := getAllIdsByLanguage(collectionCount, "en")
	elasticAliasName := indexName
	now := s.nowFn
	if now == nil {
		now = time.Now
	}
	currentTime := now()
	newIndexName := elasticAliasName + "_" + currentTime.Format("20060102150401")
	log.Println(newIndexName)

	run := s.elasticChargeInsertRunFn
	if run == nil {
		run = func(indexName string, interval int64, workers int, mapping string, newIndexName string, docsIDs []int) error {
			elasticClient := s.elastic.Client()
			ctx := context.Background()

			return s.executeElasticChargeInsert(
				indexName,
				interval,
				workers,
				newIndexName,
				func(name string) error {
					_, err := elasticClient.CreateIndex(name).BodyString(mapping).Do(context.TODO())
					return err
				},
				func(workerCount int) (bulkAdder, func(), error) {
					bulkProcessor, err := elastic.NewBulkProcessorService(elasticClient).Workers(workerCount).BulkActions(-1).After(after).Stats(true).Do(ctx)
					if err != nil {
						return nil, nil, err
					}
					return bulkProcessor, func() {
						bulkProcessor.Flush()
						bulkProcessor.Close()
					}, nil
				},
				docsIDs,
				func(ids []int, idx string, bulk bulkAdder) {
					s.handleElasticChargeInsertDocs(indexName, ids, idx, bulk)
				},
				func(alias string) ([]string, error) { return s.IndexNamesByAlias(alias, elasticClient) },
				func(index, alias string) { elasticClient.Alias().Add(index, alias).Do(ctx) },
				func(index, alias string) { elasticClient.Alias().Remove(index, alias).Do(ctx) },
				func(index string) { elasticClient.DeleteIndex(index).Do(ctx) },
				func(index string) { elasticClient.Count(index).Do(ctx) },
			)
		}
	}

	err := run(indexName, interval, workers, mapping, newIndexName, docsIds)
	if err != nil {
		log.Println("Falha ao criar o índice:", newIndexName)
		panic(err)
	}
	log.Println("Carga finalizada com sucesso!")
}

func (s *Service) executeElasticChargeInsert(
	indexName string,
	interval int64,
	workers int,
	newIndexName string,
	createIndex func(name string) error,
	newBulk func(workerCount int) (bulkAdder, func(), error),
	docsIDs []int,
	handleDocs func(ids []int, newIndexName string, bulk bulkAdder),
	indexNamesByAlias func(alias string) ([]string, error),
	addAlias func(index, alias string),
	removeAlias func(index, alias string),
	deleteIndex func(index string),
	countIndex func(index string),
) error {
	if err := createIndex(newIndexName); err != nil {
		return err
	}

	bulkProcessor, closeBulk, err := newBulk(workers)
	if err != nil {
		log.Println("bulkProcessor Error", err)
	} else {
		processIDsInBatchesWithInterval(docsIDs, interval, func(batch []int) {
			handleDocs(batch, newIndexName, bulkProcessor)
		})
		closeBulk()
	}

	existentSerieAliases, err := indexNamesByAlias(indexName)
	if err != nil {
		log.Println("Error ao buscar o index no alias: " + indexName)
	}
	log.Println(existentSerieAliases)

	rotateAliasAndCleanup(newIndexName, indexName, indexName, existentSerieAliases, addAlias, removeAlias, deleteIndex, countIndex)
	return nil
}

func (s *Service) ElasticGeneralCharge() {
	runCatalogSearchCharge := s.CatalogSearchCharge
	if s.catalogSearchChargeFn != nil {
		runCatalogSearchCharge = s.catalogSearchChargeFn
	}

	runCatalogSearchCharge()
	log.Println("FINISH ElasticGeneralCharge")
}

func (s *Service) GeneralCatalogHandler() {
	runCatalogCharge := s.CatalogCharge
	runCatalogUpdates := s.CatalogUpdates

	if s.catalogChargeFn != nil {
		runCatalogCharge = s.catalogChargeFn
	}
	if s.catalogUpdatesFn != nil {
		runCatalogUpdates = s.catalogUpdatesFn
	}

	runCatalogCharge()
	runCatalogUpdates()
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

func shouldFlushBatch(index int, interval int64) bool {
	if interval <= 0 {
		return false
	}

	return int64(index)%interval == 0
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
	for i := 0; i < len(ids); i++ {
		batch = append(batch, ids[i])
		if shouldFlushBatch(i, interval) {
			handleBatch(batch)
			batch = []int{}
		}
	}

	if len(batch) > 0 {
		handleBatch(batch)
	}
}

func rotateAliasAndCleanup(
	newIndexName string,
	aliasName string,
	countIndexName string,
	existingAliases []string,
	addAlias func(index, alias string),
	removeAlias func(index, alias string),
	deleteIndex func(index string),
	countIndex func(index string),
) {
	addAlias(newIndexName, aliasName)

	if oldIndex, ok := oldIndexToRetire(existingAliases); ok {
		removeAlias(oldIndex, aliasName)
		deleteIndex(oldIndex)
	}

	countIndex(countIndexName)
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
