package app

import (
	"encoding/json"
	"fmt"
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
	HandleMovieImagesPath()
}

// tvPopulator is the subset of tv.Service used in the worker handler.
type tvPopulator interface {
	PopulateSerieByIdAndLanguage(id int, language string)
	HandleTvEpisodeUpdate(episodeId int, language string)
	HandleTvLocales()
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

	if !catalogProcessMessage.MediaCatalogCheck {
		switch catalogProcessMessage.MediaType {
		case common.MEDIA_TYPE_MOVIE:
			movieSvc.PopulateMovieByIdAndLanguage(catalogProcessMessage.Id, common.LANGUAGE_PTBR, "Y")
		case common.MEDIA_TYPE_TV:
			tvSvc.PopulateSerieByIdAndLanguage(catalogProcessMessage.Id, common.LANGUAGE_PTBR)
		case common.MEDIA_TYPE_TV_EPISODE:
			tvSvc.HandleTvEpisodeUpdate(catalogProcessMessage.Id, common.LANGUAGE_PTBR)
		case common.MEDIA_TYPE_PERSON:
			personSvc.PopulatePersonByIdAndLanguage(catalogProcessMessage.Id, common.LANGUAGE_PTBR, "Y")
		default:
			log.Printf("Unknown media type: %s", catalogProcessMessage.MediaType)
		}
	} else {
		log.Printf("Catalog check for media type: %s, id: %d", catalogProcessMessage.MediaType, catalogProcessMessage.Id)
		switch catalogProcessMessage.MediaType {
		case common.MEDIA_TYPE_TV:
			tvSvc.HandleTvLocales()
		case common.MEDIA_TYPE_MOVIE:
			movieSvc.HandleMovieImagesPath()
		}
	}

	return nil
}

func RunCatalog() error {
	mongoService, err := bootstrapInitialize()
	if err != nil {
		return err
	}

	parameterService := parameter.NewService(mongoService)
	parameter := parameterService.GetByType("CHARGE_CATALOG_HANDLER")

	config := services.DefaultConfig()
	catalogService := catalogRuntimeFactory(mongoService, config)

	fmt.Printf("parameter.Options.EnableUpdateCatalogDb: %v\n", parameter.Options.EnableUpdateCatalogDb)
	fmt.Printf("parameter.Options.EnableChargeCache: %v\n", parameter.Options.EnableChargeCache)

	c := schedulerFactory()
	if err := c.AddFunc("@daily", func() {
		if parameter.Options.EnableUpdateCatalogDb {
			log.Println("[Job] General Catalog Handler")
			catalogService.GeneralCatalogHandler()
			log.Println("PROCESS COMPLETE")
		} else {
			log.Println("[Job] General Catalog Handler is disabled by configuration")
		}
	}); err != nil {
		return err
	}

	if err := c.AddFunc("0 0 3 * * *", func() {
		if parameter.Options.EnableChargeCache {
			log.Println("[Job] Elastic General Charge")
			catalogService.ElasticGeneralCharge()
			log.Println("PROCESS COMPLETE")
		} else {
			log.Println("[Job] Elastic General Charge is disabled by configuration")
		}
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
