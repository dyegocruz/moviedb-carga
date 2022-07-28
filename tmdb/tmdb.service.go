package tmdb

import (
	"encoding/json"
	"moviedb/parametro"
	"moviedb/util"
	"net/http"
	"strconv"
)

const (
	DATATYPE_MOVIE  = "movie"
	DATATYPE_TV     = "tv"
	DATATYPE_PERSON = "person"
)

func getApiConfig() (string, string) {
	parametro := parametro.GetByTipo("CARGA_TMDB_CONFIG")
	apiKey := parametro.Options.TmdbApiKey
	apiHost := parametro.Options.TmdbHost

	return apiKey, apiHost
}

func GetChangesByDataType(dataType string) ChangeResults {
	apiKey, apiHost := getApiConfig()
	responseChange := util.HttpGet(apiHost + "/" + dataType + "/changes?api_key=" + apiKey + "&start_date=" + util.GetDateNowByFormatUrl())

	var changes ChangeResults
	json.NewDecoder(responseChange.Body).Decode(&changes)

	return changes
}

func GetDetailsByIdLanguageAndDataType(id int, language string, dataType string) *http.Response {
	apiKey, apiHost := getApiConfig()
	response := util.HttpGet(apiHost + "/" + dataType + "/" + strconv.Itoa(id) + "?api_key=" + apiKey + "&language=" + language + "&append_to_response=credits")
	return response
}

// func GetMovieCreditsByIdAndLanguage(id int, language string) *http.Response {
// 	apiKey, apiHost := getApiConfig()
// 	return util.HttpGet(apiHost + "/movie/" + strconv.Itoa(id) + "/credits?api_key=" + apiKey + "&language=" + language)
// }

// func GetTvCreditsByIdAndLanguage(id int, language string) *http.Response {
// 	apiKey, apiHost := getApiConfig()
// 	return util.HttpGet(apiHost + "/tv/" + strconv.Itoa(id) + "/credits?api_key=" + apiKey + "&language=" + language)
// }

// func GetPersonCreditsByIdAndLanguage(id int, language string) *http.Response {
// 	apiKey, apiHost := getApiConfig()
// 	return util.HttpGet(apiHost + "/person/" + strconv.Itoa(id) + "/combined_credits?api_key=" + apiKey + "&language=" + language)
// }

func GetDiscoverMoviesByLanguageGenreAndPage(language string, idGenre string, page string) *http.Response {
	apiKey, apiHost := getApiConfig()
	return util.HttpGet(apiHost + "/discover/movie?api_key=" + apiKey + "&language=" + language + "&sort_by=popularity.desc&include_adult=false&include_video=false&page=" + page + "&with_genres=" + idGenre)
}

func GetDiscoverTvByLanguageGenreAndPage(language string, idGenre string, page string) *http.Response {
	apiKey, apiHost := getApiConfig()
	return util.HttpGet(apiHost + "/discover/tv?api_key=" + apiKey + "&language=" + language + "&sort_by=popularity.desc&include_adult=false&include_video=false&page=" + page + "&with_genres=" + idGenre)
}

func GetPopularPerson(language string, page string) *http.Response {
	apiKey, apiHost := getApiConfig()
	return util.HttpGet(apiHost + "/person/popular?api_key=" + apiKey + "&language=" + language + "&sort_by=popularity.desc&include_adult=false&include_video=false&page=" + page)
}

func GetTvSeason(id int, seasonNumber int, language string) *http.Response {
	apiKey, apiHost := getApiConfig()
	return util.HttpGet(apiHost + "/tv/" + strconv.Itoa(id) + "/season/" + strconv.Itoa(seasonNumber) + "?api_key=" + apiKey + "&language=" + language)
}

func GetTvSeasonEpisodeCredits(id int, seasonNumber int, episode int, language string) *http.Response {
	apiKey, apiHost := getApiConfig()
	return util.HttpGet(apiHost + "/tv/" + strconv.Itoa(id) + "/season/" + strconv.Itoa(seasonNumber) + "/episode/" + strconv.Itoa(episode) + "/credits?api_key=" + apiKey + "&language=" + language)
}

// func GetTvSeasonEpisode(id int, seasonNumber int, episode int, language string) *http.Response {
// 	apiKey, apiHost := getApiConfig()
// 	return util.HttpGet(apiHost + "/tv/" + strconv.Itoa(id) + "/season/" + strconv.Itoa(seasonNumber) + "/episode/" + strconv.Itoa(episode) + "?api_key=" + apiKey + "&language=" + language + "&append_to_response=credits")
// }

func GetTvSeasonEpisode(id int, seasonNumber int, episode int, language string) *http.Response {
	apiKey, apiHost := getApiConfig()
	return util.HttpGet(apiHost + "/tv/" + strconv.Itoa(id) + "/season/" + strconv.Itoa(seasonNumber) + "/episode/" + strconv.Itoa(episode) + "?api_key=" + apiKey + "&language=" + language + "&append_to_response=credits")
}
