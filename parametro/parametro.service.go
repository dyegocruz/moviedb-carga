package parametro

import (
	"context"
	"moviedb/database"
	"os"

	"gopkg.in/mgo.v2/bson"
)

func GetByTipo(tipo string) Parametro {
	client, ctx, cancel := database.GetConnection()
	defer cancel()
	defer client.Disconnect(ctx)

	var parametro Parametro
	client.Database(os.Getenv("MONGO_DATABASE")).Collection("parametro").FindOne(context.TODO(), bson.M{"tipo": tipo}).Decode(&parametro)

	return parametro
}
