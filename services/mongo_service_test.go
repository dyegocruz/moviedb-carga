package services

import (
	"context"
	"reflect"
	"testing"

	"moviedb/common"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type createIndexesCall struct {
	collectionName string
	indexesLen     int
	firstIndexKey  string
	hasMaxTime     bool
}

type fakeMongoCursor struct {
	items []common.CatalogCheck
	idx   int
	closed bool
}

func (f *fakeMongoCursor) Next(ctx context.Context) bool {
	return f.idx < len(f.items)
}

func (f *fakeMongoCursor) Decode(val interface{}) error {
	v, ok := val.(*common.CatalogCheck)
	if !ok {
		return nil
	}
	*v = f.items[f.idx]
	f.idx++
	return nil
}

func (f *fakeMongoCursor) Close(ctx context.Context) error {
	f.closed = true
	return nil
}

type fakeMongoCollection struct {
	countResult int64
	countErr    error
	findCursor  mongoCursor
	findErr     error
	countFilter interface{}
	findFilter  interface{}
}

func (f *fakeMongoCollection) CountDocuments(ctx context.Context, filter interface{}, opts ...*options.CountOptions) (int64, error) {
	f.countFilter = filter
	return f.countResult, f.countErr
}

func (f *fakeMongoCollection) Find(ctx context.Context, filter interface{}, opts ...*options.FindOptions) (mongoCursor, error) {
	f.findFilter = filter
	return f.findCursor, f.findErr
}

func TestMongoService_Close_NilSafe(t *testing.T) {
	var svc *MongoService
	// Should not panic when called on a nil receiver.
	if err := svc.Close(); err != nil {
		t.Fatalf("unexpected error closing nil MongoService: %v", err)
	}
}

func TestMongoService_Close_NilClient(t *testing.T) {
	svc := &MongoService{client: nil}
	if err := svc.Close(); err != nil {
		t.Fatalf("unexpected error closing MongoService with nil client: %v", err)
	}
}

func TestMongoService_Collection_UsesHook(t *testing.T) {
	svc := &MongoService{}
	called := false
	svc.collectionFn = func(collectionName string) *mongo.Collection {
		called = true
		if collectionName != CollectionMovie {
			t.Fatalf("unexpected collection name: %s", collectionName)
		}
		return nil
	}

	if got := svc.Collection(CollectionMovie); got != nil {
		t.Fatal("expected nil collection from hook")
	}
	if !called {
		t.Fatal("expected collection hook to be called")
	}
}

func TestMongoService_GetCountAllByCollection(t *testing.T) {
	fakeColl := &fakeMongoCollection{countResult: 7}
	svc := &MongoService{collectionOpsFn: func(collectionName string) mongoCollectionOps {
		if collectionName != CollectionMovie {
			t.Fatalf("unexpected collection: %s", collectionName)
		}
		return fakeColl
	}}

	count := svc.GetCountAllByCollection(CollectionMovie)
	if count != 7 {
		t.Fatalf("expected count 7, got %d", count)
	}

	expectedFilter := bson.M{"_id": bson.M{"$ne": ""}}
	if !reflect.DeepEqual(fakeColl.countFilter, expectedFilter) {
		t.Fatalf("unexpected filter: %#v", fakeColl.countFilter)
	}
}

func TestMongoService_GetCountAllByCollectionAndLanguage(t *testing.T) {
	fakeColl := &fakeMongoCollection{countResult: 3}
	svc := &MongoService{collectionOpsFn: func(collectionName string) mongoCollectionOps {
		return fakeColl
	}}

	count := svc.GetCountAllByCollectionAndLanguage(CollectionSerie, "en")
	if count != 3 {
		t.Fatalf("expected count 3, got %d", count)
	}

	expectedFilter := bson.M{"language": "en"}
	if !reflect.DeepEqual(fakeColl.countFilter, expectedFilter) {
		t.Fatalf("unexpected filter: %#v", fakeColl.countFilter)
	}
}

func TestMongoService_GetAllIdsByLanguage(t *testing.T) {
	cursor := &fakeMongoCursor{items: []common.CatalogCheck{{Id: 11}, {Id: 22}}}
	fakeColl := &fakeMongoCollection{findCursor: cursor}
	svc := &MongoService{collectionOpsFn: func(collectionName string) mongoCollectionOps {
		return fakeColl
	}}

	ids := svc.GetAllIdsByLanguage(CollectionMovie, "pt-BR")
	if !reflect.DeepEqual(ids, []int{11, 22}) {
		t.Fatalf("unexpected ids: %#v", ids)
	}

	expectedFilter := bson.M{"language": "pt-BR"}
	if !reflect.DeepEqual(fakeColl.findFilter, expectedFilter) {
		t.Fatalf("unexpected filter: %#v", fakeColl.findFilter)
	}
	if !cursor.closed {
		t.Fatal("expected cursor to be closed")
	}
}

func TestMongoService_GenerateCatalogCheck(t *testing.T) {
	cursor := &fakeMongoCursor{items: []common.CatalogCheck{{Id: 1}, {Id: 2}}}
	fakeColl := &fakeMongoCollection{findCursor: cursor}
	svc := &MongoService{collectionOpsFn: func(collectionName string) mongoCollectionOps {
		return fakeColl
	}}

	got := svc.GenerateCatalogCheck(CollectionPerson, "en")
	if len(got) != 2 {
		t.Fatalf("expected 2 items, got %d", len(got))
	}
	if got[1].Id != 1 || got[2].Id != 2 {
		t.Fatalf("unexpected map content: %#v", got)
	}

	expectedFilter := bson.M{"language": "en"}
	if !reflect.DeepEqual(fakeColl.findFilter, expectedFilter) {
		t.Fatalf("unexpected filter: %#v", fakeColl.findFilter)
	}
}

func TestMongoService_GetAllIdsByLanguage_ReturnsEmptyOnFindError(t *testing.T) {
	fakeColl := &fakeMongoCollection{findErr: context.DeadlineExceeded}
	svc := &MongoService{collectionOpsFn: func(collectionName string) mongoCollectionOps {
		return fakeColl
	}}

	ids := svc.GetAllIdsByLanguage(CollectionMovie, "en")
	if len(ids) != 0 {
		t.Fatalf("expected empty ids, got %#v", ids)
	}
}

func TestMongoService_GenerateCatalogCheck_ReturnsEmptyOnFindError(t *testing.T) {
	fakeColl := &fakeMongoCollection{findErr: context.DeadlineExceeded}
	svc := &MongoService{collectionOpsFn: func(collectionName string) mongoCollectionOps {
		return fakeColl
	}}

	got := svc.GenerateCatalogCheck(CollectionMovie, "en")
	if len(got) != 0 {
		t.Fatalf("expected empty map, got %#v", got)
	}
}

func TestNewMongoService_UsesConnectAndPingHooks(t *testing.T) {
	origConnect := mongoConnect
	origPing := mongoPing
	defer func() {
		mongoConnect = origConnect
		mongoPing = origPing
	}()

	mongoConnect = func(ctx context.Context, opts ...*options.ClientOptions) (*mongo.Client, error) {
		if len(opts) == 0 {
			t.Fatal("expected client options")
		}
		if opts[0].GetURI() != "mongodb://fake" {
			t.Fatalf("unexpected uri: %s", opts[0].GetURI())
		}
		return &mongo.Client{}, nil
	}
	mongoPing = func(client *mongo.Client, ctx context.Context) error {
		if client == nil {
			t.Fatal("expected non-nil client")
		}
		return nil
	}

	cfg := fakeConfig{mongoURI: "mongodb://fake", mongoDB: "db_test"}
	svc := NewMongoService(cfg)
	if svc == nil {
		t.Fatal("expected non-nil mongo service")
	}
	if svc.dbName != "db_test" {
		t.Fatalf("unexpected db name: %s", svc.dbName)
	}
}

func TestApplyCollectionInitialization_CreatesMissingAndParameterIndex(t *testing.T) {
	createdCalls := make([]string, 0)
	indexCalls := make([]createIndexesCall, 0)

	created := applyCollectionInitialization(
		[]string{CollectionMovie},
		func(collectionName string) error {
			createdCalls = append(createdCalls, collectionName)
			return nil
		},
		func(collectionName string, indexes []mongo.IndexModel, opts *options.CreateIndexesOptions) {
			firstKey := ""
			if len(indexes) > 0 {
				if m, ok := indexes[0].Keys.(bson.M); ok {
					for k := range m {
						firstKey = k
						break
					}
				}
			}
			hasMaxTime := opts != nil && opts.MaxTime != nil
			indexCalls = append(indexCalls, createIndexesCall{collectionName: collectionName, indexesLen: len(indexes), firstIndexKey: firstKey, hasMaxTime: hasMaxTime})
		},
	)

	wantCreated := []string{CollectionSerie, CollectionSerieEpisode, CollectionPerson, CollectionParameter}
	if !reflect.DeepEqual(created, wantCreated) {
		t.Fatalf("unexpected created return: got=%v want=%v", created, wantCreated)
	}
	if !reflect.DeepEqual(createdCalls, wantCreated) {
		t.Fatalf("unexpected create calls: got=%v want=%v", createdCalls, wantCreated)
	}

	if len(indexCalls) != 4 {
		t.Fatalf("expected 4 index calls, got %d", len(indexCalls))
	}
	for _, c := range indexCalls {
		if !c.hasMaxTime {
			t.Fatalf("expected max time on index opts for %s", c.collectionName)
		}
		if c.collectionName == CollectionParameter {
			if c.indexesLen != 1 || c.firstIndexKey != "tipo" {
				t.Fatalf("unexpected parameter index call: %+v", c)
			}
		} else {
			if c.indexesLen != 2 {
				t.Fatalf("unexpected regular index len for %s: %+v", c.collectionName, c)
			}
		}
	}
}

func TestApplyCollectionInitialization_ContinuesOnCreateError(t *testing.T) {
	createdCalls := make([]string, 0)
	indexCalls := make([]string, 0)

	created := applyCollectionInitialization(
		nil,
		func(collectionName string) error {
			createdCalls = append(createdCalls, collectionName)
			if collectionName == CollectionSerie {
				return context.DeadlineExceeded
			}
			return nil
		},
		func(collectionName string, indexes []mongo.IndexModel, opts *options.CreateIndexesOptions) {
			indexCalls = append(indexCalls, collectionName)
		},
	)

	// Even with create error, current behavior still proceeds to index creation and next collections.
	if len(created) != 5 {
		t.Fatalf("expected all 5 collections in return, got %d (%v)", len(created), created)
	}
	if len(createdCalls) != 5 {
		t.Fatalf("expected 5 create calls, got %d", len(createdCalls))
	}
	if len(indexCalls) != 5 {
		t.Fatalf("expected 5 index calls, got %d", len(indexCalls))
	}
}
