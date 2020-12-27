package person

import (
	"context"
	"encoding/json"
	"log"
	"moviedb/database"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gosimple/slug"
	"go.mongodb.org/mongo-driver/bson"
)

func Populate(language string) {

	apiKey := os.Getenv("TMDB_API_KEY")
	apiHost := os.Getenv("TMDB_HOST")
	// mongoDatabase := os.Getenv("MONGO_DATABASE")

	for i := 0; i < 1; i++ {
		log.Println(i)
		page := strconv.Itoa(i + 1)
		response, err := http.Get(apiHost + "/person/popular?api_key=" + apiKey + "&language=" + language + "&sort_by=popularity.desc&include_adult=false&include_video=false&page=" + page)
		if err != nil {
			log.Println(err)
		}

		var result ResultPerson
		json.NewDecoder(response.Body).Decode(&result)
		// log.Println(result.Results)
		for _, item := range result.Results {

			reqItem, err := http.Get("https://api.themoviedb.org/3/person/" + strconv.Itoa(item.Id) + "?api_key=" + apiKey + "&language=" + language)
			if err != nil {
				log.Println(err)
			}

			var itemObj Person
			json.NewDecoder(reqItem.Body).Decode(&itemObj)

			t := time.Now()
			itemObj.UpdatedNew = t.Format("02/01/2006 15:04:05")

			itemObj.Language = language
			itemObj.Slug = slug.Make(itemObj.Name)
			itemObj.SlugUrl = "person-" + strconv.Itoa(itemObj.Id)

			reqCredit, err := http.Get("https://api.themoviedb.org/3/person/" + strconv.Itoa(item.Id) + "/combined_credits?api_key=" + apiKey + "&language=" + language)
			if err != nil {
				log.Println(err)
			}

			json.NewDecoder(reqCredit.Body).Decode(&itemObj.Credits)

			itemFind := GetItemByIdAndLanguage(itemObj.Id, "person", language, itemObj)

			if itemFind.Id == 0 {
				log.Println("INSERT", itemObj.Id)
				Insert("person", language, itemObj)
			} else {
				log.Println("UPDATE", itemObj.Id)
			}
		}

		time.Sleep(1 * time.Second)
	}

}

func GetAll() []Person {
	client, ctx, cancel := database.GetConnection()
	defer cancel()
	defer client.Disconnect(ctx)

	cur, err := client.Database(os.Getenv("MONGO_DATABASE")).Collection("person").Find(context.TODO(), bson.M{})
	if err != nil {
		log.Println(err)
	}

	persons := make([]Person, 0)
	for cur.Next(context.TODO()) {
		var person Person
		err := cur.Decode(&person)
		if err != nil {
			log.Fatal(err)
		}

		persons = append(persons, person)
	}

	return persons
}

func GetItemByIdAndLanguage(id int, collecionString string, language string, itemSearh Person) Person {

	client, ctx, cancel := database.GetConnection()
	defer cancel()
	defer client.Disconnect(ctx)

	var item Person
	err := client.Database(os.Getenv("MONGO_DATABASE")).Collection("person").FindOneAndUpdate(context.TODO(), bson.M{"id": id, "language": language}, bson.M{
		"$set": itemSearh,
	}).Decode(&item)
	if err != nil {
		log.Println(err)
	}

	return item
}

func Insert(collecionString string, language string, itemInsert Person) interface{} {

	client, ctx, cancel := database.GetConnection()
	defer cancel()
	defer client.Disconnect(ctx)

	result, err := client.Database(os.Getenv("MONGO_DATABASE")).Collection("person").InsertOne(context.TODO(), itemInsert)
	if err != nil {
		log.Println("EERRORRR")
		log.Println(err)
	}

	return result.InsertedID
}
