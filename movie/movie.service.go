package movie

import (
	"context"
	"encoding/json"
	"log"
	"moviedb/database"
	"moviedb/person"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gosimple/slug"
	"go.mongodb.org/mongo-driver/bson"
)

func Populate(language string, idGenre string) {

	apiKey := os.Getenv("TMDB_API_KEY")
	apiHost := os.Getenv("TMDB_HOST")
	// mongoDatabase := os.Getenv("MONGO_DATABASE")

	for i := 0; i < 1; i++ {
		log.Println(i)
		page := strconv.Itoa(i + 1)
		response, err := http.Get(apiHost + "/discover/movie?api_key=" + apiKey + "&language=" + language + "&sort_by=popularity.desc&include_adult=false&include_video=false&page=" + page + "&with_genres=" + idGenre)
		if err != nil {
			log.Println(err)
		}

		var result ResultMovie
		json.NewDecoder(response.Body).Decode(&result)
		// log.Println(result.Results)
		for _, item := range result.Results {

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

			for _, cast := range itemObj.MovieCredits.Cast {

				var personCheck person.Person
				personCheck.Id = cast.Id
				personCheck.Name = cast.Name
				// personCheck.OriginalName = cast.OriginalName
				personCheck.KnowForDepartment = cast.KnownForDepartment
				personCheck.Language = language
				personCheck.Slug = slug.Make(personCheck.Name)
				personCheck.SlugUrl = "person-" + strconv.Itoa(personCheck.Id)

				personFindUpdate := person.GetItemByIdAndLanguage(cast.Id, "person", language, personCheck)

				if personFindUpdate.Id == 0 {
					person.Insert("person", language, personCheck)
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

				personFindUpdate := person.GetItemByIdAndLanguage(crew.Id, "person", language, personCheck)

				if personFindUpdate.Id == 0 {
					person.Insert("person", language, personCheck)
				}
				crew.OriginalName = ""
			}

			// FINAL TRATAMENTO DAS PESSOAS DO CAST E CREW

			itemFind := GetItemByIdAndLanguage(itemObj.Id, "movie", language, itemObj)

			if itemFind.Id == 0 {
				log.Println("INSERT MOVIE: ", itemObj.Id)
				Insert("movie", language, itemObj)
			} else {
				log.Println("UPDATE MOVIE: ", itemObj.Id)
			}

		}

		time.Sleep(1 * time.Second)
	}

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
