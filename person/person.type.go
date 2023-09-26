package person

import (
	"time"
)

type Cast struct {
	ID        int    `json:"id" bson:"id"`
	CreditID  string `json:"credit_id,omitempty" bson:"credit_id"`
	Character string `json:"character" bson:"character"`
	MediaType string `json:"media_type" bson:"media_type"`
}

type Crew struct {
	ID         int    `json:"id" bson:"id"`
	Department string `json:"department,omitempty" bson:"department"`
	Job        string `json:"job" bson:"job"`
	MediaType  string `json:"media_type" bson:"media_type"`
}

type Credits struct {
	Cast []Cast `json:"cast" bson:"cast"`
	Crew []Crew `json:"crew" bson:"crew"`
}

type Person struct {
	Popularity        float64    `json:"popularity,omitempty" bson:"popularity"`
	Id                int        `json:"id" bson:"id"`
	Gender            int        `json:"gender,omitempty" bson:"gender"`
	KnowForDepartment string     `json:"known_for_department" bson:"known_for_department"`
	Name              string     `json:"name" bson:"name"`
	AlsoKnowAs        []string   `json:"also_known_as,omitempty" bson:"also_known_as"`
	ProfilepPath      string     `json:"profile_path" bson:"profile_path"`
	Biography         string     `json:"biography" bson:"biography"`
	Birthday          string     `json:"birthday,omitempty" bson:"birthday"`
	Deathday          string     `json:"deathday,omitempty" bson:"deathday"`
	Language          string     `json:"language" bson:"language"`
	SlugUrl           string     `json:"slugUrl,omitempty" bson:"slugUrl"`
	Slug              string     `json:"slug,omitempty" bson:"slug"`
	Updated           *time.Time `json:"updated,omitempty" bson:"updated"`
	UpdatedNew        string     `json:"updatedNew,omitempty" bson:"updatedNew"`
	OriginCountry     []string   `json:"origin_country" bson:"origin_country"`
	Languages         []string   `json:"languages,omitempty" bson:"languages"`
	Credits           Credits    `json:"combined_credits" bson:"credits"`
}

type ResultPerson struct {
	Page         string   `json:"page"`
	TotalResults int      `json:"total_results"`
	TotalPages   int      `json:"total_pages"`
	Results      []Person `json:"results"`
}
