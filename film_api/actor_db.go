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

var actorColl *mongo.Collection

func InitActorCollection(client *mongo.Client) {
	actorColl = db_connection.GetCollection(client, "films", "actors")
}

func FindActors(filter bson.M, maxCount int) []Actor {
	var results []Actor
	limit := int64(maxCount)
	cursor, err := actorColl.Find(context.TODO(), filter, &options.FindOptions{Limit: &limit, Sort: bson.D{{"title", 1}}})
	if err != nil {
		panic(err)
	}

	if err = cursor.All(context.TODO(), &results); err != nil {
		panic(err)
	}

	return results
}

func FindActor(filter bson.M) Actor {
	var actor Actor
	err := actorColl.FindOne(context.TODO(), filter).Decode(&actor)

	if err == mongo.ErrNoDocuments {
		fmt.Printf("No document was found\n")
		return actor
	}
	if err != nil {
		panic(err)
	}

	return actor
}

func AddActor(actor Actor) (Actor, error) {
	actor.Id = primitive.NewObjectID()

	_, err := actorColl.InsertOne(context.TODO(), actor)
	if err != nil {
		return Actor{}, nil
	}

	return actor, nil
}

func UpdateActorById(idString string, data interface{}) int64 {
	id, _ := primitive.ObjectIDFromHex(idString)

	result, err := actorColl.UpdateOne(context.TODO(), bson.D{{"_id", id}}, bson.D{{"$set", data}})
	if err != nil {
		panic(err)
	}

	return result.ModifiedCount
}

func ReplaceActor(idString string, newDirector Film) int64 {
	id, _ := primitive.ObjectIDFromHex(idString)
	newDirector.Id = id

	result, err := actorColl.ReplaceOne(context.TODO(), bson.D{{"_id", id}}, newDirector)
	if err != nil {
		panic(err)
	}

	return result.ModifiedCount
}

func AddFilmsToActor(idString string, films []string) (int64, error) {
	id, err := primitive.ObjectIDFromHex(idString)

	if err != nil {
		return 0, err
	}

	result, err := actorColl.UpdateOne(context.TODO(), bson.D{{"_id", id}}, bson.M{"$push": bson.M{"films": bson.M{"$each": films}}})
	if err != nil {
		return 0, err
	}

	return result.ModifiedCount, nil
}

func RemoveFilmsFromActor(idString string, films []string) (int64, error) {
	id, err := primitive.ObjectIDFromHex(idString)

	if err != nil {
		return 0, err
	}

	filmsId := make([]primitive.ObjectID, len(films))
	i := 0
	for _, v := range films {
		filmsId[i], _ = primitive.ObjectIDFromHex(v)
		i++
	}

	result, err := actorColl.UpdateOne(context.TODO(), bson.M{"_id": id}, bson.M{"$pull": bson.M{"films": bson.M{"$in": filmsId}}})
	if err != nil {
		return 0, err
	}

	return result.ModifiedCount, nil
}

func AreFilmsIdsValid(ids []string) (bool, gin.H) {
	tempIds := make(map[string]struct{})
	for i, id := range ids {
		tempIds[id] = struct{}{}
		if !primitive.IsValidObjectID(id) {
			return false, gin.H{"message": fmt.Sprintf("Id of film no. %v is invalid", i)}
		}
	}

	filmsIds := make([]primitive.ObjectID, len(tempIds))
	i := 0
	for k := range tempIds {
		filmsIds[i], _ = primitive.ObjectIDFromHex(k)
		i++
	}

	result := FindFilms(bson.M{"_id": bson.M{"$in": filmsIds}}, len(filmsIds))

	if len(result) != len(filmsIds) {
		return false, gin.H{"message": "At least one film id does not exists"}
	}

	return true, nil
}
