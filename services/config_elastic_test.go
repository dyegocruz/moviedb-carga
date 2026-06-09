package services

import (
	"testing"

	"moviedb/configs"
)

type fakeConfig struct {
	env, mongoURI, mongoDB, elasticHost, elasticUser, elasticPass string
	rabbit                                                         configs.RabbitMQConfig
}

func (f fakeConfig) Env() string                           { return f.env }
func (f fakeConfig) MongoURI() string                      { return f.mongoURI }
func (f fakeConfig) MongoDatabase() string                 { return f.mongoDB }
func (f fakeConfig) ElasticHost() string                   { return f.elasticHost }
func (f fakeConfig) ElasticUser() string                   { return f.elasticUser }
func (f fakeConfig) ElasticPassword() string               { return f.elasticPass }
func (f fakeConfig) RabbitMQ() configs.RabbitMQConfig     { return f.rabbit }

func TestDefaultConfigNotNil(t *testing.T) {
	if DefaultConfig() == nil {
		t.Fatal("expected default config")
	}
}

func TestNewElasticServiceAndClientAccessor(t *testing.T) {
	cfg := fakeConfig{
		elasticHost: "http://127.0.0.1:9200",
		elasticUser: "user",
		elasticPass: "pass",
	}

	svc := NewElasticService(cfg, "test")
	if svc == nil {
		t.Fatal("expected elastic service")
	}

	// Client may be nil if local elastic is unavailable; accessor itself must be safe.
	_ = svc.Client()
}
