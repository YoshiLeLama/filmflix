package film_api

import (
	"context"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"os"
)

func ConnectToDB() (client *mongo.Client) {
	dbUrl := os.Getenv("DB_URL")
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(dbUrl))
	if err != nil {
		panic(err)
	}
	return
}

func DisconnectFromDB(client *mongo.Client) {
	if err := client.Disconnect(context.TODO()); err != nil {
		panic(err)
	}
}

func GetCollection(client *mongo.Client, db string, coll string) *mongo.Collection {
	return client.Database(db).Collection(coll)
}

func FindFilm(collection *mongo.Collection, criteria bson.D) film {
	var film film
	err := collection.FindOne(context.TODO(), criteria).Decode(film)

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

func FindFilms(collection *mongo.Collection, criteria bson.D, maxCount int) []film {
	var results []film
	limit := int64(maxCount)
	cursor, err := collection.Find(context.TODO(), criteria, &options.FindOptions{Limit: &limit})
	if err != nil {
		panic(err)
	}

	if err = cursor.All(context.TODO(), &results); err != nil {
		panic(err)
	}

	return results
}

func AddFilm(collection *mongo.Collection, item film) film {
	item.ID = primitive.NewObjectID()

	_, err := collection.InsertOne(context.TODO(), item)
	if err != nil {
		panic(err)
	}

	return item
}

func UpdateFilm(collection *mongo.Collection, idString string, data interface{}) int64 {
	id, _ := primitive.ObjectIDFromHex(idString)

	result, err := collection.UpdateOne(context.TODO(), bson.D{{"_id", id}}, bson.D{{"$set", data}})
	if err != nil {
		panic(err)
	}

	return result.ModifiedCount
}

func DeleteFilm(collection *mongo.Collection, idString string) int64 {
	id, _ := primitive.ObjectIDFromHex(idString)

	result, err := collection.DeleteOne(context.TODO(), bson.D{{"_id", id}})
	if err != nil {
		panic(err)
	}

	return result.DeletedCount
}
