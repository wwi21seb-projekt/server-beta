package router

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/wwi21seb-projekt/server-beta/internal/controllers"
	"github.com/wwi21seb-projekt/server-beta/internal/customerrors"
	"github.com/wwi21seb-projekt/server-beta/internal/initializers"
	"github.com/wwi21seb-projekt/server-beta/internal/middleware"
	"github.com/wwi21seb-projekt/server-beta/internal/repositories"
	"github.com/wwi21seb-projekt/server-beta/internal/services"
	"github.com/wwi21seb-projekt/server-beta/internal/utils"
	"net/http"
	"os"
)

// SetupRouter configures the router: CORS, routes, etc.
func SetupRouter() *gin.Engine {
	r := gin.Default()

	// Set CORS
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization"}
	config.AllowCredentials = true
	r.Use(cors.New(config))

	// Set trusted proxies
	err := r.SetTrustedProxies([]string{os.Getenv("PROXY_HOST")})
	if err != nil {
		panic(err)
		return nil
	}

	// Recover from panics and return 500 Internal Server Error
	r.Use(RecoveryMiddleware())

	// No route found
	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Page not found"})
	})

	// Setup repositories, services, controllers
	activationTokenRepo := repositories.NewActivationTokenRepository(initializers.DB)
	userRepo := repositories.NewUserRepository(initializers.DB)
	postRepo := repositories.NewPostRepository(initializers.DB)
	hashtagRepo := repositories.NewHashtagRepository(initializers.DB)
	fileSystem := repositories.NewFileSystem()
	subscriptionRepo := repositories.NewSubscriptionRepository(initializers.DB)

	validator := utils.NewValidator()
	mailService := services.NewMailService()
	imageService := services.NewImageService(fileSystem)
	userService := services.NewUserService(userRepo, activationTokenRepo, mailService, validator, postRepo, subscriptionRepo)
	postService := services.NewPostService(postRepo, userRepo, hashtagRepo, imageService)
	subscriptionService := services.NewSubscriptionService(subscriptionRepo, userRepo)

	imprintController := controllers.NewImprintController()
	userController := controllers.NewUserController(userService)
	postController := controllers.NewPostController(postService)
	imageController := controllers.NewImageController(imageService)

	subscriptionController := controllers.NewSubscriptionController(subscriptionService)

	// API Routes
	api := r.Group("/api")

	// Imprint
	api.GET("/imprint", imprintController.GetImprint)

	// User
	api.POST("/users", userController.CreateUser)
	api.POST("/users/login", userController.Login)
	api.POST("/users/:username/activate", userController.ActivateUser)
	api.DELETE("/users/:username/activate", userController.ResendActivationToken)
	api.POST("/users/refresh", userController.RefreshToken)
	api.GET("/users", middleware.AuthorizeUser, userController.SearchUser)
	api.PUT("/users", middleware.AuthorizeUser, userController.UpdateUserInformation)
	api.PATCH("/users", middleware.AuthorizeUser, userController.ChangeUserPassword)
	api.GET("/users/:username", middleware.AuthorizeUser, userController.GetUserProfile)
	api.GET("/users/:username/feed", middleware.AuthorizeUser, postController.GetPostsByUserUsername)
	api.DELETE("/posts/:postId", middleware.AuthorizeUser, postController.DeletePost)

	// Post
	api.POST("/posts", middleware.AuthorizeUser, postController.CreatePost)
	api.GET("/feed", postController.GetPostFeed)
	api.GET("/posts", middleware.AuthorizeUser, postController.GetPostsByHashtag)

	// Image
	api.GET("/images/:filename", imageController.GetImage)

	// Subscription
	api.POST("/subscriptions", middleware.AuthorizeUser, subscriptionController.PostSubscription)
	api.DELETE("/subscriptions/:subscriptionId", middleware.AuthorizeUser, subscriptionController.DeleteSubscription)
	api.GET("/subscriptions/:username", middleware.AuthorizeUser, subscriptionController.GetSubscriptions)

	return r
}

// ReturnNotImplemented returns a 501 Not Implemented error and can be added to unfinished routes
func ReturnNotImplemented(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Not implemented",
	})
}

// RecoveryMiddleware recovers from panics and returns a 500 Internal Server Error
func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				// Respond with a 500 Internal Server Error status code
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": customerrors.InternalServerError,
				})
			}
		}()
		c.Next()
	}
}
