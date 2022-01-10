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

type Role struct {
	Name    string `json:"name"` // Name is the name of the role
	ActorId string `json:"actor" bson:"actor"`
}

type Film struct {
	Id            primitive.ObjectID `bson:"_id" json:"id,omitempty"`
	Title         string             `json:"title"`
	OriginalTitle string             `bson:"original_title" json:"original_title"`
	Description   string             `json:"description"`
	Director      string             `json:"director"` // Represents the director's name or id
	Poster        string             `json:"poster"`
	ReleaseDate   string             `bson:"release_date" json:"release_date"`
	Rating        string             `bson:"rt_score" json:"rt_score"`
	Roles         []Role             `json:"roles"`
}

func InitFilmApiRoutes(apiRoutes *gin.RouterGroup, client *mongo.Client) {
	InitFilmCollection(client)

	filmRoutes := apiRoutes.Group("/films")
	filmRoutes.GET("/", GetFilms)
	filmRoutes.POST("/", PostFilm)
	filmRoutes.GET("/:id", GetFilmById)
	filmRoutes.PATCH("/:id", UpdateFilm)
	filmRoutes.PATCH("/:id/roles", UpdateRoles)
	filmRoutes.DELETE("/:id", DeleteFilm)
}

func GetFilms(c *gin.Context) {
	limit := 20
	if l := c.Query("limit"); len(l) > 0 {
		limit, _ = strconv.Atoi(l)
	}
	movies := FindFilms(bson.M{}, limit)

	c.IndentedJSON(http.StatusOK, movies)
}

func PostFilm(c *gin.Context) {
	var newFilm Film
	newFilm.Roles = []Role{}
	if err := c.BindJSON(&newFilm); err != nil {
		return
	}

	if len(newFilm.Title) <= 0 || len(newFilm.Director) <= 0 || len(newFilm.ReleaseDate) <= 0 {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Data is invalid"})
		return
	}

	result := AddFilm(newFilm)

	c.IndentedJSON(http.StatusCreated, result)
}

func UpdateFilm(c *gin.Context) {
	var updateData interface{}

	if err := c.BindJSON(&updateData); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	if _, err := UpdateFilmById(filmColl, c.Param("id"), updateData); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusNoContent, gin.H{})
}

func DeleteFilm(c *gin.Context) {
	id := c.Param("id")
	filmIdObj, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Id is invalid"})
		return
	}

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

func GetFilmById(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))

	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Id is invalid"})
		return
	}

	film := FindFilm(bson.D{{"_id", id}})

	if film.Title != "" {
		c.IndentedJSON(http.StatusOK, film)
		return
	}

	c.IndentedJSON(http.StatusNotFound, gin.H{"message": "Film not found"})
}

type UpdateRolesReq struct {
	Replace bool   `json:"replace"`
	Roles   []Role `json:"roles"`
}

func UpdateRoles(c *gin.Context) {
	var req UpdateRolesReq

	if err := c.BindJSON(&req); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "JSON is invalid"})
		return
	}

	if valid, msg := AreActorsIdsValid(req.Roles); !valid {
		c.IndentedJSON(http.StatusBadRequest, msg)
		return
	}

	for _, role := range req.Roles {
		roleInter := role
		go func() {
			_, err := AddFilmsToActor(roleInter.ActorId, []string{c.Param("id")})
			if err != nil {
				panic(err)
			}
		}()
	}

	fmt.Println(req)

	if _, err := AddRoles(c.Param("id"), req.Roles); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusNoContent, gin.H{})
}
