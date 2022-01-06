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

type film struct {
	ID            primitive.ObjectID `bson:"_id" json:"id,omitempty"`
	Title         string             `json:"title"`
	OriginalTitle string             `bson:"original_title" json:"original_title"`
	Description   string             `json:"description"`
	Director      string             `json:"director"`
	Image         string             `json:"image"`
	MovieBanner   string             `bson:"movie_banner" json:"movie_banner"`
	ReleaseDate   string             `bson:"release_date" json:"release_date"`
	Rating        string             `bson:"rt_score" json:"rt_score"`
}

func ConnectToDB() (client *mongo.Client) {
	dbUrl := os.Getenv("DB_URL")
	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(dbUrl))
	if err != nil {
		panic(err)
	}
	return
}

func GetCollection(client *mongo.Client, db string, coll string) *mongo.Collection {
	return client.Database(db).Collection(coll)
}

func FindFilm(collection *mongo.Collection, criteria bson.D) film {
	var result film
	err := collection.FindOne(context.TODO(), criteria).Decode(&result)

	if err == mongo.ErrNoDocuments {
		fmt.Printf("No document was found\n")
		return film{}
	}
	if err != nil {
		panic(err)
	}

	fmt.Println(result)

	return result
}

func FindFilms(collection *mongo.Collection, criteria bson.D) []film {
	cursor, err := collection.Find(context.TODO(), criteria)
	if err != nil {
		panic(err)
	}

	var results []film
	if err = cursor.All(context.TODO(), &results); err != nil {
		panic(err)
	}

	return results
}

func AddFilm(collection *mongo.Collection, film film) interface{} {
	film.ID = primitive.NewObjectID()

	_, err := collection.InsertOne(context.TODO(), film)
	if err != nil {
		panic(err)
	}

	return film
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

func DisconnectFromDB(client *mongo.Client) {
	if err := client.Disconnect(context.TODO()); err != nil {
		panic(err)
	}
}
