package film_api

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func DeleteItemById(collection *mongo.Collection, idString string) int64 {
	id, _ := primitive.ObjectIDFromHex(idString)

	result, err := collection.DeleteOne(context.TODO(), bson.D{{"_id", id}})
	if err != nil {
		panic(err)
	}

	return result.DeletedCount
}
