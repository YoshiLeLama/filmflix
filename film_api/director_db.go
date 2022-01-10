package film_api

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func FindDirectors(collection *mongo.Collection, filter bson.D, maxCount int) []Director {
	var results []Director
	limit := int64(maxCount)
	cursor, err := collection.Find(context.TODO(), filter, &options.FindOptions{Limit: &limit, Sort: bson.D{{"title", 1}}})
	if err != nil {
		panic(err)
	}

	if err = cursor.All(context.TODO(), &results); err != nil {
		panic(err)
	}

	return results
}

func FindDirector(collection *mongo.Collection, filter bson.D) Director {
	var director Director
	err := collection.FindOne(context.TODO(), filter).Decode(&director)

	if err == mongo.ErrNoDocuments {
		fmt.Printf("No document was found\n")
		return director
	}
	if err != nil {
		panic(err)
	}

	return director
}

func AddDirector(collection *mongo.Collection, director Director) Director {
	director.Id = primitive.NewObjectID()

	_, err := collection.InsertOne(context.TODO(), director)
	if err != nil {
		panic(err)
	}

	return director
}

func UpdateDirectorById(collection *mongo.Collection, idString string, data interface{}) int64 {
	id, _ := primitive.ObjectIDFromHex(idString)

	result, err := collection.UpdateOne(context.TODO(), bson.D{{"_id", id}}, bson.D{{"$set", data}})
	if err != nil {
		panic(err)
	}

	return result.ModifiedCount
}

func ReplaceDirector(collection *mongo.Collection, idString string, newDirector Film) int64 {
	id, _ := primitive.ObjectIDFromHex(idString)
	newDirector.Id = id

	result, err := collection.ReplaceOne(context.TODO(), bson.D{{"_id", id}}, newDirector)
	if err != nil {
		panic(err)
	}

	return result.ModifiedCount
}
