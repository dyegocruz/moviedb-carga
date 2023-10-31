package configs

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func GetEnv() string {
	return os.Getenv("GO_ENV")
}

func MongoURI() string {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	return os.Getenv("MONGO_URI")
}

func MongoDatabase() string {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	return os.Getenv("MONGO_DATABASE")
}

func IsProduction() bool {
	return GetEnv() == "production"
}
