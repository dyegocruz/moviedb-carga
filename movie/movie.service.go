package movie

import (
	"context"
	"encoding/json"
	"log"
	"moviedb/common"
	"moviedb/database"
	"moviedb/parameter"

	"moviedb/person"
	"moviedb/tmdb"
	"strconv"
	"time"

	"github.com/gosimple/slug"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var movieCollectionString = database.COLLECTION_MOVIE
var movieCollection *mongo.Collection = database.GetCollection(database.DB, movieCollectionString)

func CheckMoviesChanges() {
	movieChanges := tmdb.GetChangesByDataType(tmdb.DATATYPE_MOVIE, 1)
	for _, movie := range movieChanges {

		if !movie.Adult {
			PopulateMovieByIdAndLanguage(movie.Id, common.LANGUAGE_PTBR, "Y")
			go PopulateMovieByIdAndLanguage(movie.Id, common.LANGUAGE_EN, "Y")
		}
	}
	log.Println("CheckMoviesChanges CONCLUDED")
}

func GetMovieDetailsOnTMDBApi(id int, language string) Movie {
	movieResponse := tmdb.GetDetailsByIdLanguageAndDataType(id, language, tmdb.DATATYPE_MOVIE)

	var movie Movie
	json.NewDecoder(movieResponse.Body).Decode(&movie)

	return movie
}

func PopulateMovieByIdAndLanguage(id int, language string, updateCast string) {
	itemObj := GetMovieDetailsOnTMDBApi(id, language)
	PopulateMovieByLanguage(itemObj, language, updateCast)
}

func PopulateMovieByLanguage(itemObj Movie, language string, updateCast string) {
	t := time.Now()
	itemObj.UpdatedNew = t.Format("02/01/2006 15:04:05")

	itemObj.MediaType = "movie"
	itemObj.Language = language
	itemObj.Slug = slug.Make(itemObj.Title)
	itemObj.SlugUrl = "movie-" + strconv.Itoa(itemObj.Id)

	itemFind := GetMovieByIdAndLanguage(itemObj.Id, language)

	if itemFind.Id == 0 {

		for _, cast := range itemObj.MovieCredits.Cast {
			person.PopulatePersonByIdAndLanguage(cast.Id, language, updateCast)
		}

		for _, crew := range itemObj.MovieCredits.Crew {
			person.PopulatePersonByIdAndLanguage(crew.Id, language, updateCast)
		}

		if itemObj.Id > 0 {
			log.Println("===>INSERT MOVIE: ", itemObj.Id)
			InsertMovie(itemObj, language)
		}
	} else {
		log.Println("===>UPDATE MOVIE: ", itemObj.Id)
		UpdateMovie(itemObj, language)
	}
}

func PopulateMovies(language string, idGenre string) {

	parametro := parameter.GetByType("CHARGE_TMDB_CONFIG")
	apiMaxPage := parametro.Options.TmdbMaxPageLoad

	for i := 1; i < apiMaxPage+1; i++ {
		log.Println("======> MOVIE PAGE: ", language, i)
		page := strconv.Itoa(i)

		// Busca filmes por página
		response := tmdb.GetDiscoverMoviesByLanguageGenreAndPage(language, idGenre, page)

		var result ResultMovie
		json.NewDecoder(response.Body).Decode(&result)
		for _, item := range result.Results {

			if item.Id > 0 {
				checkMovieExist := GetMovieByIdAndLanguage(item.Id, common.LANGUAGE_PTBR)

				if checkMovieExist.Id == 0 {
					itemObjPtBr := GetMovieDetailsOnTMDBApi(item.Id, common.LANGUAGE_PTBR)
					PopulateMovieByLanguage(itemObjPtBr, common.LANGUAGE_PTBR, "N")

					itemObjEn := GetMovieDetailsOnTMDBApi(item.Id, language)
					go PopulateMovieByLanguage(itemObjEn, language, "N")
				}
			}

		}
	}
}

func GetAll(skip int64, limit int64) []Movie {

	ctx2 := context.TODO()

	projection := bson.M{"_id": 0, "slug": 0, "slugUrl": 0, "adult": 0, "credits.cast.gender": 0, "credits.cast.knownfordepartment": 0, "credits.cast.popularity": 0, "credits.cast.originalname": 0, "credits.crew.originalname": 0, "credits.crew.knownfordepartment": 0, "credits.crew.gender": 0, "credits.crew.popularity": 0, "credits.crew.department": 0, "updated": 0, "updatedNew": 0}
	optionsFind := options.Find().SetSort(bson.D{{Key: "id", Value: 1}, {Key: "language", Value: 1}}).SetLimit(limit).SetSkip(skip).SetProjection(projection)
	cur, err := movieCollection.Find(ctx2, bson.D{}, optionsFind)
	if err != nil {
		log.Println(err)
	}

	movies := make([]Movie, 0)
	for cur.Next(ctx2) {
		var movie Movie
		err := cur.Decode(&movie)
		if err != nil {
			log.Fatal(err)
		}
		movies = append(movies, movie)
	}

	return movies
}

func GetCatalogSearch() []Movie {

	ctx2 := context.Background()

	projection := bson.M{"_id": 0, "id": 1, "language": 1, "original_title": 1, "original_language": 1, "title": 1, "poster_path": 1, "release_date": 1, "popularity": 1}
	optionsFind := options.Find().SetSort(bson.D{{Key: "id", Value: 1}, {Key: "language", Value: 1}}).SetProjection(projection)
	cur, err := movieCollection.Find(ctx2, bson.M{}, optionsFind)
	if err != nil {
		log.Println(err)
	}

	movies := make([]Movie, 0)
	for cur.Next(ctx2) {
		var movie Movie
		err := cur.Decode(&movie)
		if err != nil {
			log.Fatal(err)
		}
		movies = append(movies, movie)
	}

	return movies
}

func GetMovieByIdAndLanguage(id int, language string) Movie {

	var item Movie
	movieCollection.FindOne(context.TODO(), bson.M{"id": id, "language": language}).Decode(&item)

	return item
}

func InsertMovie(itemInsert Movie, language string) interface{} {

	result, err := movieCollection.InsertOne(context.TODO(), itemInsert)
	if err != nil {
		log.Println("EERRORRR")
		log.Println(err)
	}

	return result.InsertedID
}

func UpdateMovie(movie Movie, language string) {

	movieCollection.UpdateOne(context.TODO(), bson.M{"id": movie.Id, "language": language}, bson.M{
		"$set": movie,
	})
}

func GetCountAll() int64 {
	return database.GetCountAllByColletcion(database.COLLECTION_MOVIE)
}

func GenerateMovieCatalogCheck(language string) map[int]common.CatalogCheck {
	return database.GenerateCatalogCheck(database.COLLECTION_MOVIE, language)
}
