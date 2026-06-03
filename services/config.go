package services

import "moviedb/configs"

type Config interface {
	Env() string
	MongoURI() string
	MongoDatabase() string
	AWSAccessKeyID() string
	AWSSecretAccessKey() string
	ElasticHost() string
	ElasticUser() string
	ElasticPassword() string
	RabbitMQ() configs.RabbitMQConfig
}

func DefaultConfig() Config {
	return configs.NewProvider()
}
