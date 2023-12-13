package router

import (
	"github.com/gin-gonic/gin"
	"github.com/marcbudd/server-beta/internal/controllers"
	"net/http"
	"os"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// Set trusted proxies
	err := r.SetTrustedProxies([]string{os.Getenv("PROXY_HOST")})
	if err != nil {
		panic(err)
		return nil
	}

	// API Routes
	api := r.Group("/api")

	// Example:
	api.GET("/test/:number", func(context *gin.Context) {
		// Get number from url
		number := context.Param("number")

		//Respond
		context.JSON(http.StatusOK, number)
	})

	// Imprint
	api.GET("/imprint", controllers.GetImprint)

	return r
}
