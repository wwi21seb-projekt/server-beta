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

	validator := utils.NewValidator()
	mailService := services.NewMailService()
	userService := services.NewUserService(userRepo, activationTokenRepo, mailService, validator, postRepo)
	postService := services.NewPostService(postRepo, userRepo, hashtagRepo)

	imprintController := controllers.NewImprintController()
	userController := controllers.NewUserController(userService)
	postController := controllers.NewPostController(postService)

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
	api.PUT("/users", middleware.AuthorizeUser, userController.UpdateUserInformation)
	api.PATCH("/users", middleware.AuthorizeUser, userController.ChangeUserPassword)
	api.GET("/users/:username", middleware.AuthorizeUser, userController.GetUserProfile)

	// Post
	api.POST("/posts", middleware.AuthorizeUser, postController.CreatePost)

	return r
}
