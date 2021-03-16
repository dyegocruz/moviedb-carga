package person

import (
	"context"
	"encoding/json"
	"log"
	"moviedb/database"
	"moviedb/util"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gosimple/slug"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func Populate(language string) {

	apiKey := os.Getenv("TMDB_API_KEY")
	apiHost := os.Getenv("TMDB_HOST")
	apiMaxPage := util.StringToInt(os.Getenv("TMDB_MAX_PAGE_LOAD"))
	// mongoDatabase := os.Getenv("MONGO_DATABASE")

	personsInsert := make([]interface{}, 0)
	personsUpdate := make([]Person, 0)
	for i := 1; i < apiMaxPage+1; i++ {
		log.Println("PAGE: ", language, i)
		page := strconv.Itoa(i)
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

			itemFind := GetItemByIdAndLanguage2(itemObj.Id, "person", language, itemObj)

			if itemFind.Id == 0 {
				log.Println("INSERT PERSON: ", itemObj.Id)
				// Insert("person", language, itemObj)
				personsInsert = append(personsInsert, itemObj)
			} else {
				log.Println("UPDATE PERSON: ", itemObj.Id)
				personsUpdate = append(personsUpdate, itemObj)
			}
		}

		if len(personsInsert) > 0 {
			log.Println("INSERT ALL PERSON")
			InsertMany(personsInsert)
			personsInsert = make([]interface{}, 0)
		}

		if len(personsUpdate) > 0 {
			log.Println("UPDATE ALL PERSON")
			UpdateMany(personsUpdate, language)
			personsUpdate = make([]Person, 0)
		}

		// time.Sleep(1 * time.Second)
	}
}

func GetAll(skip int64, limit int64) []Person {
	client, ctx, cancel := database.GetConnection()
	defer cancel()
	defer client.Disconnect(ctx)

	optionsFind := options.Find().SetLimit(limit).SetSkip(skip)
	cur, err := client.Database(os.Getenv("MONGO_DATABASE")).Collection("person").Find(context.TODO(), bson.M{}, optionsFind)
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

	cur.Close(context.TODO())

	return persons
}

func GetCountAll() int64 {
	client, ctx, cancel := database.GetConnection()
	defer cancel()
	defer client.Disconnect(ctx)

	count, err := client.Database(os.Getenv("MONGO_DATABASE")).Collection("person").CountDocuments(context.TODO(), bson.M{})
	if err != nil {
		log.Println(err)
	}

	return count
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

func GetItemByIdAndLanguage2(id int, collecionString string, language string, itemSearh Person) Person {

	client, ctx, cancel := database.GetConnection()
	defer cancel()
	defer client.Disconnect(ctx)

	var item Person
	// err := client.Database(os.Getenv("MONGO_DATABASE")).Collection("person").FindOne(context.TODO(), bson.M{"id": id, "language": language}).Decode(&item)
	// if err != nil {
	// 	log.Println(err)
	// }
	client.Database(os.Getenv("MONGO_DATABASE")).Collection("person").FindOne(context.TODO(), bson.M{"id": id, "language": language}).Decode(&item)

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

func InsertMany(persons []interface{}) interface{} {

	client, ctx, cancel := database.GetConnection()
	defer cancel()
	defer client.Disconnect(ctx)

	result, err := client.Database(os.Getenv("MONGO_DATABASE")).Collection("person").InsertMany(context.TODO(), persons)
	if err != nil {
		log.Println("EERRORRR")
		log.Println(err)
	}

	log.Println("Persons Inserted: ", len(persons))

	return result.InsertedIDs
}

func Update(person Person, language string) {

	client, ctx, cancel := database.GetConnection()
	defer cancel()
	defer client.Disconnect(ctx)

	client.Database(os.Getenv("MONGO_DATABASE")).Collection("person").UpdateOne(context.TODO(), bson.M{"id": person.Id, "language": language}, bson.M{
		"$set": person,
	})
}

func UpdateMany(persons []Person, language string) {

	client, ctx, cancel := database.GetConnection()
	defer cancel()
	defer client.Disconnect(ctx)

	for _, person := range persons {
		client.Database(os.Getenv("MONGO_DATABASE")).Collection("person").UpdateOne(context.TODO(), bson.M{"id": person.Id, "language": language}, bson.M{
			"$set": persons,
		})
	}

	log.Println("Persons Updated: ", len(persons))
}
