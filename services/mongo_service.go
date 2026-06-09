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

var mongoConnect = mongo.Connect
var mongoPing = func(client *mongo.Client, ctx context.Context) error {
	return client.Ping(ctx, nil)
}

const (
	CollectionParameter     = "parameter"
	CollectionMovie         = "movie"
	CollectionPerson        = "person"
	CollectionSerie         = "serie"
	CollectionSerieEpisode  = "serie-episode"
)

type MongoService struct {
	client          *mongo.Client
	dbName          string
	config          Config
	collectionFn    func(collectionName string) *mongo.Collection
	collectionOpsFn func(collectionName string) mongoCollectionOps
}

type mongoCursor interface {
	Next(ctx context.Context) bool
	Decode(val interface{}) error
	Close(ctx context.Context) error
}

type mongoCollectionOps interface {
	CountDocuments(ctx context.Context, filter interface{}, opts ...*options.CountOptions) (int64, error)
	Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (mongoCursor, error)
}

type mongoCursorAdapter struct {
	cursor *mongo.Cursor
}

func (a mongoCursorAdapter) Next(ctx context.Context) bool {
	return a.cursor.Next(ctx)
}

func (a mongoCursorAdapter) Decode(val interface{}) error {
	return a.cursor.Decode(val)
}

func (a mongoCursorAdapter) Close(ctx context.Context) error {
	return a.cursor.Close(ctx)
}

type mongoCollectionAdapter struct {
	collection *mongo.Collection
}

func (a mongoCollectionAdapter) CountDocuments(ctx context.Context, filter interface{}, opts ...*options.CountOptions) (int64, error) {
	return a.collection.CountDocuments(ctx, filter, opts...)
}

func (a mongoCollectionAdapter) Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (mongoCursor, error) {
	cur, err := a.collection.Find(ctx, filter, opts...)
	if err != nil {
		return nil, err
	}

	return mongoCursorAdapter{cursor: cur}, nil
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
	client, err := mongoConnect(ctx, options.Client().ApplyURI(config.MongoURI()))
	if err != nil {
		log.Fatal(err)
	}

	if err := mongoPing(client, ctx); err != nil {
		log.Fatal(err)
	}

	log.Println("Connected to MongoDB")
	return &MongoService{client: client, dbName: config.MongoDatabase(), config: config}
}

func (s *MongoService) Collection(collectionName string) *mongo.Collection {
	if s.collectionFn != nil {
		return s.collectionFn(collectionName)
	}

	return s.client.Database(s.dbName).Collection(collectionName)
}

func (s *MongoService) collectionOps(collectionName string) mongoCollectionOps {
	if s.collectionOpsFn != nil {
		return s.collectionOpsFn(collectionName)
	}

	return mongoCollectionAdapter{collection: s.Collection(collectionName)}
}

func (s *MongoService) CheckCreateCollections() {
	names, err := s.client.Database(s.dbName).ListCollectionNames(context.TODO(), bson.M{})
	if err != nil {
		log.Printf("Failed to get coll names: %v", err)
		return
	}

	_ = applyCollectionInitialization(
		names,
		func(collectionName string) error {
			return s.client.Database(s.dbName).CreateCollection(context.TODO(), collectionName)
		},
		func(collectionName string, indexes []mongo.IndexModel, opts *options.CreateIndexesOptions) {
			s.client.Database(s.dbName).Collection(collectionName).Indexes().CreateMany(context.TODO(), indexes, opts)
		},
	)
}

func applyCollectionInitialization(
	existingNames []string,
	createCollection func(collectionName string) error,
	createIndexes func(collectionName string, indexes []mongo.IndexModel, opts *options.CreateIndexesOptions),
) []string {
	index := []mongo.IndexModel{
		{Keys: bson.M{"id": 1}},
		{Keys: bson.M{"language": 1}},
	}
	parameterIndex := []mongo.IndexModel{{Keys: bson.M{"tipo": 1}}}
	opts := options.CreateIndexes().SetMaxTime(10 * time.Second)

	collections := []string{CollectionMovie, CollectionSerie, CollectionSerieEpisode, CollectionPerson, CollectionParameter}
	created := make([]string, 0)
	for _, collectionName := range collections {
		if util.ArrayContainsString(existingNames, collectionName) {
			continue
		}

		log.Println("create collection " + collectionName)
		if err := createCollection(collectionName); err != nil {
			log.Println(err)
		}

		if collectionName == CollectionParameter {
			createIndexes(collectionName, parameterIndex, opts)
			created = append(created, collectionName)
			continue
		}

		createIndexes(collectionName, index, opts)
		created = append(created, collectionName)
	}

	return created
}

func (s *MongoService) GetCountAllByCollection(collection string) int64 {
	count, err := s.collectionOps(collection).CountDocuments(context.TODO(), bson.M{"_id": bson.M{"$ne": ""}})
	if err != nil {
		log.Println(err)
	}

	return count
}

func (s *MongoService) GetCountAllByCollectionAndLanguage(collection string, language string) int64 {
	count, err := s.collectionOps(collection).CountDocuments(context.TODO(), bson.M{"language": language})
	if err != nil {
		log.Println(err)
	}

	return count
}

func (s *MongoService) GetAllIdsByLanguage(collection string, language string) []int {
	filter := bson.M{"language": language}
	opts := options.Find().SetProjection(bson.M{"id": 1, "_id": 0}).SetNoCursorTimeout(true)

	cur, err := s.collectionOps(collection).Find(context.TODO(), filter, opts)
	if err != nil {
		log.Println(err)
		return []int{}
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

	cur, err := s.collectionOps(collection).Find(ctx, filter, opts)
	if err != nil {
		log.Println(err)
		return map[int]common.CatalogCheck{}
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
