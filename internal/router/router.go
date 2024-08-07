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
		c.JSON(http.StatusNotFound, gin.H{
			"error": customerrors.EndpointNotFound,
		})
	})

	// Setup repositories, services, controllers
	activationTokenRepo := repositories.NewActivationTokenRepository(initializers.DB)
	userRepo := repositories.NewUserRepository(initializers.DB)
	postRepo := repositories.NewPostRepository(initializers.DB)
	commentRepo := repositories.NewCommentRepository(initializers.DB)
	hashtagRepo := repositories.NewHashtagRepository(initializers.DB)
	subscriptionRepo := repositories.NewSubscriptionRepository(initializers.DB)
	likeRepo := repositories.NewLikeRepository(initializers.DB)
	notificationRepo := repositories.NewNotificationRepository(initializers.DB)
	pushSubscriptionRepo := repositories.NewPushSubscriptionRepository(initializers.DB)
	passwordResetRepo := repositories.NewPasswordResetRepository(initializers.DB)
	chatRepo := repositories.NewChatRepository(initializers.DB)
	messageRepo := repositories.NewMessageRepository(initializers.DB)
	imageRepo := repositories.NewImageRepository(initializers.DB)

	validator := utils.NewValidator()
	mailService := services.NewMailService()
	imageService := services.NewImageService(imageRepo)
	userService := services.NewUserService(userRepo, activationTokenRepo, mailService, validator, postRepo, imageRepo, subscriptionRepo)
	feedService := services.NewFeedService(postRepo, userRepo, likeRepo, commentRepo)
	likeService := services.NewLikeService(likeRepo, postRepo)
	pushSubscriptionService := services.NewPushSubscriptionService(pushSubscriptionRepo)
	notificationService := services.NewNotificationService(notificationRepo, pushSubscriptionService)
	subscriptionService := services.NewSubscriptionService(subscriptionRepo, userRepo, notificationService)
	commentService := services.NewCommentService(commentRepo, postRepo, userRepo)
	postService := services.NewPostService(postRepo, userRepo, hashtagRepo, validator, likeRepo, commentRepo, notificationService)
	passwordResetService := services.NewPasswordResetService(userRepo, passwordResetRepo, mailService, validator)
	chatService := services.NewChatService(chatRepo, userRepo, notificationService)
	messageService := services.NewMessageService(messageRepo, chatRepo, notificationService)

	imprintController := controllers.NewImprintController()
	userController := controllers.NewUserController(userService)
	postController := controllers.NewPostController(postService)
	feedController := controllers.NewFeedController(feedService)
	imageController := controllers.NewImageController(imageService)
	likeController := controllers.NewLikeController(likeService)
	chatController := controllers.NewChatController(chatService)
	messageController := controllers.NewMessageController(messageService)
	passwordResetController := controllers.NewPasswordResetController(passwordResetService)
	notificationController := controllers.NewNotificationController(notificationService)
	pushSubscriptionController := controllers.NewPushSubscriptionController(pushSubscriptionService)
	commentController := controllers.NewCommentController(commentService)
	subscriptionController := controllers.NewSubscriptionController(subscriptionService)

	// Empty route for standard information
	r.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"apiName": "Server Beta",
		})
	})

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
	api.GET("/users/:username/feed", middleware.AuthorizeUser, feedController.GetPostsByUserUsername)

	// Post
	api.POST("/posts", middleware.AuthorizeUser, postController.CreatePost)
	api.DELETE("/posts/:postId", middleware.AuthorizeUser, postController.DeletePost)
	api.GET("/feed", feedController.GetPostFeed)
	api.GET("/posts", middleware.AuthorizeUser, feedController.GetPostsByHashtag)

	// Picture
	api.GET("/images/:imageId", imageController.GetImageById)

	// Subscription
	api.POST("/subscriptions", middleware.AuthorizeUser, subscriptionController.PostSubscription)
	api.DELETE("/subscriptions/:subscriptionId", middleware.AuthorizeUser, subscriptionController.DeleteSubscription)
	api.GET("/subscriptions/:username", middleware.AuthorizeUser, subscriptionController.GetSubscriptions)

	// Like
	api.POST("/posts/:postId/likes", middleware.AuthorizeUser, likeController.PostLike)
	api.DELETE("/posts/:postId/likes", middleware.AuthorizeUser, likeController.DeleteLike)

	// Comment
	api.POST("/posts/:postId/comments", middleware.AuthorizeUser, commentController.CreateComment)
	api.GET("/posts/:postId/comments", middleware.AuthorizeUser, commentController.GetCommentsByPostId)

	// Notification
	api.GET("/notifications", middleware.AuthorizeUser, notificationController.GetNotifications)
	api.DELETE("/notifications/:notificationId", middleware.AuthorizeUser, notificationController.DeleteNotificationById)

	// Push subscription (for web or mobile push notifications)
	api.GET("/push/vapid", middleware.AuthorizeUser, pushSubscriptionController.GetVapidKey)
	api.POST("/push/register", middleware.AuthorizeUser, pushSubscriptionController.CreatePushSubscription)

	// Chat
	api.POST("/chats", middleware.AuthorizeUser, chatController.CreateChat)
	api.GET("/chats", middleware.AuthorizeUser, chatController.GetChats)
	api.GET("/chats/:chatId", middleware.AuthorizeUser, messageController.GetMessagesByChatId)
	api.GET("/chat", messageController.HandleWebSocket) // Websocket endpoint

	// Reset Password
	api.POST("/users/:username/reset-password", passwordResetController.InitiatePasswordReset)
	api.PATCH("/users/:username/reset-password", passwordResetController.ResetPassword)

	return r
}

// ReturnNotImplemented is a helper function to return a 501 Not Implemented error to the client
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
