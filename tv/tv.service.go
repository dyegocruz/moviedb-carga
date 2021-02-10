package tv

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

func Populate(language string, idGenre string) {

	apiKey := os.Getenv("TMDB_API_KEY")
	apiHost := os.Getenv("TMDB_HOST")
	apiMaxPage := util.StringToInt(os.Getenv("TMDB_MAX_PAGE_LOAD"))
	// mongoDatabase := os.Getenv("MONGO_DATABASE")

	seriesInsert := make([]interface{}, 0)
	seriesUpdate := make([]Serie, 0)
	personsInsert := make([]interface{}, 0)
	personsUpdate := make([]person.Person, 0)
	for i := 1; i < apiMaxPage+1; i++ {
		log.Println("PAGE: ", i)
		page := strconv.Itoa(i)
		response, err := http.Get(apiHost + "/discover/tv?api_key=" + apiKey + "&language=" + language + "&sort_by=popularity.desc&include_adult=false&include_video=false&page=" + page + "&with_genres=" + idGenre)
		if err != nil {
			log.Println(err)
		}

		var result ResultSerie
		json.NewDecoder(response.Body).Decode(&result)
		// log.Println(result.Results)
		for _, item := range result.Results {

			reqItem, err := http.Get(apiHost + "/tv/" + strconv.Itoa(item.Id) + "?api_key=" + apiKey + "&language=" + language)
			if err != nil {
				log.Println(err)
			}

			var itemObj Serie
			json.NewDecoder(reqItem.Body).Decode(&itemObj)

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
				// for _, episode := range seasonReq.Episodes {
				// 	log.Println("DATA: ", episode.AirDate)
				// }
				// var episodesUpdate []Episode
				// for _, episode := range seasonReq.Episodes {
				// 	teste, err := time.Parse("2006-01-02", episode.AirDate)
				// 	if err != nil {
				// 		log.Println("ERRO DATA: ", itemObj.Id)
				// 		log.Println(err)
				// 	}
				// 	// t.Format("02/01/2006 15:04:05")
				// 	log.Println(teste.Format("2006-01-02 15:04:05"))
				// 	episode.AirDate = teste.Format("2006-01-02 15:04:05")
				// 	episodesUpdate = append(episodesUpdate, episode)
				// }
				// seasonReq.Episodes = episodesUpdate
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
			reqCredits, err := http.Get("https://api.themoviedb.org/3/tv/" + strconv.Itoa(item.Id) + "/credits?api_key=" + apiKey + "&language=" + language)
			if err != nil {
				log.Println(err)
			}

			json.NewDecoder(reqCredits.Body).Decode(&itemObj.TvCredits)

			for _, cast := range itemObj.TvCredits.Cast {

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
					// person.Insert("person", language, personCheck)
					personsInsert = append(personsInsert, personCheck)
				} else {
					personsUpdate = append(personsUpdate, personCheck)
				}
			}

			for _, crew := range itemObj.TvCredits.Crew {

				var personCheck person.Person
				personCheck.Id = crew.Id
				personCheck.Name = crew.Name
				// personCheck.OriginalName = crew.OriginalName
				personCheck.KnowForDepartment = crew.KnownForDepartment
				personCheck.Language = language
				personCheck.Slug = slug.Make(personCheck.Name)
				personCheck.SlugUrl = "person-" + strconv.Itoa(personCheck.Id)

				personFindUpdate := person.GetItemByIdAndLanguage2(crew.Id, "person", language, personCheck)

				if personFindUpdate.Id == 0 {
					// person.Insert("person", language, personCheck)
					personsInsert = append(personsInsert, personCheck)
				} else {
					personsUpdate = append(personsUpdate, personCheck)
				}
				crew.OriginalName = ""
			}
			// FINAL TRATAMENTO DAS PESSOAS DO CAST E CREW
			itemFind := GetItemByIdAndLanguage2(itemObj.Id, "serie", language, itemObj)

			if itemFind.Id == 0 {
				log.Println("ADD INSERT SERIE: ", itemObj.Id)
				// Insert("serie", language, itemObj)
				seriesInsert = append(seriesInsert, itemObj)
			} else {
				log.Println("ADD UPDATE SERIE: ", itemObj.Id)
				seriesUpdate = append(seriesUpdate, itemObj)
			}
		}

		// time.Sleep(1 * time.Second)
	}

	if len(seriesInsert) > 0 {
		log.Println("INSERT ALL SERIES")
		InsertMany(seriesInsert)
	}

	if len(seriesUpdate) > 0 {
		log.Println("UPDATE ALL SERIES")
		UpdateMany(seriesUpdate, language)
	}

	if len(personsInsert) > 0 {
		log.Println("INSERT ALL PERSON")
		person.InsertMany(personsInsert)
	}

	if len(personsUpdate) > 0 {
		log.Println("UPDATE ALL PERSON")
		person.UpdateMany(personsUpdate, language)
	}

}

func GetAll() []Serie {
	client, ctx, cancel := database.GetConnection()
	defer cancel()
	defer client.Disconnect(ctx)

	cur, err := client.Database(os.Getenv("MONGO_DATABASE")).Collection("serie").Find(context.TODO(), bson.M{})
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

	return series
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

func GetItemByIdAndLanguage2(id int, collecionString string, language string, itemSearh Serie) Serie {

	client, ctx, cancel := database.GetConnection()
	defer cancel()
	defer client.Disconnect(ctx)

	var item Serie
	client.Database(os.Getenv("MONGO_DATABASE")).Collection("serie").FindOne(context.TODO(), bson.M{"id": id, "language": language}).Decode(&item)

	return item
}

func Insert(collecionString string, language string, itemInsert Serie) interface{} {

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
