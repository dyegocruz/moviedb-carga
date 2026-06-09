package configs

import "testing"

func TestEnvProviderMethods(t *testing.T) {
	t.Setenv("GO_ENV", "qa")
	t.Setenv("MONGO_URI", "mongodb://mongo")
	t.Setenv("MONGO_DATABASE", "db")
	t.Setenv("ELASTICSEARCH", "http://elastic")
	t.Setenv("ELASTICSEARCH_USER", "eu")
	t.Setenv("ELASTICSEARCH_PASS", "ep")
	t.Setenv("RABBIMQ_HOST", "h")
	t.Setenv("RABBIMQ_PORT", "1")
	t.Setenv("RABBIMQ_USER", "u")
	t.Setenv("RABBIMQ_PASS", "p")

	p := NewProvider()
	if p.Env() != "qa" {
		t.Fatalf("unexpected env: %q", p.Env())
	}
	if p.MongoURI() != "mongodb://mongo" {
		t.Fatalf("unexpected mongo uri: %q", p.MongoURI())
	}
	if p.MongoDatabase() != "db" {
		t.Fatalf("unexpected mongo db: %q", p.MongoDatabase())
	}
	if p.ElasticHost() != "http://elastic" || p.ElasticUser() != "eu" || p.ElasticPassword() != "ep" {
		t.Fatal("unexpected elastic config")
	}

	r := p.RabbitMQ()
	if r.Host != "h" || r.Port != "1" || r.User != "u" || r.Password != "p" {
		t.Fatalf("unexpected rabbit config: %+v", r)
	}
}
