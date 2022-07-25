package tv

import (
	"context"
	"encoding/json"
	"log"
	"moviedb/common"
	"moviedb/database"
	"moviedb/parametro"
	"moviedb/person"
	"moviedb/tmdb"
	"os"
	"strconv"
	"time"

	"github.com/gosimple/slug"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
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

	t := time.Now()
	itemObj.UpdatedNew = t.Format("02/01/2006 15:04:05")

	itemObj.MediaType = "serie"
	itemObj.Language = language
	itemObj.Slug = slug.Make(itemObj.Title)
	itemObj.SlugUrl = "serie-" + strconv.Itoa(itemObj.Id)

	// INÍCIO TRATAMENTO DAS PESSOAS DO CAST E CREW
	reqCredits := tmdb.GetTvCreditsByIdAndLanguage(itemObj.Id, language)

	json.NewDecoder(reqCredits.Body).Decode(&itemObj.TvCredits)

	// for _, cast := range itemObj.TvCredits.Cast {
	// 	person.PopulatePersonByIdAndLanguage(cast.Id, language)
	// }

	// for _, crew := range itemObj.TvCredits.Crew {
	// 	person.PopulatePersonByIdAndLanguage(crew.Id, language)
	// }
	// FINAL TRATAMENTO DAS PESSOAS DO CAST E CREW
	itemFind := GetSerieByIdAndLanguage(itemObj.Id, language)

	// Início tratamento para episódios de uma série
	var seasonsDetails []Season
	for _, season := range itemObj.Seasons {
		reqSeasonEpisodes := tmdb.GetTvSeason(itemObj.Id, season.SeasonNumber, language)

		var seasonReq Season
		json.NewDecoder(reqSeasonEpisodes.Body).Decode(&seasonReq)

		log.Println("TV EPISODES TOTAL: ", itemObj.NumberOfEpisodes)
		if itemObj.NumberOfEpisodes > 0 {
			// Getting cast from episode
			seasonEpisodesWithCredits := make([]Episode, 0)
			for _, episode := range seasonReq.Episodes {
				log.Println("TV - SEASON - EPISODE: ", itemObj.Id, seasonReq.SeasonNumber, episode.EpisodeNumber)
				reqTvCredits := tmdb.GetTvSeasonEpisodeCredits(itemObj.Id, season.SeasonNumber, episode.EpisodeNumber, language)
				json.NewDecoder(reqTvCredits.Body).Decode(&episode.TvEpisodeCredits)
				seasonEpisodesWithCredits = append(seasonEpisodesWithCredits, episode)
			}

			seasonReq.Episodes = seasonEpisodesWithCredits
		}

		seasonReq.EpisodeCount = season.EpisodeCount
		seasonReq.Overview = season.Overview
		seasonsDetails = append(seasonsDetails, seasonReq)
	}
	itemObj.Seasons = seasonsDetails
	// FINAL tratamento para episódios de uma série

	if itemFind.Id == 0 {

		for _, cast := range itemObj.TvCredits.Cast {
			person.PopulatePersonByIdAndLanguage(cast.Id, language)
		}

		for _, crew := range itemObj.TvCredits.Crew {
			person.PopulatePersonByIdAndLanguage(crew.Id, language)
		}

		log.Println("===>INSERT SERIE: ", itemObj.Id)
		InsertSerie(itemObj, language)
	} else {
		log.Println("===>UPDATE SERIE: ", itemObj.Id)
		UpdateSerie(itemObj, language)
	}
}

func PopulateSeries(language string, idGenre string) {

	parametro := parametro.GetByTipo("CARGA_TMDB_CONFIG")
	apiMaxPage := parametro.Options.TmdbMaxPageLoad

	for i := 1; i < apiMaxPage+1; i++ {
		// for i := 1; i < 2; i++ {
		log.Println("======> SERIE PAGE: ", language, i)
		page := strconv.Itoa(i)
		response := tmdb.GetDiscoverTvByLanguageGenreAndPage(language, idGenre, page)

		var result ResultSerie
		json.NewDecoder(response.Body).Decode(&result)

		for _, item := range result.Results {

			checkTvExist := GetSerieByIdAndLanguage(item.Id, common.LANGUAGE_PTBR)

			if checkTvExist.Id == 0 {
				itemObj := GetSerieDetailsOnTMDBApi(item.Id, language)
				PopulateSerieByLanguage(itemObj, language)

				itemObjBr := GetSerieDetailsOnTMDBApi(item.Id, common.LANGUAGE_PTBR)
				PopulateSerieByLanguage(itemObjBr, common.LANGUAGE_PTBR)
			}
		}
	}
}

func GetCountAll() int64 {
	return database.GetCountAllByColletcion(serieCollection)
}

func GetAll(skip int64, limit int64) []Serie {
	client, ctx, cancel := database.GetConnection()
	defer cancel()
	defer client.Disconnect(ctx)

	optionsFind := options.Find().SetLimit(limit).SetSkip(skip)
	cur, err := client.Database(os.Getenv("MONGO_DATABASE")).Collection(database.COLLECTION_SERIE).Find(context.TODO(), bson.M{}, optionsFind)
	if err != nil {
		log.Println(err)
	}

	series := make([]Serie, 0)
	for cur.Next(context.TODO()) {
		var movie Serie
		err := cur.Decode(&movie)
		if err != nil {
			log.Fatal(err)
		}

		series = append(series, movie)
	}

	cur.Close(context.TODO())

	return series
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
