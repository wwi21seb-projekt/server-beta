package router

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/marcbudd/server-beta/internal/controllers"
	"github.com/marcbudd/server-beta/internal/middleware"
	"net/http"
	"os"
)

// SetupRouter can be used to configure the router: CORS, routes, etc.
func SetupRouter() *gin.Engine {
	r := gin.Default()

	// Set CORS
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE"}
	config.AllowHeaders = []string{"*"}
	config.AllowCredentials = true
	r.Use(cors.New(config))

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
  
	// User
	api.POST("/users", controllers.CreateUser)
	api.POST("/users/login", controllers.Login)
	api.POST("/users/:username/activate", controllers.VerifyUser)
	api.DELETE("/users/:username/activate", controllers.ResendCode)
	api.GET("/users/validate", middleware.AuthorizeUser, controllers.ValidateLogin)

	return r
}
