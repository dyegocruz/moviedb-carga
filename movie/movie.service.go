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

type Service struct {
	mongo  *services.MongoService
	person *person.Service
	tmdb   *tmdb.Service
}

func NewService(mongo *services.MongoService, personService *person.Service, tmdbService *tmdb.Service) *Service {
	return &Service{mongo: mongo, person: personService, tmdb: tmdbService}
}

func (s *Service) CheckMoviesChanges() {
	rmq, err := services.NewRabbitMQService(nil)
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
	itemObj := s.GetMovieDetailsOnTMDBApi(id, language)
	s.PopulateMovieByLanguage(itemObj, language, updateCast)
}

func (s *Service) PopulateMovieByLanguage(itemObj Movie, language string, updateCast string) {
	t := time.Now()
	itemObj.UpdatedNew = t.Format("02/01/2006 15:04:05")

	itemObj.MediaType = "movie"
	itemObj.Language = language
	itemObj.Slug = slug.Make(itemObj.Title)
	itemObj.SlugUrl = "movie-" + strconv.Itoa(itemObj.Id)

	itemFind := s.GetMovieByIdAndLanguage(itemObj.Id, language)

	if itemFind.Id == 0 {
		for _, cast := range itemObj.MovieCredits.Cast {
			s.person.PopulatePersonByIdAndLanguage(cast.Id, language, updateCast)
		}

		for _, crew := range itemObj.MovieCredits.Crew {
			s.person.PopulatePersonByIdAndLanguage(crew.Id, language, updateCast)
		}

		if itemObj.Id > 0 {
			log.Println("===>INSERT MOVIE: ", itemObj.Id)
			s.InsertMovie(itemObj, language)
		}
	} else {
		log.Println("===>UPDATE MOVIE: ", itemObj.Id)
		s.UpdateMovie(itemObj, language)
	}
}

func (s *Service) PopulateMovies(language string, idGenre string) {
	apiMaxPage := s.tmdb.MaxPageLoad()

	for i := 1; i < apiMaxPage+1; i++ {
		log.Println("======> MOVIE PAGE: ", language, i)
		page := strconv.Itoa(i)
		response := s.tmdb.GetDiscoverMoviesByLanguageGenreAndPage(language, idGenre, page)

		var result ResultMovie
		json.NewDecoder(response.Body).Decode(&result)
		for _, item := range result.Results {
			if item.Id > 0 {
				checkMovieExist := s.GetMovieByIdAndLanguage(item.Id, common.LANGUAGE_PTBR)
				if checkMovieExist.Id == 0 {
					itemObjPtBr := s.GetMovieDetailsOnTMDBApi(item.Id, common.LANGUAGE_PTBR)
					s.PopulateMovieByLanguage(itemObjPtBr, common.LANGUAGE_PTBR, "N")

					itemObjEn := s.GetMovieDetailsOnTMDBApi(item.Id, language)
					go s.PopulateMovieByLanguage(itemObjEn, language, "N")
				}
			}
		}
	}
}

func (s *Service) GetAllByIds(ids []int) []interface{} {
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
	var item Movie
	s.mongo.Collection(services.CollectionMovie).FindOne(context.TODO(), bson.M{"id": id, "language": language}).Decode(&item)

	return item
}

func (s *Service) InsertMovie(itemInsert Movie, language string) interface{} {
	result, err := s.mongo.Collection(services.CollectionMovie).InsertOne(context.TODO(), itemInsert)
	if err != nil {
		log.Println("EERRORRR")
		log.Println(err)
	}

	return result.InsertedID
}

func (s *Service) UpdateMovie(movie Movie, language string) {
	s.mongo.Collection(services.CollectionMovie).UpdateOne(context.TODO(), bson.M{"id": movie.Id, "language": language}, bson.M{"$set": movie})
}

func (s *Service) DeleteMovie(id int) {
	s.mongo.Collection(services.CollectionMovie).DeleteMany(context.TODO(), bson.M{"id": id})
}

func (s *Service) GetCountAll() int64 {
	return s.mongo.GetCountAllByCollection(services.CollectionMovie)
}

func (s *Service) GenerateMovieCatalogCheck(language string) map[int]common.CatalogCheck {
	return s.mongo.GenerateCatalogCheck(services.CollectionMovie, language)
}
