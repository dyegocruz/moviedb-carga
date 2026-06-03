package bootstrap

import (
	"moviedb/services"

	"moviedb/configs"
)

func Initialize() (*services.MongoService, error) {
	if err := configs.Load(); err != nil {
		return nil, err
	}

	mongoService := services.NewMongoService(nil)
	mongoService.CheckCreateCollections()
	return mongoService, nil
}
