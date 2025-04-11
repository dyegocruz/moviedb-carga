package util

import (
	"log"
	"net/http"
	"time"
)

type ErrorResponse struct {
  Success      bool   `json:"success"`
  StatusCode   int    `json:"status_code"`
  StatusMessage string `json:"status_message"`
}

func HttpGet(url string) *http.Response {
  client := &http.Client{
    Timeout: 5 * time.Second, // Define o timeout de 10 segundos
  }

	response, err := client.Get(url)
	if err != nil {
    log.Println("Error making HTTP GET request:", err)
    return &http.Response{
      StatusCode: 500,
      Body:       http.NoBody,
      Header:     http.Header{"Content-Type": []string{"application/json"}},
    }
  }

	return response
}
