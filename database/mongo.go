// Define the package interacting with the database
package database

import (
	"context"
	"log"

	"moviedb/common"
	"moviedb/configs"
	"moviedb/util"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	COLLECTION_PARAMETER     = "parameter"
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

// check and create the default collections
func CheckCreateCollections() {
	conn := DB

	names, err := DB.Database(configs.MongoDatabase()).ListCollectionNames(context.TODO(), bson.M{})
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
		log.Println("create collection " + COLLECTION_MOVIE)
		conn.Database(configs.MongoDatabase()).CreateCollection(context.TODO(), COLLECTION_MOVIE)
		collMovies := conn.Database(configs.MongoDatabase()).Collection(COLLECTION_MOVIE)

		collMovies.Indexes().CreateMany(context.TODO(), index, opts)
	}

	// Series
	if !util.ArrayContainsString(names, COLLECTION_SERIE) {
		log.Println("create collection " + COLLECTION_SERIE)
		conn.Database(configs.MongoDatabase()).CreateCollection(context.TODO(), COLLECTION_SERIE)
		collSeries := conn.Database(configs.MongoDatabase()).Collection(COLLECTION_SERIE)

		collSeries.Indexes().CreateMany(context.TODO(), index, opts)
	}

	// Series Episodes
	if !util.ArrayContainsString(names, COLLECTION_SERIE_EPISODE) {
		log.Println("create collection " + COLLECTION_SERIE_EPISODE)
		conn.Database(configs.MongoDatabase()).CreateCollection(context.TODO(), COLLECTION_SERIE_EPISODE)
		collSeries := conn.Database(configs.MongoDatabase()).Collection(COLLECTION_SERIE_EPISODE)

		collSeries.Indexes().CreateMany(context.TODO(), index, opts)
	}

	// Persons
	if !util.ArrayContainsString(names, COLLECTION_PERSON) {
		log.Println("create collection " + COLLECTION_PERSON)
		conn.Database(configs.MongoDatabase()).CreateCollection(context.TODO(), COLLECTION_PERSON)
		collPerson := conn.Database(configs.MongoDatabase()).Collection(COLLECTION_PERSON)

		collPerson.Indexes().CreateMany(context.TODO(), index, opts)
	}

	// Parameter
	if !util.ArrayContainsString(names, COLLECTION_PARAMETER) {
		log.Println("create collection " + COLLECTION_PARAMETER)
		conn.Database(configs.MongoDatabase()).CreateCollection(context.TODO(), COLLECTION_PARAMETER)
		collParametro := conn.Database(configs.MongoDatabase()).Collection(COLLECTION_PARAMETER)

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

	// count, err := client.Database(configs.MongoDatabase()).Collection(collection).CountDocuments(context.TODO(), bson.M{"id": bson.M{"$gt": 0}})
	count, err := client.Database(configs.MongoDatabase()).Collection(collection).CountDocuments(context.TODO(), bson.M{"_id": bson.M{"$ne": ""}})
	if err != nil {
		log.Println(err)
	}

	return count
}

func GetCountAllByColletcionAndLanguage(collection string, language string) int64 {
	client := DB

	count, err := client.Database(configs.MongoDatabase()).Collection(collection).CountDocuments(context.TODO(), bson.M{"language": language})
	if err != nil {
		log.Println(err)
	}

	return count
}

func GetAllIdsByLanguage(collection string, language string) []int {
	client := DB

	filter := bson.M{"language": language}
	opts := options.Find().SetProjection(bson.M{"id": 1, "_id": 0}).SetNoCursorTimeout(true)

	cur, err := client.Database(configs.MongoDatabase()).Collection(collection).Find(context.TODO(), filter, opts)
	if err != nil {
		log.Println(err)
	}

	results := make([]int, 0)
	for cur.Next(context.TODO()) {
		var result common.CatalogCheck
		err := cur.Decode(&result)
		if err != nil {
			log.Fatal(err)
		}
		results = append(results, result.Id)
	}
	defer cur.Close(context.TODO())

	return results
}

func GenerateCatalogCheck(collection string, language string) map[int]common.CatalogCheck {
	client := DB

	filter := bson.M{"language": language}
	opts := options.Find().SetProjection(bson.M{"id": 1, "_id": 0}).SetSort(bson.D{{Key: "id", Value: -1}}).SetNoCursorTimeout(true)

	log.Print("STARTING Generate Catalog check for ", collection)

	cur, err := client.Database(configs.MongoDatabase()).Collection(collection).Find(context.TODO(), filter, opts)
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
  log.Println("====>results", len(results), results[0].Id)
	var resultCatalog = make(map[int]common.CatalogCheck, len(results))
	for _, result := range results {
    resultCatalog[result.Id] = result	
	}

	log.Printf("Generate Catalog check for %s completed", collection)
	return resultCatalog
}
