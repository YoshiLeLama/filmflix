package film_api

import (
	"github.com/gin-gonic/gin"
	"os"
)

func CheckAuthKey(c *gin.Context) bool {
	authKey, exists := c.GetQuery("auth")
	return exists && authKey == os.Getenv("ADMIN_KEY")
}
