package configs

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type RabbitMQConfig struct {
  Host     string
  Port     string
  User     string
  Password string
}

func getEnvString(key string) string {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
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

func GetQueueUrl() string {
	return getEnvString("AWS_QUEUE_URL_BASE") + "/" + getEnvString("AWS_ACCOUNT_ID") + "/" + getEnvString("AWS_QUEUE_NAME")
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

