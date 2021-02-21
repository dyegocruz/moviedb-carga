package movie

import (
	"context"
	"encoding/json"
	"log"
	"moviedb/database"
	"moviedb/person"
	"moviedb/util"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gosimple/slug"
	"go.mongodb.org/mongo-driver/bson"
)

// var log = logrus.New()

func Populate(language string, idGenre string) {

	// logger := logrus.New()
	// logger.Formatter = &logrus.JSONFormatter{}
	// log.SetFormatter(&logrus.TextFormatter{
	// 	FullTimestamp: true,
	// })

	apiKey := os.Getenv("TMDB_API_KEY")
	apiHost := os.Getenv("TMDB_HOST")
	apiMaxPage := util.StringToInt(os.Getenv("TMDB_MAX_PAGE_LOAD"))
	// apiMaxPage := 2
	// mongoDatabase := os.Getenv("MONGO_DATABASE")

	// moviesInsert := make([]interface{}, 0)
	// moviesUpdate := make([]Movie, 0)
	// personsInsert := make([]interface{}, 0)
	// personsUpdate := make([]person.Person, 0)
	for i := 1; i < apiMaxPage+1; i++ {
		log.Println("PAGE: ", i)
		page := strconv.Itoa(i)

		// Busca filmes por página
		response, err := http.Get(apiHost + "/discover/movie?api_key=" + apiKey + "&language=" + language + "&sort_by=popularity.desc&include_adult=false&include_video=false&page=" + page + "&with_genres=" + idGenre)
		if err != nil {
			log.Println(err)
		}

		var result ResultMovie
		json.NewDecoder(response.Body).Decode(&result)
		// log.Println(result.Results)
		personsInsert := make([]interface{}, 0)
		personsUpdate := make([]person.Person, 0)
		moviesInsert := make([]interface{}, 0)
		moviesUpdate := make([]Movie, 0)
		for _, item := range result.Results {

			// busca detalhes do filme
			reqItem, err := http.Get("https://api.themoviedb.org/3/movie/" + strconv.Itoa(item.Id) + "?api_key=" + apiKey + "&language=" + language)
			if err != nil {
				log.Println(err)
			}

			var itemObj Movie
			json.NewDecoder(reqItem.Body).Decode(&itemObj)

			t := time.Now()
			itemObj.UpdatedNew = t.Format("02/01/2006 15:04:05")

			itemObj.MediaType = "movie"
			itemObj.Language = language
			itemObj.Slug = slug.Make(itemObj.Title)
			itemObj.SlugUrl = "movie-" + strconv.Itoa(itemObj.Id)

			// INÍCIO TRATAMENTO DAS PESSOAS DO CAST E CREW
			reqCredits, err := http.Get("https://api.themoviedb.org/3/movie/" + strconv.Itoa(item.Id) + "/credits?api_key=" + apiKey + "&language=" + language)
			if err != nil {
				log.Println(err)
			}

			json.NewDecoder(reqCredits.Body).Decode(&itemObj.MovieCredits)

			// personsInsert := make([]interface{}, 0)
			// personsUpdate := make([]interface{}, 0)

			for _, cast := range itemObj.MovieCredits.Cast {

				var personCheck person.Person
				personCheck.Id = cast.Id
				personCheck.Name = cast.Name
				// personCheck.OriginalName = cast.OriginalName
				personCheck.KnowForDepartment = cast.KnownForDepartment
				personCheck.Language = language
				personCheck.Slug = slug.Make(personCheck.Name)
				personCheck.SlugUrl = "person-" + strconv.Itoa(personCheck.Id)

				personFindUpdate := person.GetItemByIdAndLanguage2(cast.Id, "person", language, personCheck)

				if personFindUpdate.Id == 0 {
					personsInsert = append(personsInsert, personCheck)
					// person.Insert("person", language, personCheck)
				} else {
					// person.Update(personCheck, language)
					// personsUpdate = append(personsUpdate, bson.M{"$set": personCheck})
					personsUpdate = append(personsUpdate, personCheck)
				}
			}

			for _, crew := range itemObj.MovieCredits.Crew {

				var personCheck person.Person
				personCheck.Id = crew.Id
				personCheck.Name = crew.Name
				// personCheck.OriginalName = crew.OriginalName
				personCheck.KnowForDepartment = crew.KnownForDepartment
				personCheck.Language = language
				personCheck.Slug = slug.Make(personCheck.Name)
				personCheck.SlugUrl = "person-" + strconv.Itoa(personCheck.Id)

				personFindUpdate := person.GetItemByIdAndLanguage2(crew.Id, "person", language, personCheck)

				// if personFindUpdate.Id == 0 {
				// 	person.Insert("person", language, personCheck)
				// }
				if personFindUpdate.Id == 0 {
					personsInsert = append(personsInsert, personCheck)
					// person.Insert("person", language, personCheck)
				} else {
					// person.Update(personCheck, language)
					// personsUpdate = append(personsUpdate, bson.M{"$set": personCheck})
					personsUpdate = append(personsUpdate, personCheck)
				}
				crew.OriginalName = ""
			}

			// FINAL TRATAMENTO DAS PESSOAS DO CAST E CREW

			itemFind := GetItemByIdAndLanguage2(itemObj.Id, "movie", language, itemObj)

			if itemFind.Id == 0 {
				log.Println("INSERT MOVIE: ", itemObj.Id)
				// Insert("movie", language, itemObj)
				moviesInsert = append(moviesInsert, itemObj)
			} else {
				log.Println("UPDATE MOVIE: ", itemObj.Id)
				// Update(itemObj, language)
				moviesUpdate = append(moviesUpdate, itemObj)
			}

			// time.Sleep(1 * time.Second / 2)

		}

		if len(moviesInsert) > 0 {
			log.Println("INSERT ALL MOVIES")
			InsertMany(moviesInsert)
			moviesInsert = make([]interface{}, 0)
		}

		if len(moviesUpdate) > 0 {
			log.Println("UPDATE ALL MOVIES")
			UpdateMany(moviesUpdate, language)
			moviesUpdate = make([]Movie, 0)
		}

		if len(personsInsert) > 0 {
			log.Println("INSERT ALL PERSON")
			person.InsertMany(personsInsert)
			personsInsert = make([]interface{}, 0)
		}

		if len(personsUpdate) > 0 {
			log.Println("UPDATE ALL PERSON")
			person.UpdateMany(personsUpdate, language)
			personsUpdate = make([]person.Person, 0)
		}

		// time.Sleep(1 * time.Second / 2)
	}

	// if len(moviesInsert) > 0 {
	// 	log.Println("INSERT ALL MOVIES")
	// 	InsertMany(moviesInsert)
	// }

	// if len(moviesUpdate) > 0 {
	// 	log.Println("UPDATE ALL MOVIES")
	// 	UpdateMany(moviesUpdate, language)
	// }

	// if len(personsInsert) > 0 {
	// 	log.Println("INSERT ALL PERSON")
	// 	person.InsertMany(personsInsert)
	// }

	// if len(personsUpdate) > 0 {
	// 	log.Println("UPDATE ALL PERSON")
	// 	person.UpdateMany(personsUpdate, language)
	// }

}

func GetAll() []Movie {
	client, ctx, cancel := database.GetConnection()
	defer cancel()
	defer client.Disconnect(ctx)

	cur, err := client.Database(os.Getenv("MONGO_DATABASE")).Collection("movie").Find(context.TODO(), bson.M{})
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

func GetItemByIdAndLanguage(id int, collecionString string, language string, itemSearh Movie) Movie {

	client, ctx, cancel := database.GetConnection()
	defer cancel()
	defer client.Disconnect(ctx)

	var item Movie
	err := client.Database(os.Getenv("MONGO_DATABASE")).Collection("movie").FindOneAndUpdate(context.TODO(), bson.M{"id": id, "language": language}, bson.M{
		"$set": itemSearh,
	}).Decode(&item)
	if err != nil {
		log.Println(err)
	}

	return item
}

func GetItemByIdAndLanguage2(id int, collecionString string, language string, itemSearh Movie) Movie {

	client, ctx, cancel := database.GetConnection()
	defer cancel()
	defer client.Disconnect(ctx)

	var item Movie
	client.Database(os.Getenv("MONGO_DATABASE")).Collection("movie").FindOne(context.TODO(), bson.M{"id": id, "language": language}).Decode(&item)

	return item
}

func Insert(collecionString string, language string, itemInsert Movie) interface{} {

	client, ctx, cancel := database.GetConnection()
	defer cancel()
	defer client.Disconnect(ctx)

	result, err := client.Database(os.Getenv("MONGO_DATABASE")).Collection("movie").InsertOne(context.TODO(), itemInsert)
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

	result, err := client.Database(os.Getenv("MONGO_DATABASE")).Collection("movie").InsertMany(context.TODO(), movies)
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

	client.Database(os.Getenv("MONGO_DATABASE")).Collection("movie").UpdateOne(context.TODO(), bson.M{"id": movie.Id, "language": language}, bson.M{
		"$set": movie,
	})

	// for _, movie := range movies {
	// 	client.Database(os.Getenv("MONGO_DATABASE")).Collection("movie").UpdateOne(context.TODO(), bson.M{"id": movie.Id, "language": language}, bson.M{
	// 		"$set": movie,
	// 	})
	// }

	// log.Println("Movies Updated: ", len(movies))
}

func UpdateMany(movies []Movie, language string) {

	client, ctx, cancel := database.GetConnection()
	defer cancel()
	defer client.Disconnect(ctx)

	for _, movie := range movies {
		client.Database(os.Getenv("MONGO_DATABASE")).Collection("movie").UpdateOne(context.TODO(), bson.M{"id": movie.Id, "language": language}, bson.M{
			"$set": movie,
		})
	}

	log.Println("Movies Updated: ", len(movies))
}
