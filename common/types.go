package common

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
	MEDIA_TYPE_MOVIE      = "MOVIE"
	MEDIA_TYPE_TV         = "TV"
	MEDIA_TYPE_TV_EPISODE = "TV_EPISODE"
	MEDIA_TYPE_PERSON     = "PERSON"
)

type CatalogCheck struct {
	Id int `json:"id" bson:"id"`
}

type ResultAlternativeTitle struct {
	Id      int                `json:"id" bson:"id"`
	Results []AlternativeTitle `json:"results" bson:"results"`
}

type AlternativeTitle struct {
	Iso3166_1 string `json:"iso_3166_1" bson:"iso_3166_1"`
	Title     string `json:"title" bson:"title"`
	Type      string `json:"type" bson:"type"`
}
