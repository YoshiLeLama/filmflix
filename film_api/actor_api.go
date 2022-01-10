package film_api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"strconv"
)

type Actor struct {
	Id    primitive.ObjectID `bson:"_id" json:"id,omitempty"`
	Name  string             `json:"name"`
	Films []string           `json:"films"` // Films is the slice of the films the actor played in
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
	var updateData interface{}

	if err := c.BindJSON(&updateData); err != nil {
		return
	}

	result := UpdateActorById(c.Param("id"), updateData)
	c.IndentedJSON(http.StatusNoContent, result)
}

func DeleteActor(c *gin.Context) {
	id := c.Param("id")
	filmIdObj, _ := primitive.ObjectIDFromHex(id)

	film := FindFilm(bson.D{{"_id", filmIdObj}})

	result := DeleteItemById(filmColl, id)

	if result == 0 {
		c.IndentedJSON(http.StatusNotModified, gin.H{"message": "No film with the specified id"})
		return
	}

	directorId, _ := primitive.ObjectIDFromHex(film.Director)
	director := FindDirector(directorColl, bson.D{{"_id", directorId}})
	fmt.Println(directorId)

	var newDirectorFilms []string
	for _, v := range director.Films {
		if v != id {
			newDirectorFilms = append(newDirectorFilms, v)
		}
	}
	director.Films = newDirectorFilms
	_ = UpdateDirectorById(directorColl, director.Id.Hex(), director)

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
