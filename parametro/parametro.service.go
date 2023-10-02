package parametro

import (
	"context"
	"moviedb/database"

	"go.mongodb.org/mongo-driver/mongo"
	"gopkg.in/mgo.v2/bson"
)

var parametroCollectionString = database.COLLECTION_PARAMETRO
var parametroCollection *mongo.Collection = database.GetCollection(database.DB, parametroCollectionString)

func GetByTipo(tipo string) Parametro {

	var parametro Parametro
	parametroCollection.FindOne(context.TODO(), bson.M{"tipo": tipo}).Decode(&parametro)

	return parametro
}
