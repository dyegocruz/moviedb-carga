package parameter

import (
	"context"

	"moviedb/services"

	"gopkg.in/mgo.v2/bson"
)

type Service struct {
	mongo       *services.MongoService
	getByTypeFn func(paramType string) Parameter
}

func NewService(mongo *services.MongoService) *Service {
	return &Service{mongo: mongo}
}

func (s *Service) GetByType(paramType string) Parameter {
	if s.getByTypeFn != nil {
		return s.getByTypeFn(paramType)
	}

	var parameter Parameter
	s.mongo.Collection(services.CollectionParameter).FindOne(context.TODO(), bson.M{"paramType": paramType}).Decode(&parameter)

	return parameter
}
