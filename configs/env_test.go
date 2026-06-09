package configs

import "testing"

func TestGettersFromEnvironment(t *testing.T) {
	t.Setenv("GO_ENV", "test")
	t.Setenv("MONGO_URI", "mongodb://localhost:27017")
	t.Setenv("MONGO_DATABASE", "moviedb_test")
	t.Setenv("ELASTICSEARCH", "http://localhost:9200")
	t.Setenv("ELASTICSEARCH_USER", "u")
	t.Setenv("ELASTICSEARCH_PASS", "p")
	t.Setenv("RABBIMQ_HOST", "rabbit")
	t.Setenv("RABBIMQ_PORT", "5672")
	t.Setenv("RABBIMQ_USER", "admin")
	t.Setenv("RABBIMQ_PASS", "admin123")

	if GetEnv() != "test" {
		t.Fatalf("expected GO_ENV=test, got %q", GetEnv())
	}
	if MongoURI() != "mongodb://localhost:27017" {
		t.Fatalf("unexpected mongo uri: %q", MongoURI())
	}
	if MongoDatabase() != "moviedb_test" {
		t.Fatalf("unexpected mongo db: %q", MongoDatabase())
	}
	if GetElkHost() != "http://localhost:9200" {
		t.Fatalf("unexpected elastic host: %q", GetElkHost())
	}
	if GetElkUser() != "u" {
		t.Fatalf("unexpected elastic user: %q", GetElkUser())
	}
	if GetELKPassword() != "p" {
		t.Fatalf("unexpected elastic pass: %q", GetELKPassword())
	}

	rmq := GetRabbitMQEnv()
	if rmq.Host != "rabbit" || rmq.Port != "5672" || rmq.User != "admin" || rmq.Password != "admin123" {
		t.Fatalf("unexpected rabbit config: %+v", rmq)
	}
}

func TestIsProduction(t *testing.T) {
	t.Setenv("GO_ENV", "production")
	if !IsProduction() {
		t.Fatal("expected IsProduction to be true")
	}

	t.Setenv("GO_ENV", "dev")
	if IsProduction() {
		t.Fatal("expected IsProduction to be false")
	}
}
