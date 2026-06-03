package person

import (
	"context"
	"encoding/json"
	"log"
	"strconv"
	"time"

	"moviedb/common"
	"moviedb/services"

	"moviedb/tmdb"

	"github.com/gosimple/slug"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Service struct {
	mongo *services.MongoService
	tmdb  *tmdb.Service
}

func NewService(mongo *services.MongoService, tmdbService *tmdb.Service) *Service {
	return &Service{mongo: mongo, tmdb: tmdbService}
}

func (s *Service) CheckPersonChanges() {
	personChanges := s.tmdb.GetChangesByDataType(tmdb.DATATYPE_PERSON, 1)

	for _, person := range personChanges {
		s.PopulatePersonByIdAndLanguage(person.Id, common.LANGUAGE_PTBR, "Y")
		go s.PopulatePersonByIdAndLanguage(person.Id, common.LANGUAGE_EN, "Y")
	}
}

func (s *Service) GetPersonDetailsOnApiDb(id int, language string) Person {
	reqPerson := s.tmdb.GetDetailsByIdLanguageAndDataType(id, language, tmdb.DATATYPE_PERSON)

	var person Person
	json.NewDecoder(reqPerson.Body).Decode(&person)

	return person
}

func (s *Service) PopulatePersonByLanguage(itemObj Person, language string, updatePerson string) {
	t := time.Now()
	itemObj.UpdatedNew = t.Format("02/01/2006 15:04:05")

	itemObj.Language = language
	itemObj.Slug = slug.Make(itemObj.Name)
	itemObj.SlugUrl = "person-" + strconv.Itoa(itemObj.Id)

	itemFind := s.GetPersonByIdAndLanguage(itemObj.Id, language)

	if itemFind.Id == 0 {
		if itemObj.Id > 0 {
			log.Println("INSERT PERSON: ", language, itemObj.Id)
			s.InsertPerson(itemObj)
		}
	} else {
		if updatePerson == "Y" {
			log.Println("UPDATE PERSON: ", language, itemObj.Id)
			s.UpdatePerson(itemObj, language)
		}
	}
}

func (s *Service) PopulatePersonByIdAndLanguage(id int, language string, updatePerson string) {
	itemObj := s.GetPersonDetailsOnApiDb(id, language)
	s.PopulatePersonByLanguage(itemObj, language, updatePerson)
}

func (s *Service) PopulatePersons(language string) {
	apiMaxPage := s.tmdb.MaxPageLoad()

	for i := 1; i < apiMaxPage+1; i++ {
		log.Println("======> PERSON PAGE: ", language, i)
		page := strconv.Itoa(i)
		response := s.tmdb.GetPopularPerson(language, page)

		var result ResultPerson
		json.NewDecoder(response.Body).Decode(&result)

		for _, item := range result.Results {
			if item.Id > 0 {
				itemObj := s.GetPersonDetailsOnApiDb(item.Id, common.LANGUAGE_PTBR)
				s.PopulatePersonByLanguage(itemObj, common.LANGUAGE_PTBR, "N")

				itemObjEn := s.GetPersonDetailsOnApiDb(item.Id, language)
				go s.PopulatePersonByLanguage(itemObjEn, language, "N")
			}
		}
	}
}

func (s *Service) GetAllByIds(ids []int) []interface{} {
	ctx2 := context.TODO()
	projection := bson.M{"_id": 0, "slug": 0, "slugUrl": 0, "languages": 0, "updated": 0, "updatedNew": 0, "also_known_as": 0, "credits.cast.credit_id": 0}
	optionsFind := options.Find().SetSort(bson.D{{Key: "id", Value: 1}, {Key: "language", Value: 1}}).SetProjection(projection)
	cur, err := s.mongo.Collection(services.CollectionPerson).Find(ctx2, bson.M{"id": bson.M{"$in": ids}}, optionsFind)
	if err != nil {
		log.Println(err)
	}

	persons := make([]interface{}, 0)
	for cur.Next(ctx2) {
		var person Person
		if err := cur.Decode(&person); err != nil {
			log.Fatal(err)
		}
		persons = append(persons, person)
	}

	return persons
}

func (s *Service) GetCatalogSearchIn(language string, ids []int) []Person {
	ctx2 := context.TODO()
	projection := bson.M{"_id": 0, "id": 1, "name": 1, "profile_path": 1, "language": 1, "popularity": 1}
	optionsFind := options.Find().SetSort(bson.D{{Key: "id", Value: 1}}).SetProjection(projection)
	cur, err := s.mongo.Collection(services.CollectionPerson).Find(ctx2, bson.M{"language": language, "id": bson.M{"$in": ids}}, optionsFind)
	if err != nil {
		log.Println(err)
	}

	persons := make([]Person, 0)
	for cur.Next(ctx2) {
		var person Person
		if err := cur.Decode(&person); err != nil {
			log.Fatal(err)
		}
		persons = append(persons, person)
	}

	return persons
}

func (s *Service) GetPersonByIdAndLanguage(id int, language string) Person {
	var item Person
	s.mongo.Collection(services.CollectionPerson).FindOne(context.TODO(), bson.M{"id": id, "language": language}).Decode(&item)

	return item
}

func (s *Service) GetPersonsWithCredits(language string) []Person {
	optionsFind := options.Find()
	cur, err := s.mongo.Collection(services.CollectionPerson).Find(context.TODO(), bson.M{"credits.cast": bson.M{"$ne": nil}, "language": language}, optionsFind)
	if err != nil {
		log.Println(err)
	}

	persons := make([]Person, 0)
	for cur.Next(context.TODO()) {
		var person Person
		if err := cur.Decode(&person); err != nil {
			log.Fatal(err)
		}

		persons = append(persons, person)
	}

	cur.Close(context.TODO())

	return persons
}

func (s *Service) GetPersonsWithoutCredits(language string) []Person {
	optionsFind := options.Find()
	cur, err := s.mongo.Collection(services.CollectionPerson).Find(context.TODO(), bson.M{"credits.cast": nil, "language": language}, optionsFind)
	if err != nil {
		log.Println(err)
	}

	persons := make([]Person, 0)
	for cur.Next(context.TODO()) {
		var person Person
		if err := cur.Decode(&person); err != nil {
			log.Fatal(err)
		}

		persons = append(persons, person)
	}

	cur.Close(context.TODO())

	return persons
}

func (s *Service) InsertPerson(itemInsert Person) interface{} {
	result, err := s.mongo.Collection(services.CollectionPerson).InsertOne(context.TODO(), itemInsert)
	if err != nil {
		log.Println("EERRORRR")
		log.Println(err)
	}

	return result.InsertedID
}

func (s *Service) UpdatePerson(person Person, language string) {
	s.mongo.Collection(services.CollectionPerson).UpdateOne(context.TODO(), bson.M{"id": person.Id, "language": language}, bson.M{"$set": person})
}

func (s *Service) GetCountAll() int64 {
	return s.mongo.GetCountAllByCollection(services.CollectionPerson)
}

func (s *Service) GeneratePersonCatalogCheck(language string) map[int]common.CatalogCheck {
	return s.mongo.GenerateCatalogCheck(services.CollectionPerson, language)
}
