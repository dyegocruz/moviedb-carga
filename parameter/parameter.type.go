package parameter

type Options struct {
	TmdbHost        string `bson:"tmdbHost"`
	TmdbApiKey      string `bson:"tmdbApiKey"`
	TmdbMaxPageLoad int    `bson:"tmdbMaxPageLoad"`
}

type Parameter struct {
	ParamType string  `bson:"paramType"`
	Options   Options `bson:"options"`
}
