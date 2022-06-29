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

// {
//   "tipo": "CARGA_TMDB_CONFIG",
//   "options": {
//     "tmdbHost": "https://api.themoviedb.org/3",
//     "tmdbApiKey": "26fe6f55e55736490dee0811901cccac",
//     "tmdbMaxPageLoad": 20
//   }
// }
