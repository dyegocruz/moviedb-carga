package main

import (
	"encoding/json"
	"log"

	"moviedb/common"
	"moviedb/movie"
	"moviedb/person"
	"moviedb/queue"
	"moviedb/tv"
)

func main() {
	// Initialize RabbitMQ connection
	rmq, err := queue.NewRabbitMQ()
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %s", err)
	}
	defer rmq.Close()

	// Set prefetch count
	err = rmq.SetPrefetch(10) // Each consumer will prefetch 1 message at a time
	if err != nil {
		log.Fatalf("Failed to set prefetch: %s", err)
	}

	handler := func(body []byte) error {
		var catalogProcessMessage queue.CatalogProcessMessage
		if err := json.Unmarshal(body, &catalogProcessMessage); err != nil {
			return err
		}
		log.Printf("Consumer %d received a catalogProcessMessage: %+v", 1, catalogProcessMessage)

		switch catalogProcessMessage.MediaType {
		case common.MEDIA_TYPE_MOVIE:
			go movie.PopulateMovieByIdAndLanguage(catalogProcessMessage.Id, common.LANGUAGE_PTBR, "Y")
			movie.PopulateMovieByIdAndLanguage(catalogProcessMessage.Id, common.LANGUAGE_EN, "Y")
		case common.MEDIA_TYPE_TV:
			go tv.PopulateSerieByIdAndLanguage(catalogProcessMessage.Id, common.LANGUAGE_PTBR)
			tv.PopulateSerieByIdAndLanguage(catalogProcessMessage.Id, common.LANGUAGE_EN)
		case common.MEDIA_TYPE_PERSON:
			go person.PopulatePersonByIdAndLanguage(catalogProcessMessage.Id, common.LANGUAGE_PTBR, "Y")
			person.PopulatePersonByIdAndLanguage(catalogProcessMessage.Id, common.LANGUAGE_EN, "Y")
		}
		return nil
	}

	// Consume messages
	err = rmq.ConsumeJSON(queue.QueueCatalogProcess, handler)
	if err != nil {
		log.Fatalf("Consumer %d failed to consume messages: %s", 1, err)
	}

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	select {}
}
