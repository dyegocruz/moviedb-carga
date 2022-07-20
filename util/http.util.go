package util

import (
	"log"
	"net/http"
)

func HttpGet(url string) *http.Response {
	response, err := http.Get(url)
	if err != nil {
		log.Println(err)
	}

	return response
}
