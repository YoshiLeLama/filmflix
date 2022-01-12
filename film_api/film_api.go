package film_api

import (
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
	Id            primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Title         string             `bson:"title,omitempty" json:"title"`
	OriginalTitle string             `bson:"original_title,omitempty" json:"original_title"`
	Description   string             `bson:"description,omitempty" json:"description"`
	Directors     []string           `bson:"directors,omitempty" json:"directors"` // Represents the directors ids
	Poster        string             `bson:"poster,omitempty" json:"poster"`
	ReleaseDate   string             `bson:"release_date,omitempty" json:"release_date"`
	Rating        string             `bson:"rt_score,omitempty" json:"rt_score"`
	Roles         []Role             `bson:"roles,omitempty" json:"roles"`
}

func InitFilmApiRoutes(apiRoutes *gin.RouterGroup, client *mongo.Client) {
	InitFilmCollection(client)

	filmRoutes := apiRoutes.Group("/films")
	filmRoutes.GET("/", GetFilms)
	filmRoutes.POST("/", PostFilm)
	filmRoutes.GET("/:id", GetFilmById)
	filmRoutes.PATCH("/:id", UpdateFilm)
	filmRoutes.PATCH("/:id/roles", UpdateRoles)
	filmRoutes.PATCH("/:id/directors", UpdateDirectors)
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
	if !CheckAuthKey(c) {
		c.IndentedJSON(http.StatusForbidden, gin.H{"message": "Authentication failed"})
		return
	}

	var newFilm Film
	newFilm.Roles = []Role{}
	newFilm.Directors = []string{}
	if err := c.BindJSON(&newFilm); err != nil {
		return
	}
	actorsValid, _ := AreActorsIdsValid(newFilm.Roles)
	directorsValid, _ := AreDirectorsIdsValid(newFilm.Directors)

	if !actorsValid || !directorsValid || len(newFilm.Directors) <= 0 || len(newFilm.Title) <= 0 || len(newFilm.ReleaseDate) <= 0 {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Data is invalid, check if required fields are filled and if roles and director ids are valid."})
		return
	}

	newFilm = AddFilm(newFilm)

	for _, director := range newFilm.Directors {
		directorInter := director
		go func() {
			_, err := AddFilmsToDirector(directorInter, []string{newFilm.Id.Hex()})
			if err != nil {
				panic(err)
			}
		}()
	}

	// Add the actor id to the films he played in
	for _, role := range newFilm.Roles {
		roleInter := role
		go func() {
			_, err := AddFilmsToActor(roleInter.ActorId, []string{newFilm.Id.Hex()})
			if err != nil {
				panic(err)
			}
		}()
	}

	c.IndentedJSON(http.StatusCreated, newFilm)
}

// UpdateFilm is used to update all the fields of a film EXCEPT the roles (use UPDATE /api/films/<id>/roles instead) and the directors (use UPDATE /api/films/<id>/roles instead)
func UpdateFilm(c *gin.Context) {
	if !CheckAuthKey(c) {
		c.IndentedJSON(http.StatusForbidden, gin.H{"message": "Authentication failed"})
		return
	}

	var updateData Film

	if err := c.BindJSON(&updateData); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	updateData.Roles = []Role{}
	updateData.Directors = []string{}

	if _, err := UpdateFilmById(c.Param("id"), updateData); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusNoContent, gin.H{})
}

func DeleteFilm(c *gin.Context) {
	if !CheckAuthKey(c) {
		c.IndentedJSON(http.StatusForbidden, gin.H{"message": "Authentication failed"})
		return
	}

	id := c.Param("id")
	filmIdObj, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Id is invalid"})
		return
	}

	film := FindFilm(bson.M{"_id": filmIdObj})

	for _, director := range film.Directors {
		director := director
		go func() {
			_, err := RemoveFilmsFromDirector(director, []string{id})
			if err != nil {
				panic(err)
			}
		}()
	}

	for _, role := range film.Roles {
		role := role
		go func() {
			_, err := RemoveFilmsFromActor(role.ActorId, []string{id})
			if err != nil {
				panic(err)
			}
		}()
	}

	result := DeleteItemById(filmColl, id)

	c.IndentedJSON(http.StatusNoContent, result)
}

func GetFilmById(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))

	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "Id is invalid"})
		return
	}

	film := FindFilm(bson.M{"_id": id})

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
	if !CheckAuthKey(c) {
		c.IndentedJSON(http.StatusForbidden, gin.H{"message": "Authentication failed"})
		return
	}

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

	if _, err := UpdateFilmById(c.Param("id"), Film{Roles: req.Roles}); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusNoContent, gin.H{})
}

type UpdateDirectorsReq struct {
	Replace   bool     `json:"replace"`
	Directors []string `json:"directors"`
}

func UpdateDirectors(c *gin.Context) {
	if !CheckAuthKey(c) {
		c.IndentedJSON(http.StatusForbidden, gin.H{"message": "Authentication failed"})
		return
	}

	var req UpdateDirectorsReq

	if err := c.BindJSON(&req); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": "JSON is invalid"})
		return
	}

	if valid, msg := AreDirectorsIdsValid(req.Directors); !valid {
		c.IndentedJSON(http.StatusBadRequest, msg)
		return
	}

	for _, director := range req.Directors {
		director := director
		go func() {
			_, err := AddFilmsToDirector(director, []string{c.Param("id")})
			if err != nil {
				panic(err)
			}
		}()
	}

	if _, err := UpdateFilmById(c.Param("id"), Film{Directors: req.Directors}); err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"message": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusNoContent, gin.H{})
}
