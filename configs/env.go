package configs

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

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
