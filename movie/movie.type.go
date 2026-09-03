package movie

import (
	"moviedb/common"
)

type ProductionCompanie struct {
	Id            int    `json:"id" bson:"id"`
	LogoPath      string `json:"logo_path" bson:"logo_path"`
	Name          string `json:"name" bson:"name"`
	OriginCountry string `json:"origin_country" bson:"origin_country"`
}

type ProductionCountries struct {
	Iso  string `json:"iso_3166_1" bson:"iso_3166_1"`
	Name string `json:"name" bson:"name"`
}

type SpokenLanguages struct {
	Iso  string `json:"iso_639_1" bson:"iso_639_1"`
	Name string `json:"name" bson:"name"`
}

type MovieCast struct {
	Gender             int     `json:"gender,omitempty"`
	Id                 int     `json:"id" bson:"id"`
	KnownForDepartment string  `json:"known_for_department,omitempty"`
	Name               string  `json:"name"`
	OriginalName       string  `json:"original_name,omitempty"`
	Popularity         float64 `json:"popularity,omitempty"`
	ProfilePath        string  `json:"profile_path"`
	Character          string  `json:"character" bson:"character"`
	Order              int     `json:"order"`
}

type MovieCrew struct {
	Gender             int     `json:"gender,omitempty"`
	Id                 int     `json:"id" bson:"id"`
	KnownForDepartment string  `json:"known_for_department,omitempty"`
	Name               string  `json:"name"`
	OriginalName       string  `json:"original_name,omitempty"`
	Popularity         float64 `json:"popularity,omitempty"`
	ProfilePath        string  `json:"profile_path"`
	Department         string  `json:"department,omitempty" bson:"department,omitempty"`
	Job                string  `json:"job" bson:"job"`
}

type MovieCredits struct {
	Cast []MovieCast `json:"cast" bson:"cast"`
	Crew []MovieCrew `json:"crew" bson:"crew"`
}

type Movie struct {
	Title               string                       `json:"title"`
	Overview            string                       `json:"overview"`
	PosterPath          string                       `json:"poster_path" bson:"poster_path"`
	Popularity          float64                      `json:"popularity" bson:"popularity"`
	Genres              []common.Genres              `json:"genres"`
	Id                  int                          `json:"id" bson:"id"`
	Video               bool                         `json:"video" bson:"video"`
	VoteCount           int                          `json:"vote_count" bson:"vote_count"`
	VoteAverage         float64                      `json:"vote_average" bson:"vote_average"`
	Localizations       []common.LocalizationMovieTv `json:"localizations" bson:"localizations"`
	ReleaseDate         string                       `json:"release_date,omitempty" bson:"release_date"`
	Runtime             int                          `json:"runtime,omitempty" bson:"runtime"`
	OriginalLanguage    string                       `json:"original_language" bson:"original_language"`
	OriginalTitle       string                       `json:"original_title" bson:"original_title"`
	BackdropPath        string                       `json:"backdrop_path" bson:"backdrop_path"`
	Adult               bool                         `json:"adult,omitempty" bson:"adult"`
	MediaType           string                       `json:"media_type" bson:"media_type"`
	Language            string                       `json:"language" bson:"language"`
	UpdatedAt           string                       `json:"updated_at,omitempty" bson:"updated_at"`
	ProductionCompanies []ProductionCompanie         `json:"production_companies" bson:"production_companies"`
	ProductionCountries []ProductionCountries        `json:"production_countries" bson:"production_countries"`
	SpokenLanguages     []SpokenLanguages            `json:"spoken_languages" bson:"spoken_languages"`
	MovieCredits        MovieCredits                 `json:"credits" bson:"credits"`
	AlternativeTitles   AlternativeTitlesMovie       `json:"alternative_titles" bson:"-"`
	AlternativeTitlesDb []common.AlternativeTitle    `json:"alternative_titles_db" bson:"alternative_titles_db"`
}
type AlternativeTitlesMovie struct {
	Titles []common.AlternativeTitle `json:"titles" bson:"titles"`
}

type ResultMovie struct {
	Page         string  `json:"page"`
	TotalResults int     `json:"total_results"`
	TotalPages   int     `json:"total_pages"`
	Results      []Movie `json:"results"`
}
