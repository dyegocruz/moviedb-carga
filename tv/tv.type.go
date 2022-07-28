package tv

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

// type EpisodeToAir struct {
type Episode struct {
	AirDate          string           `json:"air_date,omitempty" bson:"air_date"`
	EpisodeNumber    int              `json:"episode_number" bson:"episode_number"`
	Id               int              `json:"id" bson:"id"`
	Language         string           `json:"language" bson:"language"`
	Name             string           `json:"name" bson:"name"`
	Overview         string           `json:"overview" bson:"overview"`
	ProdctionCode    string           `json:"production_code" bson:"production_code"`
	SeasonNumber     int              `json:"season_number" bson:"season_number"`
	ShowId           int              `json:"show_id" bson:"show_id"`
	StillPath        string           `json:"still_path" bson:"still_path"`
	VoteAverage      float64          `json:"vote_average" bson:"vote_average"`
	VoteCount        float64          `json:"vote_count" bson:"vote_count"`
	TvEpisodeCredits TvEpisodeCredits `json:"credits" bson:"credits"`
}

type Network struct {
	Id            string `json:"name" bson:"name"`
	Name          int    `json:"id" bson:"id"`
	LogoPath      string `json:"logo_path" bson:"logo_path"`
	OriginCountry string `json:"origin_country" bson:"origin_country"`
}

type Season struct {
	AirDate      string    `json:"air_date,omitempty" bson:"air_date"`
	EpisodeCount int       `json:"episode_count" bson:"episode_count"`
	Id           int       `json:"id" bson:"id"`
	Name         string    `json:"name" bson:"name"`
	Overview     string    `json:"overview" bson:"overview"`
	PosterPath   string    `json:"poster_path" bson:"poster_path"`
	SeasonNumber int       `json:"season_number" bson:"season_number"`
	Episodes     []Episode `json:"episodes" bson:"episodes"`
}

type CreatedBy struct {
	Id          int    `json:"id" bson:"id"`
	CreditId    string `json:"credit_id" bson:"credit_id"`
	Name        string `json:"name" bson:"name"`
	Genre       int    `json:"gender" bson:"gender"`
	ProfilePath string `json:"profile_path" bson:"profile_path"`
}

type TvCast struct {
	Gender             int     `json:"gender"`
	Id                 int     `json:"id" bson:"id"`
	KnownForDepartment string  `json:"known_for_department"`
	Name               string  `json:"name"`
	OriginalName       string  `json:"original_name"`
	Popularity         float64 `json:"popularity"`
	ProfilePath        string  `json:"profile_path"`
	Character          string  `json:"character" bson:"character"`
	Order              int     `json:"order" bson:"order"`
}

type TvCrew struct {
	Gender             int     `json:"gender"`
	Id                 int     `json:"id" bson:"id"`
	KnownForDepartment string  `json:"known_for_department"`
	Name               string  `json:"name"`
	OriginalName       string  `json:"original_name"`
	Popularity         float64 `json:"popularity"`
	ProfilePath        string  `json:"profile_path"`
	Department         string  `json:"department" bson:"department"`
	Job                string  `json:"job" bson:"job"`
}

type TvGuestStar struct {
	Id        int    `json:"id" bson:"id"`
	Character string `json:"character" bson:"character"`
	Order     int    `json:"order" bson:"order"`
}

type TvCredits struct {
	Cast []TvCast `json:"cast" bson:"cast"`
	Crew []TvCrew `json:"crew" bson:"crew"`
}

type TvEpisodeCredits struct {
	Cast        []TvCast      `json:"cast" bson:"cast"`
	Crew        []TvCrew      `json:"crew" bson:"crew"`
	TvGuestStar []TvGuestStar `json:"guest_stars" bson:"guest_stars"`
}

type Serie struct {
	Popularity          float64               `json:"popularity" bson:"popularity"`
	Id                  int                   `json:"id" bson:"id"`
	Video               bool                  `json:"video" bson:"video"`
	VoteCount           int                   `json:"vote_count" bson:"voteCount"`
	VoteAverage         float64               `json:"vote_average" bson:"voteAverage"`
	FirstAirDate        string                `json:"first_air_date,omitempty" bson:"first_air_date"`
	LastAirDate         string                `json:"last_air_date,omitempty" bson:"last_air_date"`
	LastEpisodeToAir    Episode               `json:"last_episode_to_air" bson:"last_episode_to_air"`
	NextEpisodeToAir    Episode               `json:"next_episode_to_air" bson:"next_episode_to_air"`
	OriginalLanguage    string                `json:"original_language" bson:"original_language"`
	Title               string                `json:"name" bson:"title"`
	OriginalTitle       string                `json:"original_name" bson:"original_title"`
	Networks            []Network             `json:"networks" bson:"networks"`
	NumberOfEpisodes    int                   `json:"number_of_episodes" bson:"number_of_episodes"`
	NumberOfSeasons     int                   `json:"number_of_seasons" bson:"number_of_seasons"`
	Seasons             []Season              `json:"seasons" bson:"seasons"`
	GenreIds            []int                 `json:"genre_ids" bson:"genre_ids"`
	BackdropPath        string                `json:"backdrop_path" bson:"backdrop_path"`
	Adult               bool                  `json:"adult" bson:"adult"`
	Overview            string                `json:"overview" bson:"overview"`
	PosterPath          string                `json:"poster_path" bson:"poster_path"`
	EpisodeRunTime      []int                 `json:"episode_run_time" bson:"episode_run_time"`
	Homepage            string                `json:"homepage" bson:"homepage"`
	InProduction        bool                  `json:"in_production" bson:"in_production"`
	MediaType           string                `json:"media_type" bson:"media_type"`
	Language            string                `json:"language" bson:"language"`
	Status              string                `json:"status" bson:"status"`
	Type                string                `json:"type" bson:"type"`
	SlugUrl             string                `json:"slugUrl" bson:"slugUrl"`
	Slug                string                `json:"slug" bson:"slug"`
	Updated             time.Time             `json:"updated" bson:"updated"`
	UpdatedNew          string                `json:"updatedNew" bson:"updatedNew"`
	OriginCountry       []string              `json:"origin_country" bson:"origin_country"`
	CreatedBy           []CreatedBy           `json:"created_by" bson:"created_by"`
	Genres              []Genres              `json:"genres" bson:"genres"`
	ProductionCompanies []ProductionCompanie  `json:"production_companies" bson:"production_companies"`
	ProductionCountries []ProductionCountries `json:"production_countries" bson:"production_countries"`
	Languages           []string              `json:"languages" bson:"languages"`
	TvCredits           TvCredits             `json:"credits" bson:"credits"`
}

type ResultSerie struct {
	Page         string  `json:"page"`
	TotalResults int     `json:"total_results"`
	TotalPages   int     `json:"total_pages"`
	Results      []Serie `json:"results"`
}
