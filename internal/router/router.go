package router

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/marcbudd/server-beta/internal/controllers"
	"github.com/marcbudd/server-beta/internal/initializers"
	"github.com/marcbudd/server-beta/internal/middleware"
	"github.com/marcbudd/server-beta/internal/repositories"
	"github.com/marcbudd/server-beta/internal/services"
	"github.com/marcbudd/server-beta/internal/utils"
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

	// Setup repositories, services, controllers
	activationTokenRepo := repositories.NewActivationTokenRepository(initializers.DB)
	userRepo := repositories.NewUserRepository(initializers.DB)
	postRepo := repositories.NewPostRepository(initializers.DB)
	hashtagRepo := repositories.NewHashtagRepository(initializers.DB)
	fileSystem := repositories.NewFileSystem()

	validator := utils.NewValidator()
	mailService := services.NewMailService()
	imageService := services.NewImageService(fileSystem)
	userService := services.NewUserService(userRepo, activationTokenRepo, mailService, validator)
	postService := services.NewPostService(postRepo, userRepo, hashtagRepo, imageService)

	imprintController := controllers.NewImprintController()
	userController := controllers.NewUserController(userService)
	postController := controllers.NewPostController(postService)
	imageController := controllers.NewImageController(imageService)

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
	api.GET("/imprint", imprintController.GetImprint)

	// User
	api.POST("/users", userController.CreateUser)
	api.POST("/users/login", userController.Login)
	api.POST("/users/:username/activate", userController.ActivateUser)
	api.DELETE("/users/:username/activate", userController.ResendActivationToken)
	api.GET("/users/validate", middleware.AuthorizeUser, userController.ValidateLogin)
	api.GET("/users", ReturnNotImplemented)
	api.GET("/users/:username", ReturnNotImplemented)
	api.GET("/users/:username/feed", ReturnNotImplemented)
	api.PUT("/users", ReturnNotImplemented)
	api.PATCH("/users", ReturnNotImplemented)

	// Post
	api.POST("/posts", middleware.AuthorizeUser, postController.CreatePost)

	// Image
	api.GET("/images/:filename", imageController.GetImage)

  // Subscriptions
	api.POST("/subscriptions", ReturnNotImplemented)
	api.DELETE("/subscriptions", ReturnNotImplemented)

	// Feed
	api.GET("/feed", ReturnNotImplemented)

	return r
}

func ReturnNotImplemented(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Not implemented",
	})
}
