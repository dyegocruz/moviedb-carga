package tv

import (
	"context"
	"encoding/json"
	"log"
	"moviedb/common"
	"moviedb/database"
	"moviedb/parameter"
	"moviedb/queue"

	"moviedb/person"
	"moviedb/tmdb"
	"strconv"
	"time"

	"github.com/gosimple/slug"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var serieCollectionString = database.COLLECTION_SERIE
var serieCollection *mongo.Collection = database.GetCollection(database.DB, serieCollectionString)

var serieEpisodeCollectionString = database.COLLECTION_SERIE_EPISODE
var serieEpisodeCollection *mongo.Collection = database.GetCollection(database.DB, serieEpisodeCollectionString)

func CheckTvChanges() {
	// Initialize RabbitMQ connection
	rmq, err := queue.NewRabbitMQ()
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %s", err)
	}
	defer rmq.Close()

	tvChanges := tmdb.GetChangesByDataType(tmdb.DATATYPE_TV, 1)

	for _, serie := range tvChanges {

		// Publish a message
		err = rmq.PublishJSON(queue.QueueCatalogProcess, queue.CatalogProcessMessage{Id: serie.Id, MediaType: common.MEDIA_TYPE_TV})
		if err != nil {
			log.Fatalf("Failed to publish a message: %s", err)
		}

		log.Println("Message published successfully!")
	}

	log.Println("CheckTvChanges CONCLUDED")
}

func PopulateSerieByIdAndLanguage(id int, language string) {
	itemObj := GetSerieDetailsOnTMDBApi(id, language)
	log.Println(itemObj.Id, itemObj.Title, itemObj.OriginalTitle, itemObj.OriginalLanguage, itemObj.FirstAirDate, itemObj.Popularity)
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

	itemFind := GetSerieByIdAndLanguage(itemObj.Id, language)

	// INIT treatment of tv's episodes
	var seasonsDetails []Season
	for _, season := range itemObj.Seasons {
		reqSeasonEpisodes := tmdb.GetTvSeason(itemObj.Id, season.SeasonNumber, language)

		var seasonReq Season
		json.NewDecoder(reqSeasonEpisodes.Body).Decode(&seasonReq)

		for _, episode := range seasonReq.Episodes {
			findEpisode := GetEpisodeByIdAndLanguage(episode.Id, language)

			if findEpisode.Id == 0 {
				reqTvEpisode := tmdb.GetTvSeasonEpisode(itemObj.Id, season.SeasonNumber, episode.EpisodeNumber, language)
				json.NewDecoder(reqTvEpisode.Body).Decode(&episode)

				episode.Language = language

				log.Println("INSERT TV - SEASON - EPISODE: ", itemObj.Id, seasonReq.SeasonNumber, episode.EpisodeNumber, episode.Id)
				InsertEpisode(episode, language)
			} else {
				// check and update just the last season episodes
				if episode.SeasonNumber == itemObj.NumberOfSeasons {
					// cehck and update just the last 10 episodes of the last season
					if episode.EpisodeNumber >= (season.EpisodeCount - 10) {
						// UPDATE EPISODE
						reqTvEpisode := tmdb.GetTvSeasonEpisode(itemObj.Id, season.SeasonNumber, episode.EpisodeNumber, language)
						json.NewDecoder(reqTvEpisode.Body).Decode(&episode)

						episode.Language = language

						log.Println("UPDATE TV - SEASON - EPISODE: ", itemObj.Id, seasonReq.SeasonNumber, episode.EpisodeNumber, episode.Id)
						UpdateEpisode(episode, language)
					}
				}
			}
		}

		seasonReq.EpisodeCount = season.EpisodeCount
		seasonReq.Overview = season.Overview
		seasonsDetails = append(seasonsDetails, seasonReq)
	}
	itemObj.Seasons = seasonsDetails
	// FINAL treatment of tv's episodes

	if itemFind.Id == 0 {

		for _, cast := range itemObj.TvCredits.Cast {
			person.PopulatePersonByIdAndLanguage(cast.Id, language, "Y")
		}

		for _, crew := range itemObj.TvCredits.Crew {
			person.PopulatePersonByIdAndLanguage(crew.Id, language, "Y")
		}

		if itemObj.Id > 0 {
			log.Println("===>INSERT TV: ", itemObj.Id)
			InsertSerie(itemObj, language)
		}

	} else {
		log.Println("===>UPDATE TV: ", itemObj.Id)
		UpdateSerie(itemObj, language)
	}
}

func PopulateSeries(language string, idGenre string) {

	parametro := parameter.GetByType("CHARGE_TMDB_CONFIG")
	apiMaxPage := parametro.Options.TmdbMaxPageLoad

	for i := 1; i < apiMaxPage+1; i++ {
		log.Println("======> TV PAGE: ", language, i)
		page := strconv.Itoa(i)
		response := tmdb.GetDiscoverTvByLanguageGenreAndPage(language, idGenre, page)

		var result ResultSerie
		json.NewDecoder(response.Body).Decode(&result)

		for _, item := range result.Results {
			if item.Id > 0 {
				checkTvExist := GetSerieByIdAndLanguage(item.Id, common.LANGUAGE_PTBR)

				if checkTvExist.Id == 0 {
					itemObjBr := GetSerieDetailsOnTMDBApi(item.Id, common.LANGUAGE_PTBR)
					PopulateSerieByLanguage(itemObjBr, common.LANGUAGE_PTBR)

					itemObj := GetSerieDetailsOnTMDBApi(item.Id, language)
					go PopulateSerieByLanguage(itemObj, language)
				}
			}
		}
	}
}

func GetAllByIds(ids []int) []interface{} {

	ctx2 := context.Background()

	projection := bson.M{"_id": 0, "slug": 0, "slugUrl": 0, "adult": 0, "seasons.episodes": 0, "credits.cast.gender": 0, "credits.cast.popularity": 0, "credits.cast.originalname": 0, "credits.crew.originalname": 0, "credits.crew.knownfordepartment": 0, "credits.crew.popularity": 0, "credits.crew.gender": 0, "updated": 0, "updatedNew": 0, "created_by.credit_id": 0, "created_by.gender": 0}
	optionsFind := options.Find().SetSort(bson.D{{Key: "id", Value: 1}, {Key: "language", Value: 1}}).SetProjection(projection)
	cur, err := serieCollection.Find(ctx2, bson.M{"id": bson.M{"$in": ids}}, optionsFind)
	if err != nil {
		log.Println(err)
	}

	series := make([]interface{}, 0)
	for cur.Next(ctx2) {
		var serie Serie
		err := cur.Decode(&serie)
		if err != nil {
			log.Fatal(err)
		}
		series = append(series, serie)
	}

	return series
}

func GetCatalogSearchIn(ids []int) []Serie {

	ctx2 := context.TODO()

	projection := bson.M{"_id": 0, "id": 1, "language": 1, "original_title": 1, "original_language": 1, "title": 1, "poster_path": 1, "first_air_date": 1, "popularity": 1}
	optionsFind := options.Find().SetSort(bson.D{{Key: "id", Value: 1}}).SetProjection(projection)
	cur, err := serieCollection.Find(ctx2, bson.M{"id": bson.M{"$in": ids}}, optionsFind)
	if err != nil {
		log.Println(err)
	}

	series := make([]Serie, 0)
	for cur.Next(ctx2) {
		var serie Serie
		err := cur.Decode(&serie)
		if err != nil {
			log.Fatal(err)
		}
		series = append(series, serie)
	}

	return series
}

func GetSerieByIdAndLanguage(id int, language string) Serie {

	var item Serie
	serieCollection.FindOne(context.TODO(), bson.M{"id": id, "language": language}).Decode(&item)

	return item
}

func InsertSerie(itemInsert Serie, language string) interface{} {

	result, err := serieCollection.InsertOne(context.TODO(), itemInsert)
	if err != nil {
		log.Println("EERRORRR")
		log.Println(err)
	}

	return result.InsertedID
}

func UpdateSerie(serie Serie, language string) {
	serieCollection.UpdateOne(context.TODO(), bson.M{"id": serie.Id, "language": language}, bson.M{
		"$set": serie,
	})
}

func DeleteSerie(id int) {
	serieCollection.DeleteMany(context.TODO(), bson.M{"id": id})
}

func InsertEpisode(itemInsert Episode, language string) interface{} {

	result, err := serieEpisodeCollection.InsertOne(context.TODO(), itemInsert)
	if err != nil {
		log.Println("EERRORRR")
		log.Println(err)
	}

	return result.InsertedID
}

func GetEpisodeByIdAndLanguage(id int, language string) Episode {

	var item Episode
	serieEpisodeCollection.FindOne(context.TODO(), bson.M{"id": id, "language": language}).Decode(&item)

	return item
}

func GetEpisodeBySerieSeasonAndLanguage(showId int, seasonNumber int, language string) []Episode {

	cur, err := serieEpisodeCollection.Find(context.TODO(), bson.M{"show_id": showId, "season_number": seasonNumber, "language": language})
	if err != nil {
		log.Println(err)
	}

	episodes := make([]Episode, 0)
	cur.All(context.TODO(), &episodes)
	cur.Close(context.TODO())

	return episodes
}

func UpdateEpisode(espisode Episode, language string) {

	serieEpisodeCollection.UpdateOne(context.TODO(), bson.M{"id": espisode.Id, "language": language}, bson.M{
		"$set": espisode,
	})
}

func DeleteSerieEpisodes(showId int) {
	serieEpisodeCollection.DeleteMany(context.TODO(), bson.M{"show_id": showId})
}

func GetCountAll() int64 {
	return database.GetCountAllByColletcion(serieCollectionString)
}

func GenerateTvCatalogCheck(language string) map[int]common.CatalogCheck {
	return database.GenerateCatalogCheck(serieCollectionString, language)
}

func GenerateTvEpisodesCatalogCheck(language string) map[int]common.CatalogCheck {
	return database.GenerateCatalogCheck(serieEpisodeCollectionString, language)
}
