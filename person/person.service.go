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

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Service struct {
	mongo                        *services.MongoService
	tmdb                         *tmdb.Service
	getPersonDetailsFn           func(id int, language string) Person
	maxPageLoadFn                func() int
	getPopularPersonFn           func(language string, page string) ResultPerson
	populatePersonByLanguageFn   func(itemObj Person, language string, updatePerson string)
	populatePersonByIdFn         func(id int, language string, updatePerson string)
	getAllByIdsFn                func(ids []int) []interface{}
	getCatalogSearchInFn         func(language string, ids []int) []Person
	getPersonByIdLanguageFn      func(id int, language string) Person
	getPersonsWithCreditsFn      func(language string) []Person
	getPersonsWithoutCreditsFn   func(language string) []Person
	insertPersonFn               func(itemObj Person) interface{}
	updatePersonFn               func(itemObj Person, language string)
	getCountAllFn                func() int64
	generatePersonCatalogCheckFn func(language string) map[int]common.CatalogCheck
}

func NewService(mongo *services.MongoService, tmdbService *tmdb.Service) *Service {
	return &Service{mongo: mongo, tmdb: tmdbService}
}

func (s *Service) CheckPersonChanges() {
	personChanges := s.tmdb.GetChangesByDataType(tmdb.DATATYPE_PERSON, 1)
	populateByID := s.PopulatePersonByIdAndLanguage
	if s.populatePersonByIdFn != nil {
		populateByID = s.populatePersonByIdFn
	}

	for _, person := range personChanges {
		populateByID(person.Id, common.LANGUAGE_PTBR, "Y")
		go populateByID(person.Id, common.LANGUAGE_EN, "Y")
	}
}

func (s *Service) GetPersonDetailsOnApiDb(id int, language string) Person {
	reqPerson := s.tmdb.GetDetailsByIdLanguageAndDataType(id, language, tmdb.DATATYPE_PERSON)

	var person Person
	json.NewDecoder(reqPerson.Body).Decode(&person)

	return person
}

func (s *Service) PopulatePersonByLanguage(itemObj Person, language string, updatePerson string) {
	itemObj = applyPersonMetadata(itemObj, language, time.Now())

	getByIdLang := s.GetPersonByIdAndLanguage
	if s.getPersonByIdLanguageFn != nil {
		getByIdLang = s.getPersonByIdLanguageFn
	}
	insertFn := s.InsertPerson
	if s.insertPersonFn != nil {
		insertFn = s.insertPersonFn
	}
	updateFn := s.UpdatePerson
	if s.updatePersonFn != nil {
		updateFn = s.updatePersonFn
	}

	itemFind := getByIdLang(itemObj.Id, language)
	action := decidePersonUpsertAction(itemFind.Id, itemObj.Id, updatePerson)

	if action == "insert" {
		log.Println("INSERT PERSON: ", language, itemObj.Id)
		insertFn(itemObj)
	} else if action == "update" {
		log.Println("UPDATE PERSON: ", language, itemObj.Id)
		updateFn(itemObj, language)
	}
}

func (s *Service) PopulatePersonByIdAndLanguage(id int, language string, updatePerson string) {
	getter := s.GetPersonDetailsOnApiDb
	if s.getPersonDetailsFn != nil {
		getter = s.getPersonDetailsFn
	}

	populator := s.PopulatePersonByLanguage
	if s.populatePersonByLanguageFn != nil {
		populator = s.populatePersonByLanguageFn
	}

	itemObjMain := getter(id, language)
	itemObjEN := getter(id, common.LANGUAGE_EN)

	localizationBR := common.LocalizationPerson{
		Locale:    common.LANGUAGE_PTBR,
		Name:      itemObjMain.Name,
		Biography: itemObjMain.Biography,
	}
	localizationEN := common.LocalizationPerson{
		Locale:    common.LANGUAGE_EN,
		Name:      itemObjEN.Name,
		Biography: itemObjEN.Biography,
	}
	itemObjMain.Localizations = append(itemObjMain.Localizations, localizationBR)
	itemObjMain.Localizations = append(itemObjMain.Localizations, localizationEN)

	populator(itemObjMain, language, updatePerson)
}

func (s *Service) PopulatePersons(language string) {
	maxPageLoad := s.maxPageLoadFn
	if maxPageLoad == nil {
		maxPageLoad = s.tmdb.MaxPageLoad
	}
	apiMaxPage := maxPageLoad()

	getPopular := s.getPopularPersonFn
	if getPopular == nil {
		getPopular = func(language string, page string) ResultPerson {
			response := s.tmdb.GetPopularPerson(language, page)
			var result ResultPerson
			json.NewDecoder(response.Body).Decode(&result)
			return result
		}
	}

	getDetails := s.GetPersonDetailsOnApiDb
	if s.getPersonDetailsFn != nil {
		getDetails = s.getPersonDetailsFn
	}

	populate := s.PopulatePersonByLanguage
	if s.populatePersonByLanguageFn != nil {
		populate = s.populatePersonByLanguageFn
	}

	for i := 1; i < apiMaxPage+1; i++ {
		log.Println("======> PERSON PAGE: ", language, i)
		page := strconv.Itoa(i)
		result := getPopular(language, page)

		for _, item := range result.Results {
			if shouldPopulateDiscoveredPerson(item.Id) {
				itemObj := getDetails(item.Id, common.LANGUAGE_PTBR)
				populate(itemObj, common.LANGUAGE_PTBR, "N")

				itemObjEn := getDetails(item.Id, language)
				go populate(itemObjEn, language, "N")
			}
		}
	}
}

func (s *Service) GetAllByIds(ids []int) []interface{} {
	if s.getAllByIdsFn != nil {
		return s.getAllByIdsFn(ids)
	}

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
	if s.getCatalogSearchInFn != nil {
		return s.getCatalogSearchInFn(language, ids)
	}

	ctx2 := context.TODO()
	projection := bson.M{"_id": 0, "id": 1, "name": 1, "profile_path": 1, "language": 1, "popularity": 1, "localizations": 1}
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
	if s.getPersonByIdLanguageFn != nil {
		return s.getPersonByIdLanguageFn(id, language)
	}

	var item Person
	s.mongo.Collection(services.CollectionPerson).FindOne(context.TODO(), bson.M{"id": id, "language": language}).Decode(&item)

	return item
}

func (s *Service) GetPersonsWithCredits(language string) []Person {
	if s.getPersonsWithCreditsFn != nil {
		return s.getPersonsWithCreditsFn(language)
	}

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
	if s.getPersonsWithoutCreditsFn != nil {
		return s.getPersonsWithoutCreditsFn(language)
	}

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
	if s.insertPersonFn != nil {
		return s.insertPersonFn(itemInsert)
	}

	result, err := s.mongo.Collection(services.CollectionPerson).InsertOne(context.TODO(), itemInsert)
	if err != nil {
		log.Println("EERRORRR")
		log.Println(err)
	}

	return result.InsertedID
}

func (s *Service) UpdatePerson(person Person, language string) {
	if s.updatePersonFn != nil {
		s.updatePersonFn(person, language)
		return
	}

	s.mongo.Collection(services.CollectionPerson).UpdateOne(context.TODO(), bson.M{"id": person.Id, "language": language}, bson.M{"$set": person})
}

func (s *Service) GetCountAll() int64 {
	if s.getCountAllFn != nil {
		return s.getCountAllFn()
	}

	return s.mongo.GetCountAllByCollection(services.CollectionPerson)
}

func (s *Service) GeneratePersonCatalogCheck(language string) map[int]common.CatalogCheck {
	if s.generatePersonCatalogCheckFn != nil {
		return s.generatePersonCatalogCheckFn(language)
	}

	return s.mongo.GenerateCatalogCheck(services.CollectionPerson, language)
}

func applyPersonMetadata(itemObj Person, language string, now time.Time) Person {
	itemObj.UpdatedAt = now.Format("02/01/2006 15:04:05")
	itemObj.Language = language
	// itemObj.Slug = slug.Make(itemObj.Name)
	// itemObj.SlugUrl = "person-" + strconv.Itoa(itemObj.Id)
	return itemObj
}

func decidePersonUpsertAction(existingID int, itemID int, updatePerson string) string {
	if existingID == 0 && itemID > 0 {
		return "insert"
	}
	if existingID != 0 && updatePerson == "Y" {
		return "update"
	}
	return "skip"
}

func shouldPopulateDiscoveredPerson(itemID int) bool {
	return itemID > 0
}
