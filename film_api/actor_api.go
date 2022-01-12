package film_api

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"strconv"
)

type Actor struct {
	Id    primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Name  string             `json:"name,omitempty" bson:"name"`
	Films []string           `json:"films" bson:"films"` // Films is the slice of the films the actor played in
}

func InitActorApiRoutes(apiRoutes *gin.RouterGroup, client *mongo.Client) {
	InitActorCollection(client)

	actorRoutes := apiRoutes.Group("/actors")
	actorRoutes.GET("/", GetActors)
	actorRoutes.POST("/", PostActor)
	actorRoutes.PATCH("/:id", UpdateActor)
	actorRoutes.DELETE("/:id", DeleteActor)
	actorRoutes.GET("/:id", GetActorById)
}

func GetActors(c *gin.Context) {
	limit := 20
	if l := c.Query("limit"); len(l) > 0 {
		limit, _ = strconv.Atoi(l)
	}
	actors := FindActors(bson.M{}, limit)

	c.IndentedJSON(http.StatusOK, actors)
}

func PostActor(c *gin.Context) {
	if !CheckAuthKey(c) {
		c.IndentedJSON(http.StatusForbidden, gin.H{"message": "Authentication failed"})
		return
	}

	var newActor Actor

	if err := c.BindJSON(&newActor); err != nil {
		return
	}

	if valid, err := AreFilmsIdsValid(newActor.Films); !valid {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": err})
		return
	}

	newActor, err := AddActor(newActor)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": err})
	}

	// Add the actor id to the films he played in
	for _, film := range newActor.Films {
		filmInter := film
		go func() {
			_, err := AddActorsToFilm(filmInter, []Role{{ActorId: newActor.Id.Hex()}})
			if err != nil {
				panic(err)
			}
		}()
	}

	c.IndentedJSON(http.StatusCreated, newActor)
}

func UpdateActor(c *gin.Context) {
	if !CheckAuthKey(c) {
		c.IndentedJSON(http.StatusForbidden, gin.H{"message": "Authentication failed"})
		return
	}

	var updateData Actor
	idString := c.Param("id")

	id, err := primitive.ObjectIDFromHex(idString)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"0message": err.Error()})
		return
	}

	oldData := FindActor(bson.M{"_id": id})

	if err := c.BindJSON(&updateData); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"1message": err.Error()})
		return
	}

	result := UpdateActorById(idString, updateData)

	removedFilms := difference(oldData.Films, updateData.Films)
	for _, film := range removedFilms {
		film := film
		go func() {
			_, err := RemoveActorsFromFilm(film, []string{idString})
			if err != nil {
				panic(err)
			}
		}()
	}

	newFilms := difference(updateData.Films, oldData.Films)
	for _, film := range newFilms {
		film := film
		go func() {
			_, err := AddActorsToFilm(film, []Role{{ActorId: idString}})
			if err != nil {
				panic(err)
			}
		}()
	}

	c.IndentedJSON(http.StatusNoContent, result)
}

func DeleteActor(c *gin.Context) {
	if !CheckAuthKey(c) {
		c.IndentedJSON(http.StatusForbidden, gin.H{"message": "Authentication failed"})
		return
	}

	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	oldActor := FindActor(bson.M{"_id": id})

	result := DeleteItemById(actorColl, id.Hex())

	if result == 0 {
		c.IndentedJSON(http.StatusNotModified, gin.H{"message": "No film with the specified id"})
		return
	}

	for _, film := range oldActor.Films {
		_, err := RemoveActorsFromFilm(film, []string{id.Hex()})
		if err != nil {
			return
		}
	}

	c.IndentedJSON(http.StatusNoContent, result)
}

func GetActorById(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))

	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Id is invalid"})
		return
	}

	actor := FindActor(bson.M{"_id": id})

	if actor.Name != "" {
		c.IndentedJSON(http.StatusOK, actor)
		return
	}

	c.IndentedJSON(http.StatusNotFound, gin.H{"message": "Actor not found"})
}
