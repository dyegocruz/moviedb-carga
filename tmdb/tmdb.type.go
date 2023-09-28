package tmdb

type ChangedElement struct {
	Id    int  `json:"id"`
	Adult bool `json:"adult"`
}

type ChangeResults struct {
	Results      []ChangedElement `json:"results"`
	Page         int              `json:"page"`
	TotalPages   int              `json:"total_pages"`
	TotalResults int              `json:"total_results"`
}

type TmdbDailyFile struct {
	Id           int `json:"id" bson:"id"`
	OriginalName int `json:"original_name" bson:"original_name"`
	Popularity   int `json:"popularity" bson:"popularity"`
}
