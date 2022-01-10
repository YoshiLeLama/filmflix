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
	Id    primitive.ObjectID `json:"id" bson:"_id"`
	Name  string             `json:"name"`
	Films []string           `json:"films"`
}

var directorColl *mongo.Collection

func InitDirectorApiRoutes(apiRoutes *gin.RouterGroup, client *mongo.Client) {
	directorColl = db_connection.GetCollection(client, "films", "directors")

	directorRoutes := apiRoutes.Group("/directors")
	directorRoutes.GET("/", GetDirectors)
	directorRoutes.GET("/:id", GetDirectorById)
	directorRoutes.POST("/")
}

func GetDirectors(c *gin.Context) {
	limit := 20
	if l := c.Query("limit"); len(l) > 0 {
		limit, _ = strconv.Atoi(l)
	}

	directors := FindDirectors(directorColl, bson.D{}, limit)
	c.IndentedJSON(http.StatusOK, directors)
}

func GetDirectorById(c *gin.Context) {
	id, _ := primitive.ObjectIDFromHex(c.Param("id"))
	director := FindDirector(directorColl, bson.D{{"_id", id}})
	if len(director.Name) > 0 {
		c.IndentedJSON(http.StatusOK, director)
	} else {
		c.IndentedJSON(http.StatusNotFound, gin.H{"message": "Director not found"})
	}
}

func PostDirector(c *gin.Context) {
	// TODO
}

func UpdateDirector(c *gin.Context) {
	// TODO
}

func DeleteDirector(c *gin.Context) {

}
