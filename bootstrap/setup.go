package bootstrap

import (
	"moviedb/services"

	"moviedb/configs"
)

type mongoInitializer interface {
	CheckCreateCollections()
}

var initializeFn = initializeWith

func initializeWith(load func() error, newMongo func() mongoInitializer) (mongoInitializer, error) {
	if err := load(); err != nil {
		return nil, err
	}

	mongoService := newMongo()
	mongoService.CheckCreateCollections()
	return mongoService, nil
}

func Initialize() (*services.MongoService, error) {
	svc, err := initializeFn(configs.Load, func() mongoInitializer {
		return services.NewMongoService(nil)
	})
	if err != nil {
		return nil, err
	}

	if mongoService, ok := svc.(*services.MongoService); ok {
		return mongoService, nil
	}

	return nil, nil
}
