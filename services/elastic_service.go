package services

import (
	"log"
	"os"

	"github.com/olivere/elastic"
)

type ElasticService struct {
	client *elastic.Client
}

func NewElasticService(config Config, logPrefix string) *ElasticService {
	if config == nil {
		config = DefaultConfig()
	}

	client, err := elastic.NewClient(
		elastic.SetURL(config.ElasticHost()),
		elastic.SetSniff(false),
		elastic.SetBasicAuth(config.ElasticUser(), config.ElasticPassword()),
		elastic.SetErrorLog(log.New(os.Stderr, "ELASTIC ", log.LstdFlags)),
		elastic.SetInfoLog(log.New(os.Stdout, logPrefix+": ", log.LstdFlags)),
	)
	if err != nil {
		log.Println(err)
	}

	return &ElasticService{client: client}
}

func (s *ElasticService) Client() *elastic.Client {
	return s.client
}
