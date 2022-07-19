package tv

import (
	"context"
	"encoding/json"
	"log"
	"moviedb/common"
	"moviedb/database"
	"moviedb/person"
	"moviedb/tmdb"
	"os"
	"strconv"
	"time"

	"github.com/gosimple/slug"
	"go.mongodb.org/mongo-driver/bson"
)

var serieCollection = database.COLLECTION_SERIE

func CheckTvChanges() {
	tvChanges := tmdb.GetChangesByDataType(tmdb.DATATYPE_TV)

	for _, serie := range tvChanges.Results {
		PopulateSerieByIdAndLanguage(serie.Id, common.LANGUAGE_EN)
		PopulateSerieByIdAndLanguage(serie.Id, common.LANGUAGE_PTBR)
	}
}

func PopulateSerieByIdAndLanguage(id int, language string) {
	itemObj := GetSerieDetailsOnTMDBApi(id, language)
	PopulateSerieByLanguage(itemObj, language)
}

func GetSerieDetailsOnTMDBApi(id int, language string) Serie {
	reqSerie := tmdb.GetDetailsByIdLanguageAndDataType(id, language, tmdb.DATATYPE_TV)

	var serie Serie
	json.NewDecoder(reqSerie.Body).Decode(&serie)

	return serie
}

func PopulateSerieByLanguage(itemObj Serie, language string) {

	// Início tratamento para episódios de uma série
	var seasonsDetails []Season
	for _, season := range itemObj.Seasons {
		reqSeasonEpisodes := tmdb.GetTvSeason(itemObj.Id, season.SeasonNumber, language)

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
	reqCredits := tmdb.GetTvCreditsByIdAndLanguage(itemObj.Id, language)

	json.NewDecoder(reqCredits.Body).Decode(&itemObj.TvCredits)

	for _, cast := range itemObj.TvCredits.Cast {
		person.PopulatePersonByIdAndLanguage(cast.Id, language)
	}

	for _, crew := range itemObj.TvCredits.Crew {
		person.PopulatePersonByIdAndLanguage(crew.Id, language)
	}
	// FINAL TRATAMENTO DAS PESSOAS DO CAST E CREW
	itemFind := GetSerieByIdAndLanguage(itemObj.Id, language)

	if itemFind.Id == 0 {
		log.Println("===>INSERT SERIE: ", itemObj.Id)
		InsertSerie(itemObj, language)
	} else {
		log.Println("===>UPDATE SERIE: ", itemObj.Id)
		UpdateSerie(itemObj, language)
	}
}

func PopulateSeries(language string, idGenre string) {

	// for i := 1; i < apiMaxPage+1; i++ {
	for i := 1; i < 10+1; i++ {
		log.Println("======> SERIE PAGE: ", language, i)
		page := strconv.Itoa(i)
		response := tmdb.GetDiscoverTvByLanguageGenreAndPage(language, idGenre, page)

		var result ResultSerie
		json.NewDecoder(response.Body).Decode(&result)

		for _, item := range result.Results {

			itemObj := GetSerieDetailsOnTMDBApi(item.Id, language)
			PopulateSerieByLanguage(itemObj, language)
		}
	}
}

func GetCountAll() int64 {
	return database.GetCountAllByColletcion(serieCollection)
}

func GetSerieByIdAndLanguage(id int, language string) Serie {

	client, ctx, cancel := database.GetConnection()
	defer cancel()
	defer client.Disconnect(ctx)

	var item Serie
	client.Database(os.Getenv("MONGO_DATABASE")).Collection(serieCollection).FindOne(context.TODO(), bson.M{"id": id, "language": language}).Decode(&item)

	return item
}

func InsertSerie(itemInsert Serie, language string) interface{} {

	client, ctx, cancel := database.GetConnection()
	defer cancel()
	defer client.Disconnect(ctx)

	result, err := client.Database(os.Getenv("MONGO_DATABASE")).Collection(serieCollection).InsertOne(context.TODO(), itemInsert)
	if err != nil {
		log.Println("EERRORRR")
		log.Println(err)
	}

	return result.InsertedID
}

func UpdateSerie(serie Serie, language string) {

	client, ctx, cancel := database.GetConnection()
	defer cancel()
	defer client.Disconnect(ctx)

	client.Database(os.Getenv("MONGO_DATABASE")).Collection(serieCollection).UpdateOne(context.TODO(), bson.M{"id": serie.Id, "language": language}, bson.M{
		"$set": serie,
	})
}
