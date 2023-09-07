package common

const (
	LANGUAGE_EN   = "en"
	LANGUAGE_PTBR = "pt-BR"
)

const (
	MEDIA_TYPE_MOVIE  = "MOVIE"
	MEDIA_TYPE_TV     = "TV"
	MEDIA_TYPE_PERSON = "PERSON"
)

type CatalogCheck struct {
	Id       int    `json:"id" bson:"id"`
	Language string `json:"language" bson:"language"`
}
