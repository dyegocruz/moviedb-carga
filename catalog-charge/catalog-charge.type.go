package catalogCharge

type CatalogSearch struct {
	Id               int        `json:"id" bson:"id"`
	Name             string     `json:"name" bson:"name"`
	OriginalTitle    string     `json:"originalTitle" bson:"original_title"`
	OriginalLanguage string     `json:"originalLanguage" bson:"original_language"`
	ProfilePath      string     `json:"profilePath" bson:"profile_path"`
	CatalogType      string     `json:"catalogType"`
	FirstAirDate     string     `json:"firstAirDate,omitempty" bson:"first_air_date"`
	ReleaseDate      string     `json:"releaseDate,omitempty" bson:"release_date"`
	Popularity       float64    `json:"popularity" bson:"popularity"`
	Locations        []Location `json:"locations" bson:"locations"`
}

type Location struct {
	Title      string `json:"title" bson:"title"`
	Language   string `json:"language" bson:"language"`
	PosterPath string `json:"posterPath" bson:"poster_path"`
}
