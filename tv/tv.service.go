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
		if !serie.Adult {
			PopulateSerieByIdAndLanguage(serie.Id, common.LANGUAGE_PTBR)
			go PopulateSerieByIdAndLanguage(serie.Id, common.LANGUAGE_EN)
		}
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
	// reqCredits := tmdb.GetTvCreditsByIdAndLanguage(itemObj.Id, language)
	// json.NewDecoder(reqCredits.Body).Decode(&itemObj.TvCredits)
	// FINAL TRATAMENTO DAS PESSOAS DO CAST E CREW
	itemFind := GetSerieByIdAndLanguage(itemObj.Id, language)

	// Início tratamento para episódios de uma série
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

				if episode.AirDate != "" {
					air_date, err := time.Parse("2006-01-02", episode.AirDate)
					if err != nil {
						log.Println("Date converter error: ", err)
					}

					var now = time.Now()

					if air_date.After(now.AddDate(0, 0, -7)) {
						reqTvEpisode := tmdb.GetTvSeasonEpisode(itemObj.Id, season.SeasonNumber, episode.EpisodeNumber, language)
						json.NewDecoder(reqTvEpisode.Body).Decode(&episode)

						episode.Language = language

						log.Println("UPDATE TV - SEASON - EPISODE: ", itemObj.Id, seasonReq.SeasonNumber, episode.EpisodeNumber, episode.Id)
						UpdateEpisode(episode, language)
					} else {
						log.Println("BYPASS UPDATE TV - SEASON - EPISODE: ", itemObj.Id, seasonReq.SeasonNumber, episode.EpisodeNumber, episode.Id)
					}
				} else {
					log.Println("BYPASS UPDATE TV - SEASON - EPISODE: ", itemObj.Id, seasonReq.SeasonNumber, episode.EpisodeNumber, episode.Id)
				}
			}
		}

		seasonReq.EpisodeCount = season.EpisodeCount
		seasonReq.Overview = season.Overview
		seasonsDetails = append(seasonsDetails, seasonReq)
	}
	itemObj.Seasons = seasonsDetails
	// FINAL tratamento para episódios de uma série

	if itemFind.Id == 0 {

		for _, cast := range itemObj.TvCredits.Cast {
			person.PopulatePersonByIdAndLanguage(cast.Id, language, "Y")
		}

		for _, crew := range itemObj.TvCredits.Crew {
			person.PopulatePersonByIdAndLanguage(crew.Id, language, "Y")
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

func GetAll(skip int64, limit int64) []Serie {
	client, ctx, _ := database.GetConnection()

	projection := bson.M{"_id": 0, "genre_ids": 0, "slug": 0, "slugUrl": 0, "seasons.episodes": 0, "credits.cast.gender": 0, "credits.cast.knownfordepartment": 0, "credits.cast.popularity": 0, "credits.cast.originalname": 0, "credits.crew.originalname": 0, "credits.crew.knownfordepartment": 0, "credits.crew.department": 0, "credits.crew.popularity": 0, "credits.crew.gender": 0, "updated": 0, "updatedNew": 0}
	optionsFind := options.Find().SetLimit(limit).SetSkip(skip).SetProjection(projection)
	cur, err := client.Database(os.Getenv("MONGO_DATABASE")).Collection(database.COLLECTION_SERIE).Find(context.TODO(), bson.M{"id": bson.M{"$gt": 0}}, optionsFind)
	if err != nil {
		log.Println(err)
	}

	series := make([]Serie, 0)
	defer cur.Close(context.TODO())
	for cur.Next(context.TODO()) {
		var serie Serie
		err := cur.Decode(&serie)
		if err != nil {
			log.Fatal(err)
		}
		series = append(series, serie)
	}

	client.Disconnect(ctx)
	return series
}

func GetAllTest(batchSize int32) []Serie {
	client, ctx, _ := database.GetConnection()

	projection := bson.M{"_id": 0, "genre_ids": 0, "slug": 0, "slugUrl": 0, "seasons.episodes": 0, "credits.cast.gender": 0, "credits.cast.knownfordepartment": 0, "credits.cast.popularity": 0, "credits.cast.originalname": 0, "credits.crew.originalname": 0, "credits.crew.knownfordepartment": 0, "credits.crew.department": 0, "credits.crew.popularity": 0, "credits.crew.gender": 0, "updated": 0, "updatedNew": 0}
	optionsFind := options.Find().SetProjection(projection).SetBatchSize(batchSize).SetNoCursorTimeout(true)
	cur, err := client.Database(os.Getenv("MONGO_DATABASE")).Collection(database.COLLECTION_SERIE).Find(context.TODO(), bson.M{"id": bson.M{"$gt": 0}}, optionsFind)
	if err != nil {
		log.Println(err)
	}

	series := make([]Serie, 0)
	defer cur.Close(context.TODO())
	for cur.Next(context.TODO()) {
		var serie Serie
		err := cur.Decode(&serie)
		if err != nil {
			log.Fatal(err)
		}
		series = append(series, serie)
	}

	defer client.Disconnect(ctx)
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

func GetCountAllEpisodes() int64 {
	return database.GetCountAllByColletcion(database.COLLECTION_SERIE_EPISODE)
}

func GetAllEpisodes(skip int64, limit int64) []Episode {
	client, ctx, _ := database.GetConnection()

	projection := bson.M{"_id": 0, "id": 0, "production_code": 0, "vote_average": 0, "vote_count": 0, "credits.cast.gender": 0, "credits.cast.knownfordepartment": 0, "credits.cast.popularity": 0, "credits.cast.originalname": 0, "credits.crew.originalname": 0, "credits.crew.knownfordepartment": 0, "credits.crew.gender": 0}
	optionsFind := options.Find().SetLimit(limit).SetSkip(skip).SetProjection(projection)
	cur, err := client.Database(os.Getenv("MONGO_DATABASE")).Collection(database.COLLECTION_SERIE_EPISODE).Find(context.TODO(), bson.M{"id": bson.M{"$gt": 0}}, optionsFind)
	if err != nil {
		log.Println(err)
	}

	episodes := make([]Episode, 0)
	defer cur.Close(context.TODO())
	for cur.Next(context.TODO()) {
		var episode Episode
		err := cur.Decode(&episode)
		if err != nil {
			log.Fatal(err)
		}
		episodes = append(episodes, episode)
	}

	client.Disconnect(ctx)

	return episodes
}

func GetAllEpisodesTest(batchSize int32) []Episode {
	client, ctx, cancel := database.GetConnection()
	defer cancel()
	defer client.Disconnect(ctx)

	projection := bson.M{"_id": 0, "id": 0, "production_code": 0, "vote_average": 0, "vote_count": 0, "credits.cast.gender": 0, "credits.cast.knownfordepartment": 0, "credits.cast.popularity": 0, "credits.cast.originalname": 0, "credits.crew.originalname": 0, "credits.crew.knownfordepartment": 0, "credits.crew.gender": 0}
	optionsFind := options.Find().SetProjection(projection).SetBatchSize(batchSize).SetNoCursorTimeout(true)
	cur, err := client.Database(os.Getenv("MONGO_DATABASE")).Collection(database.COLLECTION_SERIE_EPISODE).Find(context.TODO(), bson.M{"id": bson.M{"$gt": 0}}, optionsFind)
	if err != nil {
		log.Println(err)
	}

	episodes := make([]Episode, 0)
	defer cur.Close(context.TODO())
	for cur.Next(context.TODO()) {
		var episode Episode
		err := cur.Decode(&episode)
		if err != nil {
			log.Fatal(err)
		}
		episodes = append(episodes, episode)
	}

	return episodes
}

func InsertEpisode(itemInsert Episode, language string) interface{} {

	client, ctx, cancel := database.GetConnection()
	defer cancel()
	defer client.Disconnect(ctx)

	result, err := client.Database(os.Getenv("MONGO_DATABASE")).Collection(database.COLLECTION_SERIE_EPISODE).InsertOne(context.TODO(), itemInsert)
	if err != nil {
		log.Println("EERRORRR")
		log.Println(err)
	}

	return result.InsertedID
}

func GetEpisodeByIdAndLanguage(id int, language string) Episode {

	client, ctx, cancel := database.GetConnection()
	defer cancel()
	defer client.Disconnect(ctx)

	var item Episode
	client.Database(os.Getenv("MONGO_DATABASE")).Collection(database.COLLECTION_SERIE_EPISODE).FindOne(context.TODO(), bson.M{"id": id, "language": language}).Decode(&item)

	return item
}

func GetEpisodeBySerieSeasonAndLanguage(showId int, seasonNumber int, language string) []Episode {

	client, ctx, cancel := database.GetConnection()
	defer cancel()
	defer client.Disconnect(ctx)

	cur, err := client.Database(os.Getenv("MONGO_DATABASE")).Collection(database.COLLECTION_SERIE_EPISODE).Find(context.TODO(), bson.M{"show_id": showId, "season_number": seasonNumber, "language": language})
	if err != nil {
		log.Println(err)
	}

	episodes := make([]Episode, 0)
	cur.All(context.TODO(), &episodes)
	cur.Close(context.TODO())

	return episodes
}

func UpdateEpisode(espisode Episode, language string) {

	client, ctx, cancel := database.GetConnection()
	defer cancel()
	defer client.Disconnect(ctx)

	client.Database(os.Getenv("MONGO_DATABASE")).Collection(database.COLLECTION_SERIE_EPISODE).UpdateOne(context.TODO(), bson.M{"id": espisode.Id, "language": language}, bson.M{
		"$set": espisode,
	})
}

func GetCountAll() int64 {
	return database.GetCountAllByColletcion(serieCollection)
}

func GenerateTvCatalogCheck(language string) map[int]common.CatalogCheck {
	return database.GenerateCatalogCheck(serieCollection, language)
}
