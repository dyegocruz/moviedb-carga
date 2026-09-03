package movie

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"moviedb/common"
	"moviedb/person"
	"moviedb/queue"
	"moviedb/services"
	"moviedb/tmdb"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type rabbitPublisher interface {
	Close()
	PublishJSON(queueName string, message interface{}) error
}

type Service struct {
	mongo                          *services.MongoService
	person                         *person.Service
	tmdb                           *tmdb.Service
	getMovieDetailsFn              func(id int, language string) Movie
	maxPageLoadFn                  func() int
	getDiscoverMoviesFn            func(language string, idGenre string, page string) ResultMovie
	populateMovieByLanguageFn      func(itemObj Movie, language string, updateCast string)
	rabbitFactory                  func() (rabbitPublisher, error)
	getMovieByIdLanguageFn         func(id int, language string) Movie
	getAllByIdsFn                  func(ids []int) []interface{}
	getCatalogSearchInFn           func(ids []int) []Movie
	insertMovieFn                  func(itemObj Movie, language string) interface{}
	updateMovieFn                  func(itemObj Movie, language string)
	deleteMovieFn                  func(id int)
	getCountAllFn                  func() int64
	generateMovieCatalogCheckFn    func(language string) map[int]common.CatalogCheck
	generateMoviePosterPathCheckFn func(language string) map[int]common.CatalogCheck
	populatePersonFn               func(personId int, language string, updateCast string)
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

	itemObjMainLanguage := getter(id, common.LANGUAGE_PTBR)
	itemObjEN := getter(id, common.LANGUAGE_EN)
	itemObjMainLanguage = mergeMovieLocalizationAndTitles(itemObjMainLanguage, itemObjEN, itemObjMainLanguage.AlternativeTitles.Titles)

	populator(itemObjMainLanguage, language, updateCast)
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
	projection := bson.M{"_id": 0, "id": 1, "language": 1, "original_title": 1, "original_language": 1, "title": 1, "poster_path": 1, "release_date": 1, "popularity": 1, "localizations": 1}
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

func (s *Service) GeneratePosterPathCheck(language string) map[int]common.CatalogCheck {
	if s.generateMoviePosterPathCheckFn != nil {
		return s.generateMoviePosterPathCheckFn(language)
	}

	return s.mongo.GeneratePosterPathCheck(services.CollectionMovie, language)
}

func (s *Service) HandleMovieImagesPath() {
	movieIds := s.GeneratePosterPathCheck(common.LANGUAGE_PTBR)
	for id := range movieIds {
		movie := s.GetMovieByIdAndLanguage(id, common.LANGUAGE_PTBR)

		if len(movie.Localizations) > 0 {
			for i, loc := range movie.Localizations {
				if loc.PosterPath == "" {
					movie.Localizations[i].PosterPath = movie.PosterPath
				}
			}
		}

		s.UpdateMovie(movie, common.LANGUAGE_PTBR)

		log.Println("MOVIE ID: ", movie.Id, " UPDATED WITH CORRECT POSTER PATH")
	}
}

func applyMovieMetadata(itemObj Movie, language string, now time.Time) Movie {
	itemObj.UpdatedAt = now.Format("02/01/2006 15:04:05")
	itemObj.MediaType = "movie"
	itemObj.Language = language
	return itemObj
}

func mergeMovieLocalizationAndTitles(mainMovie Movie, enMovie Movie, alternativeTitles []common.AlternativeTitle) Movie {
	mainMovie.AlternativeTitlesDb = mainMovie.AlternativeTitlesDb[:0]
	for _, title := range alternativeTitles {
		mainMovie.AlternativeTitlesDb = append(mainMovie.AlternativeTitlesDb, common.AlternativeTitle{
			Iso3166_1: title.Iso3166_1,
			Title:     title.Title,
			Type:      title.Type,
		})
	}

	mainMovie.Localizations = []common.LocalizationMovieTv{
		{
			Locale:     common.LANGUAGE_PTBR,
			Title:      mainMovie.Title,
			Synopsis:   mainMovie.Overview,
			Genres:     mainMovie.Genres,
			PosterPath: mainMovie.PosterPath,
		},
		{
			Locale:     common.LANGUAGE_EN,
			Title:      enMovie.Title,
			Synopsis:   enMovie.Overview,
			Genres:     enMovie.Genres,
			PosterPath: enMovie.PosterPath,
		},
	}

	return mainMovie
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
