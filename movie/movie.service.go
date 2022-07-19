package movie

import (
	"context"
	"encoding/json"
	"log"
	"moviedb/common"
	"moviedb/database"
	"moviedb/parametro"
	"moviedb/person"
	"moviedb/tmdb"
	"os"
	"strconv"
	"time"

	"github.com/gosimple/slug"
	"go.mongodb.org/mongo-driver/bson"
)

var movieCollection = database.COLLECTION_MOVIE

func CheckMoviesChanges() {
	movieChanges := tmdb.GetChangesByDataType(tmdb.DATATYPE_MOVIE)

	for _, movie := range movieChanges.Results {
		PopulateMovieByIdAndLanguage(movie.Id, common.LANGUAGE_EN)
		PopulateMovieByIdAndLanguage(movie.Id, common.LANGUAGE_PTBR)
	}
}

func GetMovieDetailsOnTMDBApi(id int, language string) Movie {
	movieResponse := tmdb.GetDetailsByIdLanguageAndDataType(id, language, tmdb.DATATYPE_MOVIE)

	var movie Movie
	json.NewDecoder(movieResponse.Body).Decode(&movie)

	return movie
}

func PopulateMovieByIdAndLanguage(id int, language string) {
	itemObj := GetMovieDetailsOnTMDBApi(id, language)
	PopulateMovieByLanguage(itemObj, language)
}

func PopulateMovieByLanguage(itemObj Movie, language string) {

	t := time.Now()
	itemObj.UpdatedNew = t.Format("02/01/2006 15:04:05")

	itemObj.MediaType = "movie"
	itemObj.Language = language
	itemObj.Slug = slug.Make(itemObj.Title)
	itemObj.SlugUrl = "movie-" + strconv.Itoa(itemObj.Id)

	// INÍCIO TRATAMENTO DAS PESSOAS DO CAST E CREW
	reqCredits := tmdb.GetMovieCreditsByIdAndLanguage(itemObj.Id, language)

	json.NewDecoder(reqCredits.Body).Decode(&itemObj.MovieCredits)

	for _, cast := range itemObj.MovieCredits.Cast {
		person.PopulatePersonByIdAndLanguage(cast.Id, language)
	}

	for _, crew := range itemObj.MovieCredits.Crew {
		person.PopulatePersonByIdAndLanguage(crew.Id, language)
	}
	// FINAL TRATAMENTO DAS PESSOAS DO CAST E CREW

	itemFind := GetMovieByIdAndLanguage(itemObj.Id, language)

	if itemFind.Id == 0 {
		log.Println("===>INSERT MOVIE: ", itemObj.Id)
		InsertMovie(itemObj, language)
	} else {
		log.Println("===>UPDATE MOVIE: ", itemObj.Id)
		UpdateMovie(itemFind, language)
	}
}

func PopulateMovies(language string, idGenre string) {

	parametro := parametro.GetByTipo("CARGA_TMDB_CONFIG")
	apiMaxPage := parametro.Options.TmdbMaxPageLoad

	for i := 1; i < apiMaxPage+1; i++ {
		log.Println("======> MOVIE PAGE: ", language, i)
		page := strconv.Itoa(i)

		// Busca filmes por página
		response := tmdb.GetDiscoverMoviesByLanguageGenreAndPage(language, idGenre, page)

		var result ResultMovie
		json.NewDecoder(response.Body).Decode(&result)
		for _, item := range result.Results {

			itemObjEn := GetMovieDetailsOnTMDBApi(item.Id, language)
			PopulateMovieByLanguage(itemObjEn, language)

			itemObjPtBr := GetMovieDetailsOnTMDBApi(item.Id, common.LANGUAGE_PTBR)
			PopulateMovieByLanguage(itemObjPtBr, common.LANGUAGE_PTBR)

		}
	}
}

func GetCountAll() int64 {
	return database.GetCountAllByColletcion(movieCollection)
}

func GetMovieByIdAndLanguage(id int, language string) Movie {

	client, ctx, cancel := database.GetConnection()
	defer cancel()
	defer client.Disconnect(ctx)

	var item Movie
	client.Database(os.Getenv("MONGO_DATABASE")).Collection(movieCollection).FindOne(context.TODO(), bson.M{"id": id, "language": language}).Decode(&item)

	return item
}

func InsertMovie(itemInsert Movie, language string) interface{} {

	client, ctx, cancel := database.GetConnection()
	defer cancel()
	defer client.Disconnect(ctx)

	result, err := client.Database(os.Getenv("MONGO_DATABASE")).Collection(movieCollection).InsertOne(context.TODO(), itemInsert)
	if err != nil {
		log.Println("EERRORRR")
		log.Println(err)
	}

	return result.InsertedID
}

func UpdateMovie(movie Movie, language string) {

	client, ctx, cancel := database.GetConnection()
	defer cancel()
	defer client.Disconnect(ctx)

	client.Database(os.Getenv("MONGO_DATABASE")).Collection(movieCollection).UpdateOne(context.TODO(), bson.M{"id": movie.Id, "language": language}, bson.M{
		"$set": movie,
	})
}
