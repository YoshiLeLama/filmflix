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

func FindFilm(filter bson.D) Film {
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

	director := FindDirector(directorColl, bson.D{{"name", film.Director}})

	if len(director.Id.String()) == 0 {
		director = AddDirector(directorColl, Director{
			Name:  film.Director,
			Films: []string{film.Id.Hex()},
		})
	} else {
		director.Films = append(director.Films, film.Id.Hex())
		fmt.Println(director)
		UpdateDirectorById(directorColl, director.Id.Hex(), director)
	}

	film.Director = director.Id.Hex()

	_, err := filmColl.InsertOne(context.TODO(), film)
	if err != nil {
		panic(err)
	}

	return film
}

// UpdateFilmById update a film of the database with the given data and returns the number of modified items
func UpdateFilmById(collection *mongo.Collection, idString string, data interface{}) (int64, error) {
	id, _ := primitive.ObjectIDFromHex(idString)

	result, err := collection.UpdateOne(context.TODO(), bson.D{{"_id", id}}, bson.D{{"$set", data}})
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

func AreActorsIdsValid(roles []Role) (bool, gin.H) {
	tempIds := make(map[string]struct{})
	for i, role := range roles {
		tempIds[role.ActorId] = struct{}{}
		if !primitive.IsValidObjectID(role.ActorId) {
			return false, gin.H{"message": fmt.Sprintf("Id of item no. %v is invalid", i)}
		}
	}

	actorsIds := make([]primitive.ObjectID, len(tempIds))
	i := 0
	for k := range tempIds {
		actorsIds[i], _ = primitive.ObjectIDFromHex(k)
		i++
	}

	result := FindActors(bson.M{"_id": bson.M{"$in": actorsIds}}, len(actorsIds))

	if len(result) != len(actorsIds) {
		return false, gin.H{"message": "At least one actor id does not exists"}
	}
	return true, nil
}
