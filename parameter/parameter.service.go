package parameter

import (
	"context"
	"moviedb/database"

	"go.mongodb.org/mongo-driver/mongo"
	"gopkg.in/mgo.v2/bson"
)

var parameterCollectionString = database.COLLECTION_PARAMETER
var parameterCollection *mongo.Collection = database.GetCollection(database.DB, parameterCollectionString)

func GetByType(paramType string) Parameter {

	var parameter Parameter
	parameterCollection.FindOne(context.TODO(), bson.M{"paramType": paramType}).Decode(&parameter)

	return parameter
}
