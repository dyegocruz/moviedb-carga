package tv

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

func GetSerieDetailsOnApiDb(id int, language string) Serie {
	parametro := parametro.GetByTipo("CARGA_TMDB_CONFIG")
	apiKey := parametro.Options.TmdbApiKey
	reqSerie, err := http.Get("https://api.themoviedb.org/3/tv/" + strconv.Itoa(id) + "?api_key=" + apiKey + "&language=" + language)
	if err != nil {
		log.Println(err)
	}

	var serie Serie
	json.NewDecoder(reqSerie.Body).Decode(&serie)

	return serie
}

func PopulateSerieByLanguage(itemObj Serie, language string) {
	parametro := parametro.GetByTipo("CARGA_TMDB_CONFIG")

	apiKey := parametro.Options.TmdbApiKey
	apiHost := parametro.Options.TmdbHost
	// Início tratamento para episódios de uma série
	var seasonsDetails []Season
	for _, season := range itemObj.Seasons {
		reqSeasonEpisodes, err := http.Get(apiHost + "/tv/" + strconv.Itoa(itemObj.Id) + "/season/" + strconv.Itoa(season.SeasonNumber) + "?api_key=" + apiKey + "&language=" + language)
		if err != nil {
			log.Println(err)
		}

		var seasonReq Season

		json.NewDecoder(reqSeasonEpisodes.Body).Decode(&seasonReq)
		seasonReq.EpisodeCount = season.EpisodeCount
		seasonReq.Overview = season.Overview
		seasonsDetails = append(seasonsDetails, seasonReq)
	}
	itemObj.Seasons = seasonsDetails
	// FINAL tratamento para episódios de uma série

	t := time.Now()
	itemObj.UpdatedNew = t.Format("02/01/2006 15:04:05")

	itemObj.MediaType = "serie"
	itemObj.Language = language
	itemObj.Slug = slug.Make(itemObj.Title)
	itemObj.SlugUrl = "serie-" + strconv.Itoa(itemObj.Id)

	// INÍCIO TRATAMENTO DAS PESSOAS DO CAST E CREW
	reqCredits, err := http.Get("https://api.themoviedb.org/3/tv/" + strconv.Itoa(itemObj.Id) + "/credits?api_key=" + apiKey + "&language=" + language)
	if err != nil {
		log.Println(err)
	}

	json.NewDecoder(reqCredits.Body).Decode(&itemObj.TvCredits)

	// for _, cast := range itemObj.TvCredits.Cast {

	// 	var personCheck person.Person
	// 	personCheck.Id = cast.Id
	// 	personCheck.Name = cast.Name
	// 	personCheck.KnowForDepartment = cast.KnownForDepartment
	// 	personCheck.Language = language
	// 	personCheck.Slug = slug.Make(personCheck.Name)
	// 	personCheck.SlugUrl = "person-" + strconv.Itoa(personCheck.Id)

	// 	person.PopulatePersonByLanguage(personCheck, language)
	// 	// personFindUpdate := person.GetPersonByIdAndLanguage(cast.Id, language)

	// 	// if personFindUpdate.Id == 0 {
	// 	// 	log.Println("INSERT PERSON CAST: ", personCheck.Id)
	// 	// 	person.InsertPerson(personCheck)
	// 	// }
	// }

	// for _, crew := range itemObj.TvCredits.Crew {

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
	// 	// 	log.Println("INSERT PERSON CREW: ", personCheck.Id)
	// 	// 	person.InsertPerson(personCheck)
	// 	// }
	// 	crew.OriginalName = ""
	// }
	// // FINAL TRATAMENTO DAS PESSOAS DO CAST E CREW
	itemFind := GetSerieByIdAndLanguage(itemObj.Id, language)

	if itemFind.Id == 0 {
		log.Println("INSERT SERIE: ", itemObj.Id)
		InsertSerie(language, itemObj)
	}
}

func PopulateSeries(language string, idGenre string) {

	parametro := parametro.GetByTipo("CARGA_TMDB_CONFIG")

	apiKey := parametro.Options.TmdbApiKey
	apiHost := parametro.Options.TmdbHost
	// apiMaxPage := parametro.Options.TmdbMaxPageLoad

	// for i := 1; i < apiMaxPage+1; i++ {
	for i := 1; i < 10+1; i++ {
		log.Println("======> SERIE PAGE: ", language, i)
		page := strconv.Itoa(i)
		response, err := http.Get(apiHost + "/discover/tv?api_key=" + apiKey + "&language=" + language + "&sort_by=popularity.desc&include_adult=false&include_video=false&page=" + page + "&with_genres=" + idGenre)
		if err != nil {
			log.Println(err)
		}

		var result ResultSerie
		json.NewDecoder(response.Body).Decode(&result)

		for _, item := range result.Results {

			serieLocalFind := GetSerieByIdAndLanguage(item.Id, language)

			if serieLocalFind.Id == 0 {
				itemObj := GetSerieDetailsOnApiDb(item.Id, language)
				PopulateSerieByLanguage(itemObj, language)
			}
		}
	}
}

func GetAll(skip int64, limit int64) []Serie {
	client, ctx, cancel := database.GetConnection()
	defer cancel()
	defer client.Disconnect(ctx)

	optionsFind := options.Find().SetLimit(limit).SetSkip(skip)
	cur, err := client.Database(os.Getenv("MONGO_DATABASE")).Collection("serie").Find(context.TODO(), bson.M{}, optionsFind)
	if err != nil {
		log.Println(err)
	}

	series := make([]Serie, 0)
	for cur.Next(context.TODO()) {
		var serie Serie
		err := cur.Decode(&serie)
		if err != nil {
			log.Fatal(err)
		}

		series = append(series, serie)
	}

	cur.Close(context.TODO())

	return series
}

func GetCountAll() int64 {
	client, ctx, cancel := database.GetConnection()
	defer cancel()
	defer client.Disconnect(ctx)

	count, err := client.Database(os.Getenv("MONGO_DATABASE")).Collection("serie").CountDocuments(context.TODO(), bson.M{})
	if err != nil {
		log.Println(err)
	}

	return count
}

func GetItemByIdAndLanguage(id int, collecionString string, language string, itemSearh Serie) Serie {

	client, ctx, cancel := database.GetConnection()
	defer cancel()
	defer client.Disconnect(ctx)

	var item Serie
	err := client.Database(os.Getenv("MONGO_DATABASE")).Collection("serie").FindOneAndUpdate(context.TODO(), bson.M{"id": id, "language": language}, bson.M{
		"$set": itemSearh,
	}).Decode(&item)
	if err != nil {
		log.Println("ID " + strconv.Itoa(id) + " NÃO REGISTRADO")
		// log.Println(err)
	}

	return item
}

func GetSerieByIdAndLanguage(id int, language string) Serie {

	client, ctx, cancel := database.GetConnection()
	defer cancel()
	defer client.Disconnect(ctx)

	var item Serie
	client.Database(os.Getenv("MONGO_DATABASE")).Collection("serie").FindOne(context.TODO(), bson.M{"id": id, "language": language}).Decode(&item)

	return item
}

func InsertSerie(language string, itemInsert Serie) interface{} {

	client, ctx, cancel := database.GetConnection()
	defer cancel()
	defer client.Disconnect(ctx)

	result, err := client.Database(os.Getenv("MONGO_DATABASE")).Collection("serie").InsertOne(context.TODO(), itemInsert)
	if err != nil {
		log.Println("EERRORRR")
		log.Println(err)
	}

	return result.InsertedID
}

func InsertMany(series []interface{}) interface{} {

	client, ctx, cancel := database.GetConnection()
	defer cancel()
	defer client.Disconnect(ctx)

	result, err := client.Database(os.Getenv("MONGO_DATABASE")).Collection("serie").InsertMany(context.TODO(), series)
	if err != nil {
		log.Println("EERRORRR")
		log.Println(err)
	}

	return result.InsertedIDs
}

func UpdateMany(persons []Serie, language string) {

	client, ctx, cancel := database.GetConnection()
	defer cancel()
	defer client.Disconnect(ctx)

	for _, person := range persons {
		client.Database(os.Getenv("MONGO_DATABASE")).Collection("serie").UpdateOne(context.TODO(), bson.M{"id": person.Id, "language": language}, bson.M{
			"$set": persons,
		})
	}
}
