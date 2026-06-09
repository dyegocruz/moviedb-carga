package tv

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"moviedb/common"
	"moviedb/person"
	"moviedb/queue"
	"moviedb/services"
	"moviedb/tmdb"
	"moviedb/util"

	"github.com/gosimple/slug"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type rabbitPublisher interface {
	Close()
	PublishJSON(queueName string, message interface{}) error
}

type Service struct {
	mongo                             *services.MongoService
	person                            *person.Service
	tmdb                              *tmdb.Service
	getSerieDetailsFn                 func(id int, language string) Serie
	populateSerieByLanguageFn         func(itemObj Serie, language string)
	rabbitFactory                     func() (rabbitPublisher, error)
	getEpisodeByIdLanguageFn          func(id int, language string) Episode
	getEpisodeBySerieSeasonLanguageFn func(showId int, seasonNumber int, language string) []Episode
	getSerieByIdLanguageFn            func(id int, language string) Serie
	insertSerieFn                     func(itemObj Serie, language string) interface{}
	updateSerieFn                     func(itemObj Serie, language string)
	deleteSerieFn                     func(id int)
	insertEpisodeFn                   func(itemObj Episode, language string) interface{}
	updateEpisodeFn                   func(itemObj Episode, language string)
	deleteSerieEpisodesFn             func(showId int)
	getCountAllFn                     func() int64
	generateTvCatalogCheckFn          func(language string) map[int]common.CatalogCheck
	generateTvEpisodesCatalogCheckFn  func(language string) map[int]common.CatalogCheck
	populatePersonFn                  func(personId int, language, update string)
}

func NewService(mongo *services.MongoService, personService *person.Service, tmdbService *tmdb.Service) *Service {
	return &Service{mongo: mongo, person: personService, tmdb: tmdbService}
}

func (s *Service) CheckTvChanges() {
	factory := s.rabbitFactory
	if factory == nil {
		factory = func() (rabbitPublisher, error) {
			return services.NewRabbitMQService(nil)
		}
	}

	rmq, err := factory()
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %s", err)
	}
	defer rmq.Close()

	tvChanges := s.tmdb.GetChangesByDataType(tmdb.DATATYPE_TV, 1)

	for _, serie := range tvChanges {
		if err := rmq.PublishJSON(queue.QueueCatalogProcess, queue.CatalogProcessMessage{Id: serie.Id, MediaType: common.MEDIA_TYPE_TV}); err != nil {
			log.Fatalf("Failed to publish a message: %s", err)
		}

		log.Println("Message published successfully!")
	}

	log.Println("CheckTvChanges CONCLUDED")
}

func (s *Service) PopulateSerieByIdAndLanguage(id int, language string) {
	getter := s.GetSerieDetailsOnTMDBApi
	if s.getSerieDetailsFn != nil {
		getter = s.getSerieDetailsFn
	}

	populator := s.PopulateSerieByLanguage
	if s.populateSerieByLanguageFn != nil {
		populator = s.populateSerieByLanguageFn
	}

	itemObj := getter(id, language)
	populator(itemObj, language)
}

func (s *Service) GetSerieDetailsOnTMDBApi(id int, language string) Serie {
	reqSerie := s.tmdb.GetDetailsByIdLanguageAndDataType(id, language, tmdb.DATATYPE_TV)

	var serie Serie
	json.NewDecoder(reqSerie.Body).Decode(&serie)

	if shouldFetchAlternativeTitles(language, serie.OriginalLanguage) {
		alternativeTitles := s.GetTvAlternativeTitlesById(id)
		serie = applyLocalizedSerieTitle(serie, language, alternativeTitles)
	}

	return serie
}

func (s *Service) GetTvAlternativeTitlesById(id int) map[string]string {
	req := s.tmdb.GetAlternativeTitlesByIdAndDataType(id, tmdb.DATATYPE_TV)

	var result common.ResultAlternativeTitle
	json.NewDecoder(req.Body).Decode(&result)

	alternativeTitles := make(map[string]string)
	for _, title := range result.Results {
		if title.Iso3166_1 == common.LANGUAGE_ISO_JP && title.Type == common.ALTERNATIVE_TITLE_TYPE_ROMAJI {
			alternativeTitles[title.Iso3166_1] = title.Title
		} else if title.Iso3166_1 != common.LANGUAGE_ISO_JP {
			alternativeTitles[title.Iso3166_1] = title.Title
		}
	}

	return alternativeTitles
}

func (s *Service) HandleTvEpisodeUpdate(episodeId int, language string) {
	getter := s.GetEpisodeByIdAndLanguage
	if s.getEpisodeByIdLanguageFn != nil {
		getter = s.getEpisodeByIdLanguageFn
	}

	findEpisode := getter(episodeId, language)
	var episode Episode

	if findEpisode.Id == 0 {
		fmt.Printf("Episode %d not found\n", episodeId)
		return
	}

	reqTvEpisode := s.tmdb.GetTvSeasonEpisode(findEpisode.ShowId, findEpisode.SeasonNumber, findEpisode.EpisodeNumber, language)
	json.NewDecoder(reqTvEpisode.Body).Decode(&episode)

	episode.Language = language
	episode.ShowId = findEpisode.ShowId

	fmt.Printf("UPDATE TV - SEASON - EPISODE: %d %d %d %d\n", episode.ShowId, episode.SeasonNumber, episode.EpisodeNumber, episode.Id)
	s.UpdateEpisode(episode, language)
}

func (s *Service) PopulateSerieByLanguage(itemObj Serie, language string) {
	itemObj = applySerieMetadata(itemObj, language, time.Now())

	getByIdLang := s.GetSerieByIdAndLanguage
	if s.getSerieByIdLanguageFn != nil {
		getByIdLang = s.getSerieByIdLanguageFn
	}
	insertFn := s.InsertSerie
	if s.insertSerieFn != nil {
		insertFn = s.insertSerieFn
	}
	updateFn := s.UpdateSerie
	if s.updateSerieFn != nil {
		updateFn = s.updateSerieFn
	}
	populatePerson := func(personId int, lang, update string) {
		s.person.PopulatePersonByIdAndLanguage(personId, lang, update)
	}
	if s.populatePersonFn != nil {
		populatePerson = s.populatePersonFn
	}

	itemFind := getByIdLang(itemObj.Id, language)

	var seasonsDetails []Season
	for _, season := range itemObj.Seasons {
		reqSeasonEpisodes := s.tmdb.GetTvSeason(itemObj.Id, season.SeasonNumber, language)

		var seasonReq Season
		json.NewDecoder(reqSeasonEpisodes.Body).Decode(&seasonReq)

		for _, episode := range seasonReq.Episodes {
			foundEpisode := s.GetEpisodeByIdAndLanguage(episode.Id, language)

			if shouldInsertEpisode(foundEpisode.Id) {
				reqTvEpisode := s.tmdb.GetTvSeasonEpisode(itemObj.Id, season.SeasonNumber, episode.EpisodeNumber, language)
				json.NewDecoder(reqTvEpisode.Body).Decode(&episode)

				episode.Language = language
				log.Println("INSERT TV - SEASON - EPISODE: ", itemObj.Id, seasonReq.SeasonNumber, episode.EpisodeNumber, episode.Id)
				s.InsertEpisode(episode, language)
			} else if shouldUpdateLatestSeasonEpisode(foundEpisode, season, itemObj.NumberOfSeasons) {
				reqTvEpisode := s.tmdb.GetTvSeasonEpisode(itemObj.Id, season.SeasonNumber, episode.EpisodeNumber, language)
				json.NewDecoder(reqTvEpisode.Body).Decode(&episode)

				episode.Language = language
				log.Println("UPDATE TV - SEASON - EPISODE: ", itemObj.Id, seasonReq.SeasonNumber, episode.EpisodeNumber, episode.Id)
				s.UpdateEpisode(episode, language)
			}
		}

		seasonReq.EpisodeCount = season.EpisodeCount
		seasonReq.Overview = season.Overview
		seasonsDetails = append(seasonsDetails, seasonReq)
	}
	itemObj.Seasons = seasonsDetails

	action := decideSerieUpsertAction(itemFind.Id, itemObj.Id)
	if action == "insert" {
		for _, cast := range itemObj.TvCredits.Cast {
			populatePerson(cast.Id, language, "Y")
		}

		for _, crew := range itemObj.TvCredits.Crew {
			populatePerson(crew.Id, language, "Y")
		}

		log.Println("===>INSERT TV: ", itemObj.Id)
		insertFn(itemObj, language)
		return
	}

	if action == "update" {
		log.Println("===>UPDATE TV: ", itemObj.Id)
		updateFn(itemObj, language)
	}
}

func (s *Service) PopulateSeries(language string, idGenre string) {
	apiMaxPage := s.tmdb.MaxPageLoad()

	for i := 1; i < apiMaxPage+1; i++ {
		log.Println("======> TV PAGE: ", language, i)
		page := strconv.Itoa(i)
		response := s.tmdb.GetDiscoverTvByLanguageGenreAndPage(language, idGenre, page)

		var result ResultSerie
		json.NewDecoder(response.Body).Decode(&result)

		for _, item := range result.Results {
			checkTvExist := s.GetSerieByIdAndLanguage(item.Id, common.LANGUAGE_PTBR)
			if shouldPopulateDiscoveredSerie(item.Id, checkTvExist.Id) {
				itemObjBr := s.GetSerieDetailsOnTMDBApi(item.Id, common.LANGUAGE_PTBR)
				s.PopulateSerieByLanguage(itemObjBr, common.LANGUAGE_PTBR)

				itemObj := s.GetSerieDetailsOnTMDBApi(item.Id, language)
				go s.PopulateSerieByLanguage(itemObj, language)
			}
		}
	}
}

func (s *Service) GetAllByIds(ids []int) []interface{} {
	ctx2 := context.Background()
	projection := bson.M{"_id": 0, "slug": 0, "slugUrl": 0, "adult": 0, "seasons.episodes": 0, "credits.cast.gender": 0, "credits.cast.popularity": 0, "credits.cast.originalname": 0, "credits.crew.originalname": 0, "credits.crew.knownfordepartment": 0, "credits.crew.popularity": 0, "credits.crew.gender": 0, "updated": 0, "updatedNew": 0, "created_by.credit_id": 0, "created_by.gender": 0}
	optionsFind := options.Find().SetSort(bson.D{{Key: "id", Value: 1}, {Key: "language", Value: 1}}).SetProjection(projection)
	cur, err := s.mongo.Collection(services.CollectionSerie).Find(ctx2, bson.M{"id": bson.M{"$in": ids}}, optionsFind)
	if err != nil {
		log.Println(err)
	}

	series := make([]interface{}, 0)
	for cur.Next(ctx2) {
		var serie Serie
		if err := cur.Decode(&serie); err != nil {
			log.Fatal(err)
		}
		series = append(series, serie)
	}

	return series
}

func (s *Service) GetCatalogSearchIn(ids []int) []Serie {
	ctx2 := context.TODO()
	projection := bson.M{"_id": 0, "id": 1, "language": 1, "original_title": 1, "original_language": 1, "title": 1, "poster_path": 1, "first_air_date": 1, "popularity": 1}
	optionsFind := options.Find().SetSort(bson.D{{Key: "id", Value: 1}}).SetProjection(projection)
	cur, err := s.mongo.Collection(services.CollectionSerie).Find(ctx2, bson.M{"id": bson.M{"$in": ids}}, optionsFind)
	if err != nil {
		log.Println(err)
	}

	series := make([]Serie, 0)
	for cur.Next(ctx2) {
		var serie Serie
		if err := cur.Decode(&serie); err != nil {
			log.Fatal(err)
		}
		series = append(series, serie)
	}

	return series
}

func (s *Service) GetSerieByIdAndLanguage(id int, language string) Serie {
	if s.getSerieByIdLanguageFn != nil {
		return s.getSerieByIdLanguageFn(id, language)
	}

	var item Serie
	s.mongo.Collection(services.CollectionSerie).FindOne(context.TODO(), bson.M{"id": id, "language": language}).Decode(&item)

	return item
}

func (s *Service) InsertSerie(itemInsert Serie, language string) interface{} {
	if s.insertSerieFn != nil {
		return s.insertSerieFn(itemInsert, language)
	}

	result, err := s.mongo.Collection(services.CollectionSerie).InsertOne(context.TODO(), itemInsert)
	if err != nil {
		log.Println("EERRORRR")
		log.Println(err)
	}

	return result.InsertedID
}

func (s *Service) UpdateSerie(serie Serie, language string) {
	if s.updateSerieFn != nil {
		s.updateSerieFn(serie, language)
		return
	}

	s.mongo.Collection(services.CollectionSerie).UpdateOne(context.TODO(), bson.M{"id": serie.Id, "language": language}, bson.M{"$set": serie})
}

func (s *Service) DeleteSerie(id int) {
	if s.deleteSerieFn != nil {
		s.deleteSerieFn(id)
		return
	}

	s.mongo.Collection(services.CollectionSerie).DeleteMany(context.TODO(), bson.M{"id": id})
}

func (s *Service) InsertEpisode(itemInsert Episode, language string) interface{} {
	if s.insertEpisodeFn != nil {
		return s.insertEpisodeFn(itemInsert, language)
	}

	result, err := s.mongo.Collection(services.CollectionSerieEpisode).InsertOne(context.TODO(), itemInsert)
	if err != nil {
		log.Println("EERRORRR")
		log.Println(err)
	}

	return result.InsertedID
}

func (s *Service) GetEpisodeByIdAndLanguage(id int, language string) Episode {
	if s.getEpisodeByIdLanguageFn != nil {
		return s.getEpisodeByIdLanguageFn(id, language)
	}

	var item Episode
	s.mongo.Collection(services.CollectionSerieEpisode).FindOne(context.TODO(), bson.M{"id": id, "language": language}).Decode(&item)

	return item
}

func (s *Service) GetEpisodeBySerieSeasonAndLanguage(showId int, seasonNumber int, language string) []Episode {
	if s.getEpisodeBySerieSeasonLanguageFn != nil {
		return s.getEpisodeBySerieSeasonLanguageFn(showId, seasonNumber, language)
	}

	cur, err := s.mongo.Collection(services.CollectionSerieEpisode).Find(context.TODO(), bson.M{"show_id": showId, "season_number": seasonNumber, "language": language})
	if err != nil {
		log.Println(err)
	}

	episodes := make([]Episode, 0)
	cur.All(context.TODO(), &episodes)
	cur.Close(context.TODO())

	return episodes
}

func (s *Service) UpdateEpisode(espisode Episode, language string) {
	if s.updateEpisodeFn != nil {
		s.updateEpisodeFn(espisode, language)
		return
	}

	t := time.Now()
	espisode.UpdatedAt = t.Format("02/01/2006 15:04:05")

	s.mongo.Collection(services.CollectionSerieEpisode).UpdateOne(context.TODO(), bson.M{"id": espisode.Id, "language": language}, bson.M{"$set": espisode})
}

func (s *Service) DeleteSerieEpisodes(showId int) {
	if s.deleteSerieEpisodesFn != nil {
		s.deleteSerieEpisodesFn(showId)
		return
	}

	s.mongo.Collection(services.CollectionSerieEpisode).DeleteMany(context.TODO(), bson.M{"show_id": showId})
}

func (s *Service) GetCountAll() int64 {
	if s.getCountAllFn != nil {
		return s.getCountAllFn()
	}

	return s.mongo.GetCountAllByCollection(services.CollectionSerie)
}

func (s *Service) GenerateTvCatalogCheck(language string) map[int]common.CatalogCheck {
	if s.generateTvCatalogCheckFn != nil {
		return s.generateTvCatalogCheckFn(language)
	}

	return s.mongo.GenerateCatalogCheck(services.CollectionSerie, language)
}

func (s *Service) GenerateTvEpisodesCatalogCheck(language string) map[int]common.CatalogCheck {
	if s.generateTvEpisodesCatalogCheckFn != nil {
		return s.generateTvEpisodesCatalogCheckFn(language)
	}

	return s.mongo.GenerateCatalogCheck(services.CollectionSerieEpisode, language)
}

func applySerieMetadata(itemObj Serie, language string, now time.Time) Serie {
	itemObj.UpdatedNew = now.Format("02/01/2006 15:04:05")
	itemObj.MediaType = "serie"
	itemObj.Language = language
	itemObj.Slug = slug.Make(itemObj.Title)
	itemObj.SlugUrl = "serie-" + strconv.Itoa(itemObj.Id)
	return itemObj
}

func applyLocalizedSerieTitle(serie Serie, language string, alternativeTitles map[string]string) Serie {
	if language == common.LANGUAGE_PTBR && serie.OriginalLanguage == common.LANGUAGE_JA {
		if alternativeTitles[common.LANGUAGE_ISO_BR] != "" {
			serie.Title = alternativeTitles[common.LANGUAGE_ISO_BR]
		} else if alternativeTitles[common.LANGUAGE_ISO_JP] != "" {
			serie.Title = alternativeTitles[common.LANGUAGE_ISO_JP]
		}
	}

	return serie
}

func shouldFetchAlternativeTitles(language string, originalLanguage string) bool {
	return language == common.LANGUAGE_PTBR && originalLanguage == common.LANGUAGE_JA
}

func shouldInsertEpisode(existingEpisodeID int) bool {
	return existingEpisodeID == 0
}

func shouldUpdateLatestSeasonEpisode(episode Episode, season Season, numberOfSeasons int) bool {
	var lastSeason = episode.SeasonNumber == numberOfSeasons

	if !lastSeason {
		limitLine := season.EpisodeCount - 30

		if episode.EpisodeNumber < limitLine {
			return false
		}
	}

	if episode.UpdatedAt == "" {
		return true
	}

	updateTime, err := util.ParseStringToTime(episode.UpdatedAt)
	if err != nil {
		fmt.Println("Failed to convert date:", err)
		return true
	}

	sevenDaysAgo := time.Now().AddDate(0, 0, -7)
	needUpdate := updateTime.Before(sevenDaysAgo)

	return needUpdate
}

func decideSerieUpsertAction(existingID int, itemID int) string {
	if existingID == 0 && itemID > 0 {
		return "insert"
	}
	if existingID != 0 {
		return "update"
	}
	return "skip"
}

func shouldPopulateDiscoveredSerie(itemID int, existingID int) bool {
	return itemID > 0 && existingID == 0
}
