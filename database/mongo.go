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
	"go.mongodb.org/mongo-driver/x/bsonx"
)

func MongoConnect() (*mongo.Database, error) {

	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017/")
	client, err := mongo.Connect(context.TODO(), clientOptions)

	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(context.TODO(), nil)

	if err != nil {
		log.Fatal(err)
	}
	// return client.Database(os.Getenv("MONGO_DB_NAME")), err
	return client.Database("moviedb-dev"), err
}

func GetMongoCollection(collection string) *mongo.Collection {
	client, err := MongoConnect()

	// fmt.Println("Connected to MongoDB!")

	if err != nil {
		log.Fatal(err)
	}

	return client.Collection(collection)
}

const (
	// Timeout operations after N seconds
	connectTimeout           = 5
	connectionStringTemplate = "mongodb://%s:%s@%s"
)

// GetConnection - Retrieves a client to the DocumentDB
func GetConnection() (*mongo.Client, context.Context, context.CancelFunc) {
	// username := os.Getenv("MONGODB_USERNAME")
	// password := os.Getenv("MONGODB_PASSWORD")
	// clusterEndpoint := os.Getenv("MONGODB_ENDPOINT")

	// connectionURI := fmt.Sprintf(connectionStringTemplate, username, password, clusterEndpoint)
	// clientOptions := options.Client().ApplyURI("mongodb://localhost:27017/?connect=direct")
	// client, err := mongo.Connect(context.TODO(), clientOptions)

	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://localhost:27017"))
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
			Keys: bsonx.Doc{{Key: "id"}},
		},
		{
			Keys: bsonx.Doc{{Key: "language"}},
		},
	}
	opts := options.CreateIndexes().SetMaxTime(10 * time.Second)

	// Movies
	if !util.ArrayContainsString(names, "movie") {
		log.Println("criar collection movie")
		conn.Database(os.Getenv("MONGO_DATABASE")).CreateCollection(context.TODO(), "movie")
		collMovies := conn.Database(os.Getenv("MONGO_DATABASE")).Collection("movie")

		collMovies.Indexes().CreateMany(context.TODO(), index, opts)
	}

	// Series
	if !util.ArrayContainsString(names, "serie") {
		log.Println("criar collection serie")
		conn.Database(os.Getenv("MONGO_DATABASE")).CreateCollection(context.TODO(), "serie")
		collSeries := conn.Database(os.Getenv("MONGO_DATABASE")).Collection("serie")

		collSeries.Indexes().CreateMany(context.TODO(), index, opts)
	}

	// Persons
	if !util.ArrayContainsString(names, "person") {
		log.Println("criar collection person")
		conn.Database(os.Getenv("MONGO_DATABASE")).CreateCollection(context.TODO(), "person")
		collPerson := conn.Database(os.Getenv("MONGO_DATABASE")).Collection("person")

		collPerson.Indexes().CreateMany(context.TODO(), index, opts)
	}

}
