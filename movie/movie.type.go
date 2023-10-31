package movie

import (
	"time"
)

type Genres struct {
	Id   int    `json:"id" bson:"id"`
	Name string `json:"name" bson:"name"`
}

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
	Popularity          float64               `json:"popularity" bson:"popularity"`
	Id                  int                   `json:"id" bson:"id"`
	Video               bool                  `json:"video" bson:"video"`
	VoteCount           int                   `json:"vote_count" bson:"vote_count"`
	VoteAverage         float64               `json:"vote_average" bson:"vote_average"`
	Title               string                `json:"title" bson:"title"`
	ReleaseDate         string                `json:"release_date,omitempty" bson:"release_date"`
	Runtime             int                   `json:"runtime,omitempty" bson:"runtime"`
	OriginalLanguage    string                `json:"original_language" bson:"original_language"`
	OriginalTitle       string                `json:"original_title" bson:"original_title"`
	BackdropPath        string                `json:"backdrop_path" bson:"backdrop_path"`
	Adult               bool                  `json:"adult,omitempty" bson:"adult"`
	Overview            string                `json:"overview" bson:"overview"`
	PosterPath          string                `json:"poster_path" bson:"poster_path"`
	MediaType           string                `json:"media_type" bson:"media_type"`
	Language            string                `json:"language" bson:"language"`
	SlugUrl             string                `json:"slugUrl,omitempty" bson:"slugUrl"`
	Slug                string                `json:"slug,omitempty" bson:"slug"`
	Updated             *time.Time            `json:"updated,omitempty" bson:"updated"`
	UpdatedNew          string                `json:"updatedNew,omitempty" bson:"updatedNew"`
	Genres              []Genres              `json:"genres" bson:"genres"`
	ProductionCompanies []ProductionCompanie  `json:"production_companies" bson:"production_companies"`
	ProductionCountries []ProductionCountries `json:"production_countries" bson:"production_countries"`
	SpokenLanguages     []SpokenLanguages     `json:"spoken_languages" bson:"spoken_languages"`
	MovieCredits        MovieCredits          `json:"credits" bson:"credits"`
}

type ResultMovie struct {
	Page         string  `json:"page"`
	TotalResults int     `json:"total_results"`
	TotalPages   int     `json:"total_pages"`
	Results      []Movie `json:"results"`
}
