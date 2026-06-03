package parameter

import (
	"context"

	"moviedb/services"

	"gopkg.in/mgo.v2/bson"
)

type Service struct {
	mongo *services.MongoService
}

func NewService(mongo *services.MongoService) *Service {
	return &Service{mongo: mongo}
}

func (s *Service) GetByType(paramType string) Parameter {

	var parameter Parameter
	s.mongo.Collection(services.CollectionParameter).FindOne(context.TODO(), bson.M{"paramType": paramType}).Decode(&parameter)

	return parameter
}
