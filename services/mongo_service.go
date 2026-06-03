package services

import (
	"context"
	"log"
	"time"

	"moviedb/common"
	"moviedb/util"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	CollectionParameter     = "parameter"
	CollectionMovie         = "movie"
	CollectionPerson        = "person"
	CollectionSerie         = "serie"
	CollectionSerieEpisode  = "serie-episode"
)

type MongoService struct {
	client *mongo.Client
	dbName string
	config Config
}

func (s *MongoService) Close() error {
	if s == nil || s.client == nil {
		return nil
	}

	return s.client.Disconnect(context.TODO())
}

func NewMongoService(config Config) *MongoService {
	if config == nil {
		config = DefaultConfig()
	}

	ctx := context.TODO()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(config.MongoURI()))
	if err != nil {
		log.Fatal(err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		log.Fatal(err)
	}

	log.Println("Connected to MongoDB")
	return &MongoService{client: client, dbName: config.MongoDatabase(), config: config}
}

func (s *MongoService) Collection(collectionName string) *mongo.Collection {
	return s.client.Database(s.dbName).Collection(collectionName)
}

func (s *MongoService) CheckCreateCollections() {
	names, err := s.client.Database(s.dbName).ListCollectionNames(context.TODO(), bson.M{})
	if err != nil {
		log.Printf("Failed to get coll names: %v", err)
		return
	}

	index := []mongo.IndexModel{
		{Keys: bson.M{"id": 1}},
		{Keys: bson.M{"language": 1}},
	}
	opts := options.CreateIndexes().SetMaxTime(10 * time.Second)

	collections := []string{CollectionMovie, CollectionSerie, CollectionSerieEpisode, CollectionPerson, CollectionParameter}
	for _, collectionName := range collections {
		if util.ArrayContainsString(names, collectionName) {
			continue
		}

		log.Println("create collection " + collectionName)
		if err := s.client.Database(s.dbName).CreateCollection(context.TODO(), collectionName); err != nil {
			log.Println(err)
		}

		collection := s.client.Database(s.dbName).Collection(collectionName)
		if collectionName == CollectionParameter {
			parameterIndex := []mongo.IndexModel{{Keys: bson.M{"tipo": 1}}}
			collection.Indexes().CreateMany(context.TODO(), parameterIndex, opts)
			continue
		}

		collection.Indexes().CreateMany(context.TODO(), index, opts)
	}
}

func (s *MongoService) GetCountAllByCollection(collection string) int64 {
	count, err := s.Collection(collection).CountDocuments(context.TODO(), bson.M{"_id": bson.M{"$ne": ""}})
	if err != nil {
		log.Println(err)
	}

	return count
}

func (s *MongoService) GetCountAllByCollectionAndLanguage(collection string, language string) int64 {
	count, err := s.Collection(collection).CountDocuments(context.TODO(), bson.M{"language": language})
	if err != nil {
		log.Println(err)
	}

	return count
}

func (s *MongoService) GetAllIdsByLanguage(collection string, language string) []int {
	filter := bson.M{"language": language}
	opts := options.Find().SetProjection(bson.M{"id": 1, "_id": 0}).SetNoCursorTimeout(true)

	cur, err := s.Collection(collection).Find(context.TODO(), filter, opts)
	if err != nil {
		log.Println(err)
	}

	results := make([]int, 0)
	for cur.Next(context.TODO()) {
		var result common.CatalogCheck
		if err := cur.Decode(&result); err != nil {
			log.Fatal(err)
		}
		results = append(results, result.Id)
	}
	defer cur.Close(context.TODO())

	return results
}

func (s *MongoService) GenerateCatalogCheck(collection string, language string) map[int]common.CatalogCheck {
	ctx := context.TODO()
	filter := bson.M{"language": language}
	opts := options.Find().SetProjection(bson.M{"id": 1, "_id": 0})

	log.Print("STARTING Generate Catalog check for ", collection)

	cur, err := s.Collection(collection).Find(ctx, filter, opts)
	if err != nil {
		log.Println(err)
	}

	resultCatalog := make(map[int]common.CatalogCheck)
	for cur.Next(ctx) {
		var result common.CatalogCheck
		if err := cur.Decode(&result); err != nil {
			log.Fatal(err)
		}
		resultCatalog[result.Id] = result
	}

	log.Printf("Generate Catalog check for %s completed", collection)
	return resultCatalog
}
