package main

import (
	"filmflix/film_api"
	"filmflix/static_serve"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"os"
)

func main() {
	_ = godotenv.Load(".env")

	router := gin.Default()
	static_serve.InitStaticRoutes(router)

	film_api.InitFilmApiRoutes(router)
	defer film_api.CloseFilmApiRoutes()

	err := router.Run(":" + os.Getenv("PORT"))
	if err != nil {
		return
	}
}
