package parameter

type Options struct {
	TmdbHost              string `bson:"tmdbHost"`
	TmdbApiKey            string `bson:"tmdbApiKey"`
	TmdbMaxPageLoad       int    `bson:"tmdbMaxPageLoad"`
	EnableUpdateCatalogDb bool   `bson:"enableUpdateCatalogDb"`
	EnableChargeCache     bool   `bson:"enableChargeCache"`
}

type Parameter struct {
	ParamType string  `bson:"paramType"`
	Options   Options `bson:"options"`
}
