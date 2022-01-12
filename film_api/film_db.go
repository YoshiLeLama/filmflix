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

var filmColl *mongo.Collection

func InitFilmCollection(client *mongo.Client) {
	filmColl = db_connection.GetCollection(client, "films", "films")
}

func FindFilm(filter bson.M) Film {
	var film Film
	err := filmColl.FindOne(context.TODO(), filter).Decode(&film)

	if err == mongo.ErrNoDocuments {
		fmt.Printf("No document was found\n")
		return film
	}
	if err != nil {
		panic(err)
	}

	fmt.Println(film)

	return film
}

// FindFilms retrieves a given amount of films from the given collection
func FindFilms(filter bson.M, maxCount int) []Film {
	var results []Film
	limit := int64(maxCount)
	cursor, err := filmColl.Find(context.TODO(), filter, &options.FindOptions{Limit: &limit, Sort: bson.D{{"title", 1}}})

	if err != nil {
		panic(err)
	}

	if err = cursor.All(context.TODO(), &results); err != nil {
		panic(err)
	}

	return results
}

// AddFilm adds a film to the given collection and return the film with its id
func AddFilm(film Film) Film {
	film.Id = primitive.NewObjectID()

	_, err := filmColl.InsertOne(context.TODO(), film)
	if err != nil {
		panic(err)
	}

	return film
}

// UpdateFilmById update a film of the database with the given data and returns the number of modified items
func UpdateFilmById(idString string, data interface{}) (int64, error) {
	id, _ := primitive.ObjectIDFromHex(idString)

	result, err := filmColl.UpdateOne(context.TODO(), bson.D{{"_id", id}}, bson.D{{"$set", data}})
	if err != nil {
		return 0, err
	}

	return result.ModifiedCount, nil
}

// ReplaceFilm replaces the item with the given id and returns the number of modified items
func ReplaceFilm(idString string, newFilm Film) int64 {
	id, _ := primitive.ObjectIDFromHex(idString)
	newFilm.Id = id

	result, err := filmColl.ReplaceOne(context.TODO(), bson.D{{"_id", id}}, newFilm)
	if err != nil {
		panic(err)
	}

	return result.ModifiedCount
}

func AddRoles(idString string, roles []Role) (int64, error) {
	id, err := primitive.ObjectIDFromHex(idString)

	if err != nil {
		return 0, err
	}

	result, err := filmColl.UpdateOne(context.TODO(), bson.D{{"_id", id}}, bson.M{"$push": bson.M{"roles": bson.M{"$each": roles}}})
	if err != nil {
		return 0, err
	}

	return result.ModifiedCount, nil
}

func AddActorsToFilm(idString string, actors []Role) (int64, error) {
	id, err := primitive.ObjectIDFromHex(idString)

	if err != nil {
		return 0, err
	}

	result, err := filmColl.UpdateOne(context.TODO(), bson.D{{"_id", id}}, bson.M{"$push": bson.M{"roles": bson.M{"$each": actors}}})
	if err != nil {
		return 0, err
	}

	return result.ModifiedCount, nil
}

func RemoveActorsFromFilm(idString string, actorsIds []string) (int64, error) {
	id, err := primitive.ObjectIDFromHex(idString)
	if err != nil {
		return 0, err
	}

	result, err := filmColl.UpdateOne(context.TODO(), bson.M{"_id": id}, bson.M{"$pull": bson.M{"roles": bson.M{"actor": bson.M{"$in": actorsIds}}}})
	if err != nil {
		return 0, err
	}

	return result.ModifiedCount, err
}
func AddDirectorsToFilm(idString string, directors []string) (int64, error) {
	id, err := primitive.ObjectIDFromHex(idString)

	if err != nil {
		return 0, err
	}

	result, err := filmColl.UpdateOne(context.TODO(), bson.D{{"_id", id}}, bson.M{"$push": bson.M{"directors": bson.M{"$each": directors}}})
	if err != nil {
		return 0, err
	}

	return result.ModifiedCount, nil
}

func RemoveDirectorsFromFilm(idString string, directorsIds []string) (int64, error) {
	id, err := primitive.ObjectIDFromHex(idString)
	if err != nil {
		return 0, err
	}

	result, err := filmColl.UpdateOne(context.TODO(), bson.M{"_id": id}, bson.M{"$pull": bson.M{"directors": bson.M{"$in": directorsIds}}})
	if err != nil {
		return 0, err
	}

	return result.ModifiedCount, err
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
