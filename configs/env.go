package configs

import (
	"errors"
	"log"
	"os"
	"sync"

	"github.com/joho/godotenv"
)

type RabbitMQConfig struct {
	Host     string
	Port     string
	User     string
	Password string
}

var envLoadOnce sync.Once
var envLoadErr error

func Load() error {
	envLoadOnce.Do(func() {
		if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
			envLoadErr = err
		}
	})

	return envLoadErr
}

func getEnvString(key string) string {
	if err := Load(); err != nil {
		log.Fatal("Error loading .env file: ", err)
	}

	return os.Getenv(key)
}

func GetEnv() string {
	return getEnvString("GO_ENV")
}

func IsProduction() bool {
	return GetEnv() == "production"
}

func MongoURI() string {
	return getEnvString("MONGO_URI")
}

func MongoDatabase() string {
	return getEnvString("MONGO_DATABASE")
}

func GetAcessKeyId() string {
	return getEnvString("AWS_ACCESS_KEY_ID")
}

func GetSecretAccessKey() string {
	return getEnvString("AWS_SECRET_ACCESS_KEY")
}

func GetElkHost() string {
	return getEnvString("ELASTICSEARCH")
}

func GetElkUser() string {
	return getEnvString("ELASTICSEARCH_USER")
}

func GetELKPassword() string {
	return getEnvString("ELASTICSEARCH_PASS")
}

func GetRabbitMQEnv() RabbitMQConfig {
	return RabbitMQConfig{
		Host:     getEnvString("RABBIMQ_HOST"),
		Port:     getEnvString("RABBIMQ_PORT"),
		User:     getEnvString("RABBIMQ_USER"),
		Password: getEnvString("RABBIMQ_PASS"),
	}
}
