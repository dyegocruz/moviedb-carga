package tmdb

import (
	"encoding/json"
	"moviedb/parameter"
	"moviedb/util"

	"net/http"
	"strconv"
)

const (
	DATATYPE_MOVIE  = "movie"
	DATATYPE_TV     = "tv"
	DATATYPE_PERSON = "person"
)

// ParameterProvider abstracts parameter lookup so the service can be tested
// without a real Mongo connection.
type ParameterProvider interface {
	GetByType(paramType string) parameter.Parameter
}

type Service struct {
	parameter ParameterProvider
}

func NewService(parameterService ParameterProvider) *Service {
	return &Service{parameter: parameterService}
}

func (s *Service) getApiConfig() (string, string) {
	param := s.parameter.GetByType("CHARGE_TMDB_CONFIG")
	apiKey := param.Options.TmdbApiKey
	apiHost := param.Options.TmdbHost

	return apiKey, apiHost
}

func (s *Service) GetChangesByDataType(dataType string, page int) []ChangedElement {
	apiKey, apiHost := s.getApiConfig()

	urlGetChanges := apiHost + "/" + dataType + "/changes?api_key=" + apiKey + "&start_date=" + util.GetDateNowByFormatUrl() + "&page=" + strconv.Itoa(page)
	responseChange := util.HttpGet(urlGetChanges)

	var changes ChangeResults
	json.NewDecoder(responseChange.Body).Decode(&changes)

	if page < changes.TotalPages {
		changes.Results = append(changes.Results, s.GetChangesByDataType(dataType, page+1)...)
	}

	return changes.Results
}

func (s *Service) GetDetailsByIdLanguageAndDataType(id int, language string, dataType string) *http.Response {
	apiKey, apiHost := s.getApiConfig()

	appendResponse := "&append_to_response=credits,alternative_titles"

	if dataType == DATATYPE_PERSON {
		appendResponse = "&append_to_response=combined_credits,alternative_titles"
	}

	response := util.HttpGet(apiHost + "/" + dataType + "/" + strconv.Itoa(id) + "?api_key=" + apiKey + "&language=" + language + appendResponse)
	return response
}

func (s *Service) GetAlternativeTitlesByIdAndDataType(id int, dataType string) *http.Response {
	apiKey, apiHost := s.getApiConfig()

	response := util.HttpGet(apiHost + "/" + dataType + "/" + strconv.Itoa(id) + "/alternative_titles?api_key=" + apiKey)
	return response
}

func (s *Service) GetDiscoverMoviesByLanguageGenreAndPage(language string, idGenre string, page string) *http.Response {
	apiKey, apiHost := s.getApiConfig()
	return util.HttpGet(apiHost + "/discover/movie?api_key=" + apiKey + "&language=" + language + "&sort_by=popularity.desc&include_adult=false&include_video=false&page=" + page + "&with_genres=" + idGenre)
}

func (s *Service) GetDiscoverTvByLanguageGenreAndPage(language string, idGenre string, page string) *http.Response {
	apiKey, apiHost := s.getApiConfig()
	return util.HttpGet(apiHost + "/discover/tv?api_key=" + apiKey + "&language=" + language + "&sort_by=popularity.desc&include_adult=false&include_video=false&page=" + page + "&with_genres=" + idGenre)
}

func (s *Service) GetPopularPerson(language string, page string) *http.Response {
	apiKey, apiHost := s.getApiConfig()
	return util.HttpGet(apiHost + "/person/popular?api_key=" + apiKey + "&language=" + language + "&sort_by=popularity.desc&include_adult=false&include_video=false&page=" + page)
}

func (s *Service) GetTvSeason(id int, seasonNumber int, language string) *http.Response {
	apiKey, apiHost := s.getApiConfig()
	return util.HttpGet(apiHost + "/tv/" + strconv.Itoa(id) + "/season/" + strconv.Itoa(seasonNumber) + "?api_key=" + apiKey + "&language=" + language)
}

func (s *Service) GetTvSeasonEpisodeCredits(id int, seasonNumber int, episode int, language string) *http.Response {
	apiKey, apiHost := s.getApiConfig()
	return util.HttpGet(apiHost + "/tv/" + strconv.Itoa(id) + "/season/" + strconv.Itoa(seasonNumber) + "/episode/" + strconv.Itoa(episode) + "/credits?api_key=" + apiKey + "&language=" + language)
}

func (s *Service) GetTvSeasonEpisode(id int, seasonNumber int, episode int, language string) *http.Response {
	apiKey, apiHost := s.getApiConfig()
	return util.HttpGet(apiHost + "/tv/" + strconv.Itoa(id) + "/season/" + strconv.Itoa(seasonNumber) + "/episode/" + strconv.Itoa(episode) + "?api_key=" + apiKey + "&language=" + language + "&append_to_response=credits")
}

func (s *Service) MaxPageLoad() int {
	param := s.parameter.GetByType("CHARGE_TMDB_CONFIG")
	return param.Options.TmdbMaxPageLoad
}
