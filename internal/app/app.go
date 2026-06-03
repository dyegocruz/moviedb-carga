package app

import (
	"encoding/json"
	"log"

	"moviedb/bootstrap"
	catalogCharge "moviedb/catalog-charge"
	"moviedb/common"
	"moviedb/movie"
	"moviedb/parameter"
	"moviedb/person"
	"moviedb/queue"
	"moviedb/services"
	"moviedb/tmdb"
	"moviedb/tv"

	"github.com/robfig/cron"
)

func RunCatalog() error {
	mongoService, err := bootstrap.Initialize()
	if err != nil {
		return err
	}

	config := services.DefaultConfig()
	parameterService := parameter.NewService(mongoService)
	tmdbService := tmdb.NewService(parameterService)
	personService := person.NewService(mongoService, tmdbService)
	movieService := movie.NewService(mongoService, personService, tmdbService)
	tvService := tv.NewService(mongoService, personService, tmdbService)
	elasticService := services.NewElasticService(config, "catalog_search")
	catalogService := catalogCharge.NewService(config, mongoService, elasticService, movieService, personService, tvService)

	c := cron.New()
	c.AddFunc("@daily", func() {
		log.Println("[Job] General Catalog Handler")
		catalogService.GeneralCatalogHandler()
		log.Println("PROCESS COMPLETE")
	})

	c.AddFunc("0 0 3 * * *", func() {
		log.Println("[Job] Elastic General Charge")
		catalogService.ElasticGeneralCharge()
		log.Println("PROCESS COMPLETE")
	})

	log.Println("Start Job")
	c.Start()
	select {}
}

func RunCatalogWorker() error {
	mongoService, err := bootstrap.Initialize()
	if err != nil {
		return err
	}

	config := services.DefaultConfig()
	parameterService := parameter.NewService(mongoService)
	tmdbService := tmdb.NewService(parameterService)
	personService := person.NewService(mongoService, tmdbService)
	movieService := movie.NewService(mongoService, personService, tmdbService)
	tvService := tv.NewService(mongoService, personService, tmdbService)
	rabbitService, err := services.NewRabbitMQService(config)
	if err != nil {
		return err
	}
	defer rabbitService.Close()

	if err := rabbitService.SetPrefetch(10); err != nil {
		return err
	}

	handler := func(body []byte) error {
		var catalogProcessMessage queue.CatalogProcessMessage
		if err := json.Unmarshal(body, &catalogProcessMessage); err != nil {
			return err
		}

		log.Printf("Consumer %d received a catalogProcessMessage: %+v", 1, catalogProcessMessage)

		switch catalogProcessMessage.MediaType {
		case common.MEDIA_TYPE_MOVIE:
			go movieService.PopulateMovieByIdAndLanguage(catalogProcessMessage.Id, common.LANGUAGE_PTBR, "Y")
			movieService.PopulateMovieByIdAndLanguage(catalogProcessMessage.Id, common.LANGUAGE_EN, "Y")
		case common.MEDIA_TYPE_TV:
			go tvService.PopulateSerieByIdAndLanguage(catalogProcessMessage.Id, common.LANGUAGE_PTBR)
			tvService.PopulateSerieByIdAndLanguage(catalogProcessMessage.Id, common.LANGUAGE_EN)
		case common.MEDIA_TYPE_TV_EPISODE:
			go tvService.HandleTvEpisodeUpdate(catalogProcessMessage.Id, common.LANGUAGE_PTBR)
			tvService.HandleTvEpisodeUpdate(catalogProcessMessage.Id, common.LANGUAGE_EN)
		case common.MEDIA_TYPE_PERSON:
			go personService.PopulatePersonByIdAndLanguage(catalogProcessMessage.Id, common.LANGUAGE_PTBR, "Y")
			personService.PopulatePersonByIdAndLanguage(catalogProcessMessage.Id, common.LANGUAGE_EN, "Y")
		}

		return nil
	}

	return rabbitService.ConsumeJSON(queue.QueueCatalogProcess, handler)
}
