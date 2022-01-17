package main

import (
	"filmflix/db_connection"
	"filmflix/film_api"
	"filmflix/static_serve"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"os"
)

func main() {
	_ = godotenv.Load(".env")

	dbClient := db_connection.ConnectToDB()
	defer db_connection.DisconnectFromDB(dbClient)

	router := gin.Default()
	router.Use(func(context *gin.Context) {
		context.Header("Access-Control-Allow-Origin", "*")
		context.Next()
	})

	static_serve.InitStaticRoutes(router)

	apiRoutes := router.Group("/api")
	film_api.InitFilmApiRoutes(apiRoutes, dbClient)
	film_api.InitActorApiRoutes(apiRoutes, dbClient)
	film_api.InitDirectorApiRoutes(apiRoutes, dbClient)
	err := router.Run(":" + os.Getenv("PORT"))
	if err != nil {
		return
	}
}
