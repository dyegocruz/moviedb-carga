package person

import (
	"context"
	"encoding/json"
	"log"
	"moviedb/common"
	"moviedb/database"
	"moviedb/parametro"
	"moviedb/tmdb"
	"os"
	"strconv"
	"time"

	"github.com/gosimple/slug"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var personCollection = database.COLLECTION_PERSON

func CheckPersonChanges() {
	personChanges := tmdb.GetChangesByDataType(tmdb.DATATYPE_PERSON)

	for _, person := range personChanges.Results {
		PopulatePersonByIdAndLanguage(person.Id, common.LANGUAGE_PTBR, "Y")
		go PopulatePersonByIdAndLanguage(person.Id, common.LANGUAGE_EN, "Y")
	}
}

func GetPersonDetailsOnApiDb(id int, language string) Person {
	reqPerson := tmdb.GetDetailsByIdLanguageAndDataType(id, language, tmdb.DATATYPE_PERSON)

	var person Person
	json.NewDecoder(reqPerson.Body).Decode(&person)

	return person
}

func PopulatePersonByLanguage(itemObj Person, language string, updatePerson string) {
	t := time.Now()
	itemObj.UpdatedNew = t.Format("02/01/2006 15:04:05")

	itemObj.Language = language
	itemObj.Slug = slug.Make(itemObj.Name)
	itemObj.SlugUrl = "person-" + strconv.Itoa(itemObj.Id)

	// reqCredit := tmdb.GetPersonCreditsByIdAndLanguage(itemObj.Id, language)

	// json.NewDecoder(reqCredit.Body).Decode(&itemObj.Credits)

	itemFind := GetPersonByIdAndLanguage(itemObj.Id, language)

	if itemFind.Id == 0 {
		log.Println("INSERT PERSON: ", language, itemObj.Id)
		InsertPerson(itemObj)
	} else {
		if updatePerson == "Y" {
			log.Println("UPDATE PERSON: ", language, itemObj.Id)
			UpdatePerson(itemObj, language)
		}
	}
}

func PopulatePersonByIdAndLanguage(id int, language string, updatePerson string) {
	itemObj := GetPersonDetailsOnApiDb(id, language)
	PopulatePersonByLanguage(itemObj, language, updatePerson)
}

func PopulatePersons(language string) {

	parametro := parametro.GetByTipo("CARGA_TMDB_CONFIG")
	apiMaxPage := parametro.Options.TmdbMaxPageLoad

	for i := 1; i < apiMaxPage+1; i++ {
		log.Println("======> PERSON PAGE: ", language, i)
		page := strconv.Itoa(i)
		response := tmdb.GetPopularPerson(language, page)

		var result ResultPerson
		json.NewDecoder(response.Body).Decode(&result)

		for _, item := range result.Results {

			if item.Id > 0 {
				itemObj := GetPersonDetailsOnApiDb(item.Id, common.LANGUAGE_PTBR)
				PopulatePersonByLanguage(itemObj, common.LANGUAGE_PTBR, "N")

				itemObjEn := GetPersonDetailsOnApiDb(item.Id, language)
				go PopulatePersonByLanguage(itemObjEn, language, "N")
			}
		}
	}
}

func GetAll(skip int64, limit int64) []Person {
	client, ctx, cancel := database.GetConnection()
	defer cancel()
	defer client.Disconnect(ctx)

	projection := bson.M{"_id": 0, "slug": 0, "slugUrl": 0, "popularity": 0, "languages": 0, "updated": 0, "updatedNew": 0, "also_known_as": 0, "credits.cast.credit_id": 0, "credits.crew.department": 0}
	optionsFind := options.Find().SetLimit(limit).SetSkip(skip).SetProjection(projection)
	cur, err := client.Database(os.Getenv("MONGO_DATABASE")).Collection(personCollection).Find(context.TODO(), bson.M{"language": bson.M{"$in": []string{common.LANGUAGE_EN, common.LANGUAGE_PTBR}}}, optionsFind)
	if err != nil {
		log.Println(err)
	}

	persons := make([]Person, 0)
	defer cur.Close(context.TODO())
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

func GetByListId(listIds []int) []Person {
	client, ctx, _ := database.GetConnection()
	defer client.Disconnect(ctx)

	projection := bson.M{"_id": 0, "slug": 0, "slugUrl": 0, "popularity": 0, "languages": 0, "updated": 0, "updatedNew": 0, "also_known_as": 0, "credits.cast.credit_id": 0, "credits.crew.department": 0}
	optionsFind := options.Find().SetProjection(projection)
	cur, err := client.Database(os.Getenv("MONGO_DATABASE")).Collection(personCollection).Find(context.TODO(), bson.M{"id": bson.M{"$in": listIds}}, optionsFind)
	if err != nil {
		log.Println(err)
	}

	persons := make([]Person, 0)
	defer cur.Close(context.TODO())
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

func GetAllTest(batchSize int32) []Person {
	client, ctx, cancel := database.GetConnection()
	defer cancel()
	defer client.Disconnect(ctx)
	projection := bson.M{"_id": 0, "slug": 0, "slugUrl": 0, "popularity": 0, "languages": 0, "updated": 0, "updatedNew": 0, "also_known_as": 0, "credits.cast.credit_id": 0, "credits.crew.department": 0}
	optionsFind := options.Find().SetProjection(projection).SetBatchSize(batchSize).SetNoCursorTimeout(true)
	cur, err := client.Database(os.Getenv("MONGO_DATABASE")).Collection(personCollection).Find(context.TODO(), bson.D{}, optionsFind)
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

func UpdatePerson(person Person, language string) {

	client, ctx, cancel := database.GetConnection()
	defer cancel()
	defer client.Disconnect(ctx)

	client.Database(os.Getenv("MONGO_DATABASE")).Collection(personCollection).UpdateOne(context.TODO(), bson.M{"id": person.Id, "language": language}, bson.M{
		"$set": person,
	})
}

func GetCountAll() int64 {
	return database.GetCountAllByColletcion(personCollection)
}

func GeneratePersonCatalogCheck(language string) map[int]common.CatalogCheck {
	return database.GenerateCatalogCheck(personCollection, language)
}
