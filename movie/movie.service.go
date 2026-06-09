package movie

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
	"time"

	"moviedb/common"
	"moviedb/person"
	"moviedb/queue"
	"moviedb/services"
	"moviedb/tmdb"

	"github.com/gosimple/slug"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type rabbitPublisher interface {
	Close()
	PublishJSON(queueName string, message interface{}) error
}

type Service struct {
	mongo  *services.MongoService
	person *person.Service
	tmdb   *tmdb.Service
	getMovieDetailsFn         func(id int, language string) Movie
	maxPageLoadFn             func() int
	getDiscoverMoviesFn       func(language string, idGenre string, page string) ResultMovie
	populateMovieByLanguageFn func(itemObj Movie, language string, updateCast string)
	rabbitFactory             func() (rabbitPublisher, error)
	getMovieByIdLanguageFn    func(id int, language string) Movie
	getAllByIdsFn             func(ids []int) []interface{}
	getCatalogSearchInFn      func(ids []int) []Movie
	insertMovieFn             func(itemObj Movie, language string) interface{}
	updateMovieFn             func(itemObj Movie, language string)
	deleteMovieFn             func(id int)
	getCountAllFn             func() int64
	generateMovieCatalogCheckFn func(language string) map[int]common.CatalogCheck
	populatePersonFn          func(personId int, language string, updateCast string)
}

func NewService(mongo *services.MongoService, personService *person.Service, tmdbService *tmdb.Service) *Service {
	return &Service{mongo: mongo, person: personService, tmdb: tmdbService}
}

func (s *Service) CheckMoviesChanges() {
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

	movieChanges := s.tmdb.GetChangesByDataType(tmdb.DATATYPE_MOVIE, 1)

	for _, item := range movieChanges {
		if err := rmq.PublishJSON(queue.QueueCatalogProcess, queue.CatalogProcessMessage{Id: item.Id, MediaType: common.MEDIA_TYPE_MOVIE}); err != nil {
			log.Fatalf("Failed to publish a message: %s", err)
		}

		log.Println("Message published successfully!")
	}
}

func (s *Service) GetMovieDetailsOnTMDBApi(id int, language string) Movie {
	movieResponse := s.tmdb.GetDetailsByIdLanguageAndDataType(id, language, tmdb.DATATYPE_MOVIE)

	var movie Movie
	json.NewDecoder(movieResponse.Body).Decode(&movie)

	return movie
}

func (s *Service) PopulateMovieByIdAndLanguage(id int, language string, updateCast string) {
	getter := s.GetMovieDetailsOnTMDBApi
	if s.getMovieDetailsFn != nil {
		getter = s.getMovieDetailsFn
	}

	populator := s.PopulateMovieByLanguage
	if s.populateMovieByLanguageFn != nil {
		populator = s.populateMovieByLanguageFn
	}

	itemObj := getter(id, language)
	populator(itemObj, language, updateCast)
}

func (s *Service) PopulateMovieByLanguage(itemObj Movie, language string, updateCast string) {
	itemObj = applyMovieMetadata(itemObj, language, time.Now())

	getByIdLang := s.GetMovieByIdAndLanguage
	if s.getMovieByIdLanguageFn != nil {
		getByIdLang = s.getMovieByIdLanguageFn
	}
	insertFn := s.InsertMovie
	if s.insertMovieFn != nil {
		insertFn = s.insertMovieFn
	}
	updateFn := s.UpdateMovie
	if s.updateMovieFn != nil {
		updateFn = s.updateMovieFn
	}
	populatePerson := func(personId int, lang, update string) {
		s.person.PopulatePersonByIdAndLanguage(personId, lang, update)
	}
	if s.populatePersonFn != nil {
		populatePerson = s.populatePersonFn
	}

	itemFind := getByIdLang(itemObj.Id, language)
	action := decideMovieUpsertAction(itemFind.Id, itemObj.Id)

	if action == "insert" {
		for _, cast := range itemObj.MovieCredits.Cast {
			populatePerson(cast.Id, language, updateCast)
		}

		for _, crew := range itemObj.MovieCredits.Crew {
			populatePerson(crew.Id, language, updateCast)
		}

		log.Println("===>INSERT MOVIE: ", itemObj.Id)
		insertFn(itemObj, language)
	} else if action == "update" {
		log.Println("===>UPDATE MOVIE: ", itemObj.Id)
		updateFn(itemObj, language)
	}
}

func (s *Service) PopulateMovies(language string, idGenre string) {
	maxPageLoad := s.maxPageLoadFn
	if maxPageLoad == nil {
		maxPageLoad = s.tmdb.MaxPageLoad
	}
	apiMaxPage := maxPageLoad()

	getDiscover := s.getDiscoverMoviesFn
	if getDiscover == nil {
		getDiscover = func(language string, idGenre string, page string) ResultMovie {
			response := s.tmdb.GetDiscoverMoviesByLanguageGenreAndPage(language, idGenre, page)
			var result ResultMovie
			json.NewDecoder(response.Body).Decode(&result)
			return result
		}
	}

	getByIdLang := s.GetMovieByIdAndLanguage
	if s.getMovieByIdLanguageFn != nil {
		getByIdLang = s.getMovieByIdLanguageFn
	}

	getDetails := s.GetMovieDetailsOnTMDBApi
	if s.getMovieDetailsFn != nil {
		getDetails = s.getMovieDetailsFn
	}

	populateByLanguage := s.PopulateMovieByLanguage
	if s.populateMovieByLanguageFn != nil {
		populateByLanguage = s.populateMovieByLanguageFn
	}

	for i := 1; i < apiMaxPage+1; i++ {
		log.Println("======> MOVIE PAGE: ", language, i)
		page := strconv.Itoa(i)
		result := getDiscover(language, idGenre, page)
		for _, item := range result.Results {
			checkMovieExist := getByIdLang(item.Id, common.LANGUAGE_PTBR)
			if shouldPopulateDiscoveredMovie(item.Id, checkMovieExist.Id) {
				itemObjPtBr := getDetails(item.Id, common.LANGUAGE_PTBR)
				populateByLanguage(itemObjPtBr, common.LANGUAGE_PTBR, "N")

				itemObjEn := getDetails(item.Id, language)
				go populateByLanguage(itemObjEn, language, "N")
			}
		}
	}
}

func (s *Service) GetAllByIds(ids []int) []interface{} {
	if s.getAllByIdsFn != nil {
		return s.getAllByIdsFn(ids)
	}

	ctx2 := context.TODO()
	projection := bson.M{"_id": 0, "slug": 0, "slugUrl": 0, "adult": 0, "credits.cast.gender": 0, "credits.cast.popularity": 0, "credits.cast.originalname": 0, "credits.crew.originalname": 0, "credits.crew.knownfordepartment": 0, "credits.crew.gender": 0, "credits.crew.popularity": 0, "updated": 0, "updatedNew": 0}
	optionsFind := options.Find().SetSort(bson.D{{Key: "id", Value: 1}, {Key: "language", Value: 1}}).SetProjection(projection)
	cur, err := s.mongo.Collection(services.CollectionMovie).Find(ctx2, bson.M{"id": bson.M{"$in": ids}}, optionsFind)
	if err != nil {
		log.Println(err)
	}

	movies := make([]interface{}, 0)
	for cur.Next(ctx2) {
		var movie Movie
		if err := cur.Decode(&movie); err != nil {
			log.Fatal(err)
		}
		movies = append(movies, movie)
	}

	return movies
}

func (s *Service) GetCatalogSearchIn(ids []int) []Movie {
	if s.getCatalogSearchInFn != nil {
		return s.getCatalogSearchInFn(ids)
	}

	ctx2 := context.TODO()
	projection := bson.M{"_id": 0, "id": 1, "language": 1, "original_title": 1, "original_language": 1, "title": 1, "poster_path": 1, "release_date": 1, "popularity": 1}
	optionsFind := options.Find().SetSort(bson.D{{Key: "id", Value: 1}}).SetProjection(projection)
	cur, err := s.mongo.Collection(services.CollectionMovie).Find(ctx2, bson.M{"id": bson.M{"$in": ids}}, optionsFind)
	if err != nil {
		log.Println(err)
	}

	movies := make([]Movie, 0)
	for cur.Next(ctx2) {
		var movie Movie
		if err := cur.Decode(&movie); err != nil {
			log.Fatal(err)
		}
		movies = append(movies, movie)
	}

	return movies
}

func (s *Service) GetMovieByIdAndLanguage(id int, language string) Movie {
	if s.getMovieByIdLanguageFn != nil {
		return s.getMovieByIdLanguageFn(id, language)
	}

	var item Movie
	s.mongo.Collection(services.CollectionMovie).FindOne(context.TODO(), bson.M{"id": id, "language": language}).Decode(&item)

	return item
}

func (s *Service) InsertMovie(itemInsert Movie, language string) interface{} {
	if s.insertMovieFn != nil {
		return s.insertMovieFn(itemInsert, language)
	}

	result, err := s.mongo.Collection(services.CollectionMovie).InsertOne(context.TODO(), itemInsert)
	if err != nil {
		log.Println("EERRORRR")
		log.Println(err)
	}

	return result.InsertedID
}

func (s *Service) UpdateMovie(movie Movie, language string) {
	if s.updateMovieFn != nil {
		s.updateMovieFn(movie, language)
		return
	}

	s.mongo.Collection(services.CollectionMovie).UpdateOne(context.TODO(), bson.M{"id": movie.Id, "language": language}, bson.M{"$set": movie})
}

func (s *Service) DeleteMovie(id int) {
	if s.deleteMovieFn != nil {
		s.deleteMovieFn(id)
		return
	}

	s.mongo.Collection(services.CollectionMovie).DeleteMany(context.TODO(), bson.M{"id": id})
}

func (s *Service) GetCountAll() int64 {
	if s.getCountAllFn != nil {
		return s.getCountAllFn()
	}

	return s.mongo.GetCountAllByCollection(services.CollectionMovie)
}

func (s *Service) GenerateMovieCatalogCheck(language string) map[int]common.CatalogCheck {
	if s.generateMovieCatalogCheckFn != nil {
		return s.generateMovieCatalogCheckFn(language)
	}

	return s.mongo.GenerateCatalogCheck(services.CollectionMovie, language)
}

func applyMovieMetadata(itemObj Movie, language string, now time.Time) Movie {
	itemObj.UpdatedNew = now.Format("02/01/2006 15:04:05")
	itemObj.MediaType = "movie"
	itemObj.Language = language
	itemObj.Slug = slug.Make(itemObj.Title)
	itemObj.SlugUrl = "movie-" + strconv.Itoa(itemObj.Id)
	return itemObj
}

func decideMovieUpsertAction(existingID int, itemID int) string {
	if existingID == 0 && itemID > 0 {
		return "insert"
	}
	if existingID != 0 {
		return "update"
	}
	return "skip"
}

func shouldPopulateDiscoveredMovie(itemID int, existingID int) bool {
	return itemID > 0 && existingID == 0
}
