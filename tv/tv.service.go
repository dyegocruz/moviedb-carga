package tv

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
			var seasonsDetails []Season
			for _, season := range itemObj.Seasons {
				// https://api.themoviedb.org/3/tv/82856/season/1?api_key=26fe6f55e55736490dee0811901cccac&language=en-US
				reqSeasonEpisodes, err := http.Get(apiHost + "/tv/" + strconv.Itoa(itemObj.Id) + "/season/" + strconv.Itoa(season.SeasonNumber) + "?api_key=" + apiKey + "&language=" + language)
				if err != nil {
					log.Println(err)
				}

				var seasonReq Season
				json.NewDecoder(reqSeasonEpisodes.Body).Decode(&seasonReq)
				seasonsDetails = append(seasonsDetails, seasonReq)
			}
			itemObj.Seasons = seasonsDetails

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

				personFindUpdate := person.GetItemByIdAndLanguage(cast.Id, "person", language, personCheck)

				if personFindUpdate.Id == 0 {
					person.Insert("person", language, personCheck)
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

				personFindUpdate := person.GetItemByIdAndLanguage(crew.Id, "person", language, personCheck)

				if personFindUpdate.Id == 0 {
					person.Insert("person", language, personCheck)
				}
				crew.OriginalName = ""
			}

			// FINAL TRATAMENTO DAS PESSOAS DO CAST E CREW

			itemFind := GetItemByIdAndLanguage(itemObj.Id, "serie", language, itemObj)

			if itemFind.Id == 0 {
				log.Println("INSERT SERIE: ", itemObj.Id)
				Insert("serie", language, itemObj)
			} else {
				log.Println("UPDATE SERIE: ", itemObj.Id)
			}
		}

		time.Sleep(1 * time.Second / 2)
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
