package movie

import (
	"context"
	"encoding/json"
	"log"
	"moviedb/database"
	"moviedb/parametro"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gosimple/slug"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var movieCollection = database.MOVIE

func GetMovieDetailsOnApiDb(id int, language string) Movie {
	parametro := parametro.GetByTipo("CARGA_TMDB_CONFIG")
	apiKey := parametro.Options.TmdbApiKey
	reqMovie, err := http.Get("https://api.themoviedb.org/3/movie/" + strconv.Itoa(id) + "?api_key=" + apiKey + "&language=" + language)
	if err != nil {
		log.Println(err)
	}

	var movie Movie
	json.NewDecoder(reqMovie.Body).Decode(&movie)

	return movie
}

func PopulateMovieByLanguage(itemObj Movie, language string) {
	parametro := parametro.GetByTipo("CARGA_TMDB_CONFIG")
	apiKey := parametro.Options.TmdbApiKey
	t := time.Now()
	itemObj.UpdatedNew = t.Format("02/01/2006 15:04:05")

	itemObj.MediaType = "movie"
	itemObj.Language = language
	itemObj.Slug = slug.Make(itemObj.Title)
	itemObj.SlugUrl = "movie-" + strconv.Itoa(itemObj.Id)

	// INÍCIO TRATAMENTO DAS PESSOAS DO CAST E CREW
	reqCredits, err := http.Get("https://api.themoviedb.org/3/movie/" + strconv.Itoa(itemObj.Id) + "/credits?api_key=" + apiKey + "&language=" + language)
	if err != nil {
		log.Println(err)
	}

	json.NewDecoder(reqCredits.Body).Decode(&itemObj.MovieCredits)

	// for _, cast := range itemObj.MovieCredits.Cast {

	// 	var personCheck person.Person
	// 	personCheck.Id = cast.Id
	// 	personCheck.Name = cast.Name
	// 	// personCheck.OriginalName = cast.OriginalName
	// 	personCheck.KnowForDepartment = cast.KnownForDepartment
	// 	personCheck.Language = language
	// 	personCheck.Slug = slug.Make(personCheck.Name)
	// 	personCheck.SlugUrl = "person-" + strconv.Itoa(personCheck.Id)

	// 	// personFindUpdate := person.GetPersonByIdAndLanguage(cast.Id, language)

	// 	person.PopulatePersonByLanguage(personCheck, language)
	// 	// if personFindUpdate.Id == 0 {
	// 	// 	log.Println("TREAT PERSON CAST: ", personCheck.Id)
	// 	// 	person.PopulatePersonByLanguage(personCheck, language)
	// 	// 	// person.InsertPerson(personCheck)
	// 	// }
	// }

	// for _, crew := range itemObj.MovieCredits.Crew {

	// 	var personCheck person.Person
	// 	personCheck.Id = crew.Id
	// 	personCheck.Name = crew.Name
	// 	personCheck.KnowForDepartment = crew.KnownForDepartment
	// 	personCheck.Language = language
	// 	personCheck.Slug = slug.Make(personCheck.Name)
	// 	personCheck.SlugUrl = "person-" + strconv.Itoa(personCheck.Id)
	// 	person.PopulatePersonByLanguage(personCheck, language)
	// 	// personFindUpdate := person.GetPersonByIdAndLanguage(crew.Id, language)

	// 	// if personFindUpdate.Id == 0 {
	// 	// 	log.Println("TREAT PERSON CREW: ", personCheck.Id)
	// 	// 	person.PopulatePersonByLanguage(personCheck, language)
	// 	// 	// person.InsertPerson(personCheck)
	// 	// }
	// 	crew.OriginalName = ""
	// }
	// // FINAL TRATAMENTO DAS PESSOAS DO CAST E CREW

	itemFind := GetMovieByIdAndLanguage(itemObj.Id, language)

	if itemFind.Id == 0 {
		log.Println("INSERT MOVIE: ", itemObj.Id)
		InsertMovie(language, itemObj)
	} else {
		log.Println("MOVIE ALREADY INSERTED: ", itemObj.Id)
	}
}

func PopulateMovies(language string, idGenre string) {

	parametro := parametro.GetByTipo("CARGA_TMDB_CONFIG")

	apiKey := parametro.Options.TmdbApiKey
	apiHost := parametro.Options.TmdbHost
	apiMaxPage := parametro.Options.TmdbMaxPageLoad

	for i := 1; i < apiMaxPage+1; i++ {
		log.Println("======> MOVIE PAGE: ", language, i)
		page := strconv.Itoa(i)

		// Busca filmes por página
		response, err := http.Get(apiHost + "/discover/movie?api_key=" + apiKey + "&language=" + language + "&sort_by=popularity.desc&include_adult=false&include_video=false&page=" + page + "&with_genres=" + idGenre)
		if err != nil {
			log.Println(err)
		}

		var result ResultMovie
		json.NewDecoder(response.Body).Decode(&result)
		for _, item := range result.Results {

			movieLocalFind := GetMovieByIdAndLanguage(item.Id, language)
			if movieLocalFind.Id == 0 {
				itemObj := GetMovieDetailsOnApiDb(item.Id, language)
				PopulateMovieByLanguage(itemObj, language)
			}

		}
	}
}

func GetAll(skip int64, limit int64) []Movie {
	client, ctx, cancel := database.GetConnection()
	defer cancel()
	defer client.Disconnect(ctx)

	optionsFind := options.Find().SetLimit(limit).SetSkip(skip)
	cur, err := client.Database(os.Getenv("MONGO_DATABASE")).Collection(movieCollection).Find(context.TODO(), bson.M{}, optionsFind)
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

	cur.Close(context.TODO())

	return movies
}

func GetCountAll() int64 {
	client, ctx, cancel := database.GetConnection()
	defer cancel()
	defer client.Disconnect(ctx)

	count, err := client.Database(os.Getenv("MONGO_DATABASE")).Collection(movieCollection).CountDocuments(context.TODO(), bson.M{})
	if err != nil {
		log.Println(err)
	}

	return count
}

func GetMovieByIdAndLanguage(id int, language string) Movie {

	client, ctx, cancel := database.GetConnection()
	defer cancel()
	defer client.Disconnect(ctx)

	var item Movie
	client.Database(os.Getenv("MONGO_DATABASE")).Collection(movieCollection).FindOne(context.TODO(), bson.M{"id": id, "language": language}).Decode(&item)

	return item
}

func InsertMovie(language string, itemInsert Movie) interface{} {

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

func InsertMany(movies []interface{}) interface{} {

	client, ctx, cancel := database.GetConnection()
	defer cancel()
	defer client.Disconnect(ctx)

	result, err := client.Database(os.Getenv("MONGO_DATABASE")).Collection(movieCollection).InsertMany(context.TODO(), movies)
	if err != nil {
		log.Println("EERRORRR")
		log.Println(err)
	}

	log.Println("Movies Inserted: ", len(movies))

	return result.InsertedIDs
}

func Update(movie Movie, language string) {

	client, ctx, cancel := database.GetConnection()
	defer cancel()
	defer client.Disconnect(ctx)

	client.Database(os.Getenv("MONGO_DATABASE")).Collection(movieCollection).UpdateOne(context.TODO(), bson.M{"id": movie.Id, "language": language}, bson.M{
		"$set": movie,
	})
}

func UpdateMany(movies []Movie, language string) {

	client, ctx, cancel := database.GetConnection()
	defer cancel()
	defer client.Disconnect(ctx)

	for _, movie := range movies {
		client.Database(os.Getenv("MONGO_DATABASE")).Collection(movieCollection).UpdateOne(context.TODO(), bson.M{"id": movie.Id, "language": language}, bson.M{
			"$set": movie,
		})
	}

	log.Println("Movies Updated: ", len(movies))
}
