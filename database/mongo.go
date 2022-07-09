// Define the package interacting with the database
package database

import (
	"context"
	"log"
	"moviedb/util"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	PARAMETRO = "parametro"
	MOVIE     = "movie"
	PERSON    = "person"
	SERIE     = "serie"
)

const (
	// Timeout operations after N seconds
	connectTimeout           = 5
	connectionStringTemplate = "mongodb://%s:%s@%s"
)

// GetConnection - Retrieves a client to the DocumentDB
func GetConnection() (*mongo.Client, context.Context, context.CancelFunc) {

	var connectionURI = os.Getenv("MONGO_URI")

	client, err := mongo.NewClient(options.Client().ApplyURI(connectionURI))
	if err != nil {
		log.Printf("Failed to create client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.TODO(), connectTimeout*time.Second)

	err = client.Connect(ctx)
	if err != nil {
		log.Printf("Failed to connect to cluster: %v", err)
	}

	// Force a connection to verify our connection string
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Printf("Failed to ping cluster: %v", err)
	}

	// fmt.Println("Connected to MongoDB!")
	return client, ctx, cancel
}

// Checa e cria as collections defaults da api
func CheckCreateCollections() {
	conn, ctx, cancel := GetConnection()
	defer cancel()
	defer conn.Disconnect(ctx)

	names, err := conn.Database(os.Getenv("MONGO_DATABASE")).ListCollectionNames(context.TODO(), bson.M{})
	if err != nil {
		// Handle error
		log.Printf("Failed to get coll names: %v", err)
		return
	}

	index := []mongo.IndexModel{
		{
			Keys: bson.M{"id": 1},
		},
		{
			Keys: bson.M{"language": 1},
		},
	}
	opts := options.CreateIndexes().SetMaxTime(10 * time.Second)

	// Movies
	if !util.ArrayContainsString(names, MOVIE) {
		log.Println("criar collection " + MOVIE)
		conn.Database(os.Getenv("MONGO_DATABASE")).CreateCollection(context.TODO(), MOVIE)
		collMovies := conn.Database(os.Getenv("MONGO_DATABASE")).Collection(MOVIE)

		collMovies.Indexes().CreateMany(context.TODO(), index, opts)
	}

	// Series
	if !util.ArrayContainsString(names, SERIE) {
		log.Println("criar collection " + SERIE)
		conn.Database(os.Getenv("MONGO_DATABASE")).CreateCollection(context.TODO(), SERIE)
		collSeries := conn.Database(os.Getenv("MONGO_DATABASE")).Collection(SERIE)

		collSeries.Indexes().CreateMany(context.TODO(), index, opts)
	}

	// Persons
	if !util.ArrayContainsString(names, PERSON) {
		log.Println("criar collection " + PERSON)
		conn.Database(os.Getenv("MONGO_DATABASE")).CreateCollection(context.TODO(), PERSON)
		collPerson := conn.Database(os.Getenv("MONGO_DATABASE")).Collection(PERSON)

		collPerson.Indexes().CreateMany(context.TODO(), index, opts)
	}

	// Parametro
	if !util.ArrayContainsString(names, PARAMETRO) {
		log.Println("criar collection " + PARAMETRO)
		conn.Database(os.Getenv("MONGO_DATABASE")).CreateCollection(context.TODO(), PARAMETRO)
		collParametro := conn.Database(os.Getenv("MONGO_DATABASE")).Collection(PARAMETRO)

		index := []mongo.IndexModel{
			{
				Keys: bson.M{"tipo": 1},
			},
		}

		collParametro.Indexes().CreateMany(context.TODO(), index, opts)
	}

}
