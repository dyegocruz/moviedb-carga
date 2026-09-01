package common

const (
	DATATYPE_MOVIE  = "movie"
	DATATYPE_TV     = "tv"
	DATATYPE_PERSON = "person"
)

const (
	LANGUAGE_EN   = "en"
	LANGUAGE_PTBR = "pt-BR"
	LANGUAGE_JA   = "ja"
)

const (
	LANGUAGE_ISO_EN = "US"
	LANGUAGE_ISO_BR = "BR"
	LANGUAGE_ISO_JP = "JP"
)

const (
	ALTERNATIVE_TITLE_TYPE_ROMAJI       = "Romaji"
	ALTERNATIVE_TITLE_TYPE_NICKNAME     = "Nickname"
	ALTERNATIVE_TITLE_TYPE_ABBREVIATION = "Abbreviation"
)

const (
	MEDIA_TYPE_MOVIE         = "MOVIE"
	MEDIA_TYPE_TV            = "TV"
	MEDIA_TYPE_TV_EPISODE    = "TV_EPISODE"
	MEDIA_TYPE_PERSON        = "PERSON"
	MEDIA_TYPE_CATALOG_CHECK = "CATALOG_CHECK"
)

type CatalogCheck struct {
	Id int `json:"id" bson:"id"`
}

type AlternativeMovieTitlesResponse struct {
	Titles []AlternativeTitle `json:"titles" bson:"titles"`
}

type AlternativeTvTitlesResponse struct {
	Results []AlternativeTitle `json:"results" bson:"results"`
}

type AlternativeTitle struct {
	Iso3166_1 string `json:"iso_3166_1" bson:"iso_3166_1"`
	Title     string `json:"title" bson:"title"`
	Type      string `json:"type" bson:"type"`
}

type Genres struct {
	Id   int    `json:"id" bson:"id"`
	Name string `json:"name" bson:"name"`
}

type LocalizationMovieTv struct {
	Locale     string   `json:"locale" bson:"locale"`
	Title      string   `json:"title" bson:"title"`
	Synopsis   string   `json:"synopsis" bson:"synopsis"`
	Genres     []Genres `json:"genres" bson:"genres"`
	PosterPath string   `json:"poster_path" bson:"poster_path"`
}

type LocalizationCommon struct {
	Locale   string `json:"locale" bson:"locale"`
	Title    string `json:"title" bson:"title"`
	Synopsis string `json:"synopsis" bson:"synopsis"`
}

type LocalizationPerson struct {
	Locale    string `json:"locale" bson:"locale"`
	Name      string `json:"name" bson:"name"`
	Biography string `json:"biography" bson:"biography"`
}
