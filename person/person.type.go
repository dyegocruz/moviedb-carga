package person

import (
	"time"
)

type Cast struct {
	ID        int    `json:"id" bson:"id"`
	CreditID  string `json:"credit_id" bson:"credit_id"`
	MediaType string `json:"media_type" bson:"media_type"`
}

type Crew struct {
	ID         int    `json:"id" bson:"id"`
	Department string `json:"department" bson:"department"`
	Job        string `json:"job" bson:"job"`
	MediaType  string `json:"media_type" bson:"media_type"`
}

type Credits struct {
	Cast []Cast `json:"cast" bson:"cast"`
	Crew []Crew `json:"crew" bson:"crew"`
}

type Person struct {
	Popularity        float64   `json:"popularity" bson:"popularity"`
	Id                int       `json:"id" bson:"id"`
	Gender            int       `json:"gender" bson:"gender"`
	KnowForDepartment string    `json:"known_for_department" bson:"known_for_department"`
	Name              string    `json:"name" bson:"name"`
	ProfilepPath      string    `json:"profile_path" bson:"profile_path"`
	Biography         string    `json:"biography" bson:"biography"`
	Birthday          string    `json:"birthday" bson:"birthday"`
	Deathday          string    `json:"deathday" bson:"deathday"`
	Language          string    `json:"language" bson:"language"`
	SlugUrl           string    `json:"slugUrl" bson:"slugUrl"`
	Slug              string    `json:"slug" bson:"slug"`
	Updated           time.Time `json:"updated" bson:"updated"`
	UpdatedNew        string    `json:"updatedNew" bson:"updatedNew"`
	OriginCountry     []string  `json:"origin_country" bson:"origin_country"`
	Languages         []string  `json:"languages" bson:"languages"`
	Credits           Credits   `json:"credits" bson:"credits"`
}

type ResultPerson struct {
	Page         string   `json:"page"`
	TotalResults int      `json:"total_results"`
	TotalPages   int      `json:"total_pages"`
	Results      []Person `json:"results"`
}
