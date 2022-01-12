package film_api

import (
	"context"
	"filmflix/db_connection"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var directorColl *mongo.Collection

func InitDirectorCollection(client *mongo.Client) {
	directorColl = db_connection.GetCollection(client, "films", "actors")
}

func FindDirectors(filter bson.M, maxCount int) []Director {
	var results []Director
	limit := int64(maxCount)
	cursor, err := directorColl.Find(context.TODO(), filter, &options.FindOptions{Limit: &limit, Sort: bson.D{{"title", 1}}})
	if err != nil {
		panic(err)
	}

	if err = cursor.All(context.TODO(), &results); err != nil {
		panic(err)
	}

	return results
}

func FindDirector(filter bson.M) Director {
	var director Director
	err := directorColl.FindOne(context.TODO(), filter).Decode(&director)

	if err == mongo.ErrNoDocuments {
		fmt.Printf("No document was found\n")
		return director
	}
	if err != nil {
		panic(err)
	}

	return director
}

func AddDirector(director Director) (Director, error) {
	director.Id = primitive.NewObjectID()

	_, err := directorColl.InsertOne(context.TODO(), director)
	if err != nil {
		return Director{}, err
	}

	return director, nil
}

func UpdateDirectorById(idString string, data interface{}) int64 {
	id, _ := primitive.ObjectIDFromHex(idString)

	result, err := directorColl.UpdateOne(context.TODO(), bson.D{{"_id", id}}, bson.D{{"$set", data}})
	if err != nil {
		panic(err)
	}

	return result.ModifiedCount
}

func ReplaceDirector(idString string, newDirector Film) int64 {
	id, _ := primitive.ObjectIDFromHex(idString)
	newDirector.Id = id

	result, err := directorColl.ReplaceOne(context.TODO(), bson.D{{"_id", id}}, newDirector)
	if err != nil {
		panic(err)
	}

	return result.ModifiedCount
}

func AreDirectorsIdsValid(ids []string) (bool, gin.H) {
	tempIds := make(map[string]struct{})
	for i, id := range ids {
		tempIds[id] = struct{}{}
		if !primitive.IsValidObjectID(id) {
			return false, gin.H{"message": fmt.Sprintf("Id of item no. %v is invalid", i)}
		}
	}

	directorsIds := make([]primitive.ObjectID, len(tempIds))
	i := 0
	for k := range tempIds {
		directorsIds[i], _ = primitive.ObjectIDFromHex(k)
		i++
	}

	result := FindDirectors(bson.M{"_id": bson.M{"$in": directorsIds}}, len(directorsIds))

	if len(result) != len(directorsIds) {
		return false, gin.H{"message": "At least one director id does not exists"}
	}
	return true, nil
}

func AddFilmsToDirector(idString string, films []string) (int64, error) {
	id, err := primitive.ObjectIDFromHex(idString)

	if err != nil {
		return 0, err
	}

	result, err := directorColl.UpdateOne(context.TODO(), bson.D{{"_id", id}}, bson.M{"$push": bson.M{"films": bson.M{"$each": films}}})
	if err != nil {
		return 0, err
	}

	return result.ModifiedCount, nil
}

func RemoveFilmsFromDirector(idString string, films []string) (int64, error) {
	id, err := primitive.ObjectIDFromHex(idString)
	if err != nil {
		return 0, err
	}

	result, err := directorColl.UpdateOne(context.TODO(), bson.M{"_id": id}, bson.M{"$pull": bson.M{"films": bson.M{"$in": films}}})
	if err != nil {
		return 0, err
	}

	return result.ModifiedCount, nil
}
