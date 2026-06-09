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

var bootstrapInitialize = bootstrap.Initialize

type scheduler interface {
	AddFunc(spec string, cmd func()) error
	Start()
}

type catalogRuntime interface {
	GeneralCatalogHandler()
	ElasticGeneralCharge()
}

var schedulerFactory = func() scheduler {
	return cron.New()
}

var catalogRuntimeFactory = func(mongoService *services.MongoService, config services.Config) catalogRuntime {
	parameterService := parameter.NewService(mongoService)
	tmdbService := tmdb.NewService(parameterService)
	personService := person.NewService(mongoService, tmdbService)
	movieService := movie.NewService(mongoService, personService, tmdbService)
	tvService := tv.NewService(mongoService, personService, tmdbService)
	elasticService := services.NewElasticService(config, "catalog_search")
	return catalogCharge.NewService(config, mongoService, elasticService, movieService, personService, tvService)
}

var blockForever = func() {
	select {}
}

type rabbitConsumer interface {
	Close()
	SetPrefetch(count int) error
	ConsumeJSON(queueName string, handler func([]byte) error) error
}

var rabbitFactory = func(config services.Config) (rabbitConsumer, error) {
	return services.NewRabbitMQService(config)
}

// moviePopulator is the subset of movie.Service used in the worker handler.
type moviePopulator interface {
	PopulateMovieByIdAndLanguage(id int, language string, updateCast string)
}

// tvPopulator is the subset of tv.Service used in the worker handler.
type tvPopulator interface {
	PopulateSerieByIdAndLanguage(id int, language string)
	HandleTvEpisodeUpdate(episodeId int, language string)
}

// personPopulator is the subset of person.Service used in the worker handler.
type personPopulator interface {
	PopulatePersonByIdAndLanguage(id int, language string, updatePerson string)
}

// handleCatalogMessage dispatches a raw AMQP message body to the correct
// domain service. It is a package-level function so it can be unit-tested
// independently of the RabbitMQ transport.
func handleCatalogMessage(body []byte, movieSvc moviePopulator, tvSvc tvPopulator, personSvc personPopulator) error {
	var catalogProcessMessage queue.CatalogProcessMessage
	if err := json.Unmarshal(body, &catalogProcessMessage); err != nil {
		return err
	}

	log.Printf("Consumer %d received a catalogProcessMessage: %+v", 1, catalogProcessMessage)

	switch catalogProcessMessage.MediaType {
	case common.MEDIA_TYPE_MOVIE:
		go movieSvc.PopulateMovieByIdAndLanguage(catalogProcessMessage.Id, common.LANGUAGE_PTBR, "Y")
		movieSvc.PopulateMovieByIdAndLanguage(catalogProcessMessage.Id, common.LANGUAGE_EN, "Y")
	case common.MEDIA_TYPE_TV:
		go tvSvc.PopulateSerieByIdAndLanguage(catalogProcessMessage.Id, common.LANGUAGE_PTBR)
		tvSvc.PopulateSerieByIdAndLanguage(catalogProcessMessage.Id, common.LANGUAGE_EN)
	case common.MEDIA_TYPE_TV_EPISODE:
		go tvSvc.HandleTvEpisodeUpdate(catalogProcessMessage.Id, common.LANGUAGE_PTBR)
		tvSvc.HandleTvEpisodeUpdate(catalogProcessMessage.Id, common.LANGUAGE_EN)
	case common.MEDIA_TYPE_PERSON:
		go personSvc.PopulatePersonByIdAndLanguage(catalogProcessMessage.Id, common.LANGUAGE_PTBR, "Y")
		personSvc.PopulatePersonByIdAndLanguage(catalogProcessMessage.Id, common.LANGUAGE_EN, "Y")
	}

	return nil
}

func RunCatalog() error {
	mongoService, err := bootstrapInitialize()
	if err != nil {
		return err
	}

	config := services.DefaultConfig()
	catalogService := catalogRuntimeFactory(mongoService, config)

	c := schedulerFactory()
	if err := c.AddFunc("@daily", func() {
		log.Println("[Job] General Catalog Handler")
		catalogService.GeneralCatalogHandler()
		log.Println("PROCESS COMPLETE")
	}); err != nil {
		return err
	}

	if err := c.AddFunc("0 0 3 * * *", func() {
		log.Println("[Job] Elastic General Charge")
		catalogService.ElasticGeneralCharge()
		log.Println("PROCESS COMPLETE")
	}); err != nil {
		return err
	}

	log.Println("Start Job")
	c.Start()
	blockForever()
	return nil
}

func RunCatalogWorker() error {
	mongoService, err := bootstrapInitialize()
	if err != nil {
		return err
	}

	config := services.DefaultConfig()
	parameterService := parameter.NewService(mongoService)
	tmdbService := tmdb.NewService(parameterService)
	personService := person.NewService(mongoService, tmdbService)
	movieService := movie.NewService(mongoService, personService, tmdbService)
	tvService := tv.NewService(mongoService, personService, tmdbService)
	rabbitService, err := rabbitFactory(config)
	if err != nil {
		return err
	}
	defer rabbitService.Close()

	if err := rabbitService.SetPrefetch(10); err != nil {
		return err
	}

	return rabbitService.ConsumeJSON(queue.QueueCatalogProcess, func(body []byte) error {
		return handleCatalogMessage(body, movieService, tvService, personService)
	})
}
