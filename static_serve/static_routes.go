package static_serve

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"regexp"
)

func InitStaticRoutes(router *gin.Engine) {
	router.StaticFile("/favicon.ico", "./www/favicon.ico")
	router.Use(serveJS)
	router.Static("/static", "./www/static")
	router.GET("/", serveApp)
}

func serveApp(c *gin.Context) {
	c.Status(http.StatusOK)
	c.File("www/index.html")
}
func serveJS(c *gin.Context) {
	if matched, _ := regexp.Match(`js\z`, []byte(c.Request.RequestURI)); matched {
		c.Writer.Header().Set("Content-Type", "application/javascript")
	}
	c.Next()
}
