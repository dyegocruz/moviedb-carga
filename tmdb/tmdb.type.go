package tmdb

type ChangedElement struct {
	Id    int  `json:"id"`
	Adult bool `json:"adult"`
}

type ChangeResults struct {
	Results []ChangedElement `json:"results"`
}
