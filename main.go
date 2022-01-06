package main

import (
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"net/http"
	"os"
	"regexp"
)

var client *mongo.Client
var filmColl *mongo.Collection

func main() {
	_ = godotenv.Load(".env")

	client = ConnectToDB()
	filmColl = GetCollection(client, "films", "films")

	defer DisconnectFromDB(client)

	router := gin.Default()
	router.StaticFile("/favicon.ico", "./app/favicon.ico")
	router.Use(serveJS)
	router.Static("/static", "./app/static")
	router.GET("/", serveApp)

	apiRoutes := router.Group("/api")
	apiRoutes.GET("/films", getFilms)
	apiRoutes.POST("/films", postFilms)
	apiRoutes.PATCH("/films/:id", updateFilm)
	apiRoutes.DELETE("/films/:id", deleteFilm)
	apiRoutes.GET("/films/:id", getFilmByID)

	err := router.Run(":" + os.Getenv("PORT"))
	if err != nil {
		return
	}
}

func serveApp(c *gin.Context) {
	c.Status(http.StatusOK)
	c.File("app/index.html")
}
func serveJS(c *gin.Context) {
	if matched, _ := regexp.Match(`js\z`, []byte(c.Request.RequestURI)); matched {
		c.Writer.Header().Set("Content-Type", "application/javascript")
		c.Next()
	}
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
