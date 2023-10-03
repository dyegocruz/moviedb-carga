// Define the package interacting with the database
package database

import (
	"context"
	"log"

	"moviedb/common"
	"moviedb/configs"
	"moviedb/util"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	COLLECTION_PARAMETRO     = "parametro"
	COLLECTION_MOVIE         = "movie"
	COLLECTION_PERSON        = "person"
	COLLECTION_SERIE         = "serie"
	COLLECTION_SERIE_EPISODE = "serie-episode"
)

func ConnectDB() *mongo.Client {

	ctx := context.TODO()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(configs.MongoURI()))
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Connected to MongoDB")
	return client
}

// Client instance
var DB *mongo.Client = ConnectDB()

// getting database collections
func GetCollection(client *mongo.Client, collectionName string) *mongo.Collection {
	collection := client.Database(configs.MongoDatabase()).Collection(collectionName)
	return collection
}

// Checa e cria as collections defaults da api
func CheckCreateCollections() {
	conn := DB

	names, err := DB.Database(os.Getenv("MONGO_DATABASE")).ListCollectionNames(context.TODO(), bson.M{})
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
	if !util.ArrayContainsString(names, COLLECTION_MOVIE) {
		log.Println("criar collection " + COLLECTION_MOVIE)
		conn.Database(os.Getenv("MONGO_DATABASE")).CreateCollection(context.TODO(), COLLECTION_MOVIE)
		collMovies := conn.Database(os.Getenv("MONGO_DATABASE")).Collection(COLLECTION_MOVIE)

		collMovies.Indexes().CreateMany(context.TODO(), index, opts)
	}

	// Series
	if !util.ArrayContainsString(names, COLLECTION_SERIE) {
		log.Println("criar collection " + COLLECTION_SERIE)
		conn.Database(os.Getenv("MONGO_DATABASE")).CreateCollection(context.TODO(), COLLECTION_SERIE)
		collSeries := conn.Database(os.Getenv("MONGO_DATABASE")).Collection(COLLECTION_SERIE)

		collSeries.Indexes().CreateMany(context.TODO(), index, opts)
	}

	// Series Episodes
	if !util.ArrayContainsString(names, COLLECTION_SERIE_EPISODE) {
		log.Println("criar collection " + COLLECTION_SERIE_EPISODE)
		conn.Database(os.Getenv("MONGO_DATABASE")).CreateCollection(context.TODO(), COLLECTION_SERIE_EPISODE)
		collSeries := conn.Database(os.Getenv("MONGO_DATABASE")).Collection(COLLECTION_SERIE_EPISODE)

		collSeries.Indexes().CreateMany(context.TODO(), index, opts)
	}

	// Persons
	if !util.ArrayContainsString(names, COLLECTION_PERSON) {
		log.Println("criar collection " + COLLECTION_PERSON)
		conn.Database(os.Getenv("MONGO_DATABASE")).CreateCollection(context.TODO(), COLLECTION_PERSON)
		collPerson := conn.Database(os.Getenv("MONGO_DATABASE")).Collection(COLLECTION_PERSON)

		collPerson.Indexes().CreateMany(context.TODO(), index, opts)
	}

	// Parametro
	if !util.ArrayContainsString(names, COLLECTION_PARAMETRO) {
		log.Println("criar collection " + COLLECTION_PARAMETRO)
		conn.Database(os.Getenv("MONGO_DATABASE")).CreateCollection(context.TODO(), COLLECTION_PARAMETRO)
		collParametro := conn.Database(os.Getenv("MONGO_DATABASE")).Collection(COLLECTION_PARAMETRO)

		index := []mongo.IndexModel{
			{
				Keys: bson.M{"tipo": 1},
			},
		}

		collParametro.Indexes().CreateMany(context.TODO(), index, opts)
	}

}

func GetCountAllByColletcion(collection string) int64 {
	client := DB

	count, err := client.Database(os.Getenv("MONGO_DATABASE")).Collection(collection).CountDocuments(context.TODO(), bson.M{"id": bson.M{"$gt": 0}})
	if err != nil {
		log.Println(err)
	}

	return count
}

func GenerateCatalogCheck(collection string, language string) map[int]common.CatalogCheck {
	client := DB

	filter := bson.M{"language": language}
	opts := options.Find().SetProjection(bson.M{"id": 1, "_id": 0}).SetNoCursorTimeout(true)

	cur, err := client.Database(os.Getenv("MONGO_DATABASE")).Collection(collection).Find(context.TODO(), filter, opts)
	if err != nil {
		log.Println(err)
	}

	results := make([]common.CatalogCheck, 0)
	for cur.Next(context.TODO()) {
		var result common.CatalogCheck
		err := cur.Decode(&result)
		if err != nil {
			log.Fatal(err)
		}
		results = append(results, result)
	}
	defer cur.Close(context.TODO())

	var resultCatalog = make(map[int]common.CatalogCheck, len(results))

	for _, result := range results {
		resultCatalog[result.Id] = result
	}

	return resultCatalog
}
