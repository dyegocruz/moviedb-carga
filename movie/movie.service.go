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
	"go.mongodb.org/mongo-driver/mongo/options"
)

var movieCollection = database.COLLECTION_MOVIE

func CheckMoviesChanges() {
	movieChanges := tmdb.GetChangesByDataType(tmdb.DATATYPE_MOVIE)
	for _, movie := range movieChanges.Results {

		if !movie.Adult {
			PopulateMovieByIdAndLanguage(movie.Id, common.LANGUAGE_PTBR, "Y")
			go PopulateMovieByIdAndLanguage(movie.Id, common.LANGUAGE_EN, "Y")
		}
	}
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

	// INÍCIO TRATAMENTO DAS PESSOAS DO CAST E CREW
	// reqCredits := tmdb.GetMovieCreditsByIdAndLanguage(itemObj.Id, language)
	// json.NewDecoder(reqCredits.Body).Decode(&itemObj.MovieCredits)
	// FINAL TRATAMENTO DAS PESSOAS DO CAST E CREW

	itemFind := GetMovieByIdAndLanguage(itemObj.Id, language)

	if itemFind.Id == 0 {

		for _, cast := range itemObj.MovieCredits.Cast {
			person.PopulatePersonByIdAndLanguage(cast.Id, language, updateCast)
		}

		for _, crew := range itemObj.MovieCredits.Crew {
			person.PopulatePersonByIdAndLanguage(crew.Id, language, updateCast)
		}

		log.Println("===>INSERT MOVIE: ", itemObj.Id)
		InsertMovie(itemObj, language)
	} else {
		log.Println("===>UPDATE MOVIE: ", itemObj.Id)
		UpdateMovie(itemObj, language)
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

// func GetAll(skip int64, limit int64) []Movie {
// 	client, ctx, _ := database.GetConnection()
// 	defer client.Disconnect(ctx)

// 	projection := bson.M{"_id": 0, "genre_ids": 0, "slug": 0, "slugUrl": 0, "credits.cast.gender": 0, "credits.cast.knownfordepartment": 0, "credits.cast.popularity": 0, "credits.cast.originalname": 0, "credits.crew.originalname": 0, "credits.crew.knownfordepartment": 0, "credits.crew.gender": 0, "credits.crew.popularity": 0, "credits.crew.department": 0, "updated": 0, "updatedNew": 0}
// 	optionsFind := options.Find().SetLimit(limit).SetSkip(skip).SetProjection(projection)
// 	cur, err := client.Database(os.Getenv("MONGO_DATABASE")).Collection(movieCollection).Find(context.TODO(), bson.M{"language": bson.M{"$in": []string{common.LANGUAGE_EN, common.LANGUAGE_PTBR}}}, optionsFind)
// 	if err != nil {
// 		log.Println(err)
// 	}

// 	movies := make([]Movie, 0)
// 	defer cur.Close(context.TODO())
// 	for cur.Next(context.TODO()) {
// 		var movie Movie
// 		err := cur.Decode(&movie)
// 		if err != nil {
// 			log.Fatal(err)
// 		}
// 		movies = append(movies, movie)
// 	}

// 	return movies
// }

func GetByListId(listIds []int) []Movie {
	client, ctx, _ := database.GetConnection()
	defer client.Disconnect(ctx)

	projection := bson.M{"_id": 0, "genre_ids": 0, "slug": 0, "slugUrl": 0, "credits.cast.gender": 0, "credits.cast.knownfordepartment": 0, "credits.cast.popularity": 0, "credits.cast.originalname": 0, "credits.crew.originalname": 0, "credits.crew.knownfordepartment": 0, "credits.crew.gender": 0, "credits.crew.popularity": 0, "credits.crew.department": 0, "updated": 0, "updatedNew": 0}
	optionsFind := options.Find().SetProjection(projection)
	cur, err := client.Database(os.Getenv("MONGO_DATABASE")).Collection(movieCollection).Find(context.TODO(), bson.M{"id": bson.M{"$in": listIds}}, optionsFind)
	if err != nil {
		log.Println(err)
	}

	movies := make([]Movie, 0)
	for cur.Next(context.TODO()) {
		var movie Movie
		err := cur.Decode(&movie)
		if err != nil {
			log.Fatal(err)
		}
		movies = append(movies, movie)
	}

	defer cur.Close(context.TODO())

	return movies
}

func GetAllTest(batchSize int32) []Movie {
	client, ctx, _ := database.GetConnection()
	defer client.Disconnect(ctx)

	ctx2 := context.Background()

	projection := bson.M{"_id": 0, "genre_ids": 0, "slug": 0, "slugUrl": 0, "credits.cast.gender": 0, "credits.cast.knownfordepartment": 0, "credits.cast.popularity": 0, "credits.cast.originalname": 0, "credits.crew.originalname": 0, "credits.crew.knownfordepartment": 0, "credits.crew.gender": 0, "credits.crew.popularity": 0, "credits.crew.department": 0, "updated": 0, "updatedNew": 0}
	optionsFind := options.Find().SetProjection(projection).SetBatchSize(batchSize).SetNoCursorTimeout(true)
	cur, err := client.Database(os.Getenv("MONGO_DATABASE")).Collection(movieCollection).Find(ctx2, bson.D{}, optionsFind)
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

func GetCountAll() int64 {
	return database.GetCountAllByColletcion(movieCollection)
}

func GenerateMovieCatalogCheck(language string) map[int]common.CatalogCheck {
	return database.GenerateCatalogCheck(movieCollection, language)
}
