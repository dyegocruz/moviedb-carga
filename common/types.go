package common

const (
	LANGUAGE_EN   = "en"
	LANGUAGE_PTBR = "pt-BR"
)

type CatalogCheck struct {
	Id       int    `json:"id" bson:"id"`
	Language string `json:"language" bson:"language"`
}
