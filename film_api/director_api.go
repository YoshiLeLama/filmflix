package film_api

import (
	"filmflix/db_connection"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"strconv"
)

type Director struct {
	Id    primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name  string             `json:"name" bson:"name,omitempty"`
	Films []string           `json:"films"`
}

func InitDirectorApiRoutes(apiRoutes *gin.RouterGroup, client *mongo.Client) {
	InitDirectorCollection(client)
	directorColl = db_connection.GetCollection(client, "films", "directors")

	directorRoutes := apiRoutes.Group("/directors")
	directorRoutes.GET("/", GetDirectors)
	directorRoutes.GET("/:id", GetDirectorById)
	directorRoutes.POST("/", PostDirector)
	directorRoutes.PATCH("/:id", UpdateDirector)
	directorRoutes.DELETE("/:id", DeleteDirector)
}

func GetDirectors(c *gin.Context) {
	limit := 20
	if l := c.Query("limit"); len(l) > 0 {
		limit, _ = strconv.Atoi(l)
	}

	directors := FindDirectors(bson.M{}, limit)
	c.IndentedJSON(http.StatusOK, directors)
}

func GetDirectorById(c *gin.Context) {
	id, _ := primitive.ObjectIDFromHex(c.Param("id"))
	director := FindDirector(bson.M{"_id": id})
	if len(director.Name) > 0 {
		c.IndentedJSON(http.StatusOK, director)
	} else {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "Director not found"})
	}
}

func PostDirector(c *gin.Context) {
	if !CheckAuthKey(c) {
		c.IndentedJSON(http.StatusForbidden, gin.H{"message": "Authentication failed"})
		return
	}

	var newDirector Director
	newDirector.Films = []string{}

	if err := c.BindJSON(&newDirector); err != nil {
		return
	}

	if valid, err := AreFilmsIdsValid(newDirector.Films); !valid {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": err})
		return
	}

	newDirector, err := AddDirector(newDirector)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": err})
	}

	// Add the director id to the films he directed
	for _, film := range newDirector.Films {
		filmInter := film
		go func() {
			_, err := AddDirectorsToFilm(filmInter, []string{newDirector.Id.Hex()})
			if err != nil {
				panic(err)
			}
		}()
	}

	c.IndentedJSON(http.StatusCreated, newDirector)
}

func UpdateDirector(c *gin.Context) {
	if !CheckAuthKey(c) {
		c.IndentedJSON(http.StatusForbidden, gin.H{"message": "Authentication failed"})
		return
	}

	var updateData Director
	idString := c.Param("id")

	id, err := primitive.ObjectIDFromHex(idString)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	oldDirector := FindDirector(bson.M{"_id": id})

	if err := c.BindJSON(&updateData); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	result := UpdateDirectorById(idString, updateData)

	removedFilms := difference(oldDirector.Films, updateData.Films)
	for _, film := range removedFilms {
		film := film
		go func() {
			_, err := RemoveDirectorsFromFilm(film, []string{idString})
			if err != nil {
				panic(err)
			}
		}()
	}

	newFilms := difference(updateData.Films, oldDirector.Films)
	for _, film := range newFilms {
		film := film
		go func() {
			_, err := AddDirectorsToFilm(film, []string{idString})
			if err != nil {
				panic(err)
			}
		}()
	}

	c.IndentedJSON(http.StatusNoContent, result)
}

func DeleteDirector(c *gin.Context) {
	if !CheckAuthKey(c) {
		c.IndentedJSON(http.StatusForbidden, gin.H{"message": "Authentication failed"})
		return
	}

	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	oldDirector := FindDirector(bson.M{"_id": id})

	result := DeleteItemById(directorColl, id.Hex())

	if result == 0 {
		c.IndentedJSON(http.StatusNotModified, gin.H{"message": "No film with the specified id"})
		return
	}

	for _, film := range oldDirector.Films {
		_, err := RemoveDirectorsFromFilm(film, []string{id.Hex()})
		if err != nil {
			return
		}
	}

	c.IndentedJSON(http.StatusNoContent, result)
}
