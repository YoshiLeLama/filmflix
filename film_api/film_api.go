package film_api

import (
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
)

var client *mongo.Client
var filmColl *mongo.Collection

func InitFilmApiRoutes(router *gin.Engine) {
	client = ConnectToDB()
	filmColl = GetCollection(client, "films", "films")

	apiRoutes := router.Group("/api")
	apiRoutes.GET("/films", getFilms)
	apiRoutes.POST("/films", postFilms)
	apiRoutes.PATCH("/films/:id", updateFilm)
	apiRoutes.DELETE("/films/:id", deleteFilm)
	apiRoutes.GET("/films/:id", getFilmByID)
}

func CloseFilmApiRoutes() {
	defer DisconnectFromDB(client)
}

func getFilms(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")

	//c.IndentedJSON(http.StatusOK, films)
	movies := FindFilms(filmColl, bson.D{})

	c.IndentedJSON(http.StatusOK, movies)
}

func postFilms(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")

	var newFilm film

	if err := c.BindJSON(&newFilm); err != nil {
		return
	}

	result := AddFilm(filmColl, newFilm)
	c.IndentedJSON(http.StatusCreated, result)
}

func updateFilm(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")

	var updateData interface{}

	if err := c.BindJSON(&updateData); err != nil {
		return
	}

	result := UpdateFilm(filmColl, c.Param("id"), updateData)
	c.IndentedJSON(http.StatusNoContent, result)
}

func deleteFilm(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")

	result := DeleteFilm(filmColl, c.Param("id"))
	c.IndentedJSON(http.StatusNoContent, result)
}

func getFilmByID(c *gin.Context) {
	id, _ := primitive.ObjectIDFromHex(c.Param("id"))

	film := FindFilm(filmColl, bson.D{{"_id", id}})

	if film.Title != "" {
		c.IndentedJSON(http.StatusOK, film)
		return
	}

	c.IndentedJSON(http.StatusNotFound, gin.H{"message": "film not found"})
}
