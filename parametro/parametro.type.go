package parametro

type Options struct {
	TmdbHost        string `bson:"tmdbHost"`
	TmdbApiKey      string `bson:"tmdbApiKey"`
	TmdbMaxPageLoad int    `bson:"tmdbMaxPageLoad"`
}

type Parametro struct {
	Tipo    string  `bson:"tipo"`
	Options Options `bson:"options"`
}
