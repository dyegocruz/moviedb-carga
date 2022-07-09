package person

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

var personCollection = database.PERSON

func GetPersonDetailsOnApiDb(id int, language string) Person {
	parametro := parametro.GetByTipo("CARGA_TMDB_CONFIG")
	apiKey := parametro.Options.TmdbApiKey
	reqPerson, err := http.Get("https://api.themoviedb.org/3/person/" + strconv.Itoa(id) + "?api_key=" + apiKey + "&language=" + language)
	if err != nil {
		log.Println(err)
	}

	var person Person
	json.NewDecoder(reqPerson.Body).Decode(&person)

	return person
}

func PopulatePersonByLanguage(itemObj Person, language string) {
	parametro := parametro.GetByTipo("CARGA_TMDB_CONFIG")
	apiKey := parametro.Options.TmdbApiKey
	t := time.Now()
	itemObj.UpdatedNew = t.Format("02/01/2006 15:04:05")

	itemObj.Language = language
	itemObj.Slug = slug.Make(itemObj.Name)
	itemObj.SlugUrl = "person-" + strconv.Itoa(itemObj.Id)

	reqCredit, err := http.Get("https://api.themoviedb.org/3/person/" + strconv.Itoa(itemObj.Id) + "/combined_credits?api_key=" + apiKey + "&language=" + language)
	if err != nil {
		log.Println(err)
	}

	json.NewDecoder(reqCredit.Body).Decode(&itemObj.Credits)

	itemFind := GetPersonByIdAndLanguage(itemObj.Id, language)

	if itemFind.Id == 0 {
		log.Println("INSERT PERSON: ", itemObj.Id)
		InsertPerson(itemObj)
	} else {
		log.Println("PERSON ALREADY INSERTED: ", itemObj.Id)
	}

	// if itemFind.Id == 0 {
	// 	log.Println("INSERT PERSON: ", itemObj.Id)
	// 	InsertPerson(itemObj)
	// } else {
	// 	log.Println("UPDATE PERSON: ", itemObj.Id)
	// 	Update(itemObj, language)
	// }
}

func PopulatePersons(language string) {

	parametro := parametro.GetByTipo("CARGA_TMDB_CONFIG")

	apiKey := parametro.Options.TmdbApiKey
	apiHost := parametro.Options.TmdbHost
	apiMaxPage := parametro.Options.TmdbMaxPageLoad

	for i := 1; i < apiMaxPage+1; i++ {
		log.Println("======> PERSON PAGE: ", language, i)
		page := strconv.Itoa(i)
		response, err := http.Get(apiHost + "/person/popular?api_key=" + apiKey + "&language=" + language + "&sort_by=popularity.desc&include_adult=false&include_video=false&page=" + page)
		if err != nil {
			log.Println(err)
		}

		var result ResultPerson
		json.NewDecoder(response.Body).Decode(&result)

		for _, item := range result.Results {

			// personLocalFind := GetPersonByIdAndLanguage(item.Id, language)

			// if personLocalFind.Id == 0 {
			itemObj := GetPersonDetailsOnApiDb(item.Id, language)
			PopulatePersonByLanguage(itemObj, language)
			// }
		}
	}
}

func GetAll(skip int64, limit int64) []Person {
	client, ctx, cancel := database.GetConnection()
	defer cancel()
	defer client.Disconnect(ctx)

	optionsFind := options.Find().SetLimit(limit).SetSkip(skip)
	cur, err := client.Database(os.Getenv("MONGO_DATABASE")).Collection(personCollection).Find(context.TODO(), bson.M{}, optionsFind)
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

	count, err := client.Database(os.Getenv("MONGO_DATABASE")).Collection(personCollection).CountDocuments(context.TODO(), bson.M{})
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
	err := client.Database(os.Getenv("MONGO_DATABASE")).Collection(personCollection).FindOneAndUpdate(context.TODO(), bson.M{"id": id, "language": language}, bson.M{
		"$set": itemSearh,
	}).Decode(&item)
	if err != nil {
		log.Println(err)
	}

	return item
}

func GetPersonByIdAndLanguage(id int, language string) Person {

	client, ctx, cancel := database.GetConnection()
	defer cancel()
	defer client.Disconnect(ctx)

	var item Person
	client.Database(os.Getenv("MONGO_DATABASE")).Collection(personCollection).FindOne(context.TODO(), bson.M{"id": id, "language": language}).Decode(&item)

	return item
}

func GetPersonsWithCredits(language string) []Person {
	client, ctx, cancel := database.GetConnection()
	defer cancel()
	defer client.Disconnect(ctx)

	optionsFind := options.Find()
	cur, err := client.Database(os.Getenv("MONGO_DATABASE")).Collection(personCollection).Find(context.TODO(), bson.M{"credits.cast": bson.M{"$ne": nil}, "language": language}, optionsFind)
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

func GetPersonsWithoutCredits(language string) []Person {
	client, ctx, cancel := database.GetConnection()
	defer cancel()
	defer client.Disconnect(ctx)

	optionsFind := options.Find()
	cur, err := client.Database(os.Getenv("MONGO_DATABASE")).Collection(personCollection).Find(context.TODO(), bson.M{"credits.cast": nil, "language": language}, optionsFind)
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

func InsertPerson(itemInsert Person) interface{} {

	client, ctx, cancel := database.GetConnection()
	defer cancel()
	defer client.Disconnect(ctx)

	result, err := client.Database(os.Getenv("MONGO_DATABASE")).Collection(personCollection).InsertOne(context.TODO(), itemInsert)
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

	result, err := client.Database(os.Getenv("MONGO_DATABASE")).Collection(personCollection).InsertMany(context.TODO(), persons)
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

	client.Database(os.Getenv("MONGO_DATABASE")).Collection(personCollection).UpdateOne(context.TODO(), bson.M{"id": person.Id, "language": language}, bson.M{
		"$set": person,
	})
}

func UpdateMany(persons []Person, language string) {

	client, ctx, cancel := database.GetConnection()
	defer cancel()
	defer client.Disconnect(ctx)

	for _, person := range persons {
		client.Database(os.Getenv("MONGO_DATABASE")).Collection(personCollection).UpdateOne(context.TODO(), bson.M{"id": person.Id, "language": language}, bson.M{
			"$set": persons,
		})
	}

	log.Println("Persons Updated: ", len(persons))
}
