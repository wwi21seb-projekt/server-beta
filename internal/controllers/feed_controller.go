package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/wwi21seb-projekt/server-beta/internal/customerrors"
	"github.com/wwi21seb-projekt/server-beta/internal/middleware"
	"github.com/wwi21seb-projekt/server-beta/internal/services"
	"net/http"
	"strconv"
)

type FeedControllerInterface interface {
	GetPostsByUserUsername(c *gin.Context)
	GetPostFeed(c *gin.Context)
	GetPostsByHashtag(c *gin.Context)
}

type FeedController struct {
	feedService services.FeedServiceInterface
}

// NewFeedController can be used as a constructor to create a FeedController "object"
func NewFeedController(feedService services.FeedServiceInterface) *FeedController {
	return &FeedController{feedService: feedService}
}

func (controller *FeedController) GetPostsByUserUsername(c *gin.Context) {
	// Read parameters from url
	username := c.Param("username")
	offsetQuery := c.DefaultQuery("offset", "0")
	limitQuery := c.DefaultQuery("limit", "10")

	offset, err := strconv.Atoi(offsetQuery)
	if err != nil {
		offset = 0
	}
	limit, err := strconv.Atoi(limitQuery)
	if err != nil {
		limit = 10
	}

	// Check if user is logged in
	currentUsername, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": customerrors.UserUnauthorized,
		})
		return
	}

	// Get posts by username
	feedDto, serviceErr, httpStatus := controller.feedService.GetPostsByUsername(username, offset, limit, currentUsername.(string))
	if serviceErr != nil {
		c.JSON(httpStatus, gin.H{
			"error": serviceErr,
		})
		return
	}

	c.JSON(httpStatus, feedDto)
}

// GetPostFeed is a controller function that gets a global or personal post feed and can be called from router
func (controller *FeedController) GetPostFeed(c *gin.Context) {
	// Read query parameters for lastPostId, limit and feedType
	lastPostId := c.DefaultQuery("postId", "")
	limitStr := c.DefaultQuery("limit", "10")
	feedType := c.DefaultQuery("feedType", "global")

	if feedType != "global" && feedType != "personal" {
		feedType = "global"
	}

	var limit int
	var err error

	// convert limit from string to int value
	limit, err = strconv.Atoi(limitStr)
	if err != nil {
		limit = 10
	}

	// Get username from request using middleware function
	username, ok := middleware.GetLoggedInUsername(c)

	// If feed type is set to global, get Global GeneralFeedDTO
	if feedType == "global" {
		postFeed, serviceErr, httpStatus := controller.feedService.GetPostsGlobalFeed(lastPostId, limit, username)
		if serviceErr != nil {
			c.JSON(httpStatus, gin.H{
				"error": serviceErr,
			})
			return
		}
		c.JSON(httpStatus, postFeed)
		return
	}

	// If feed type is set to personal, but user is not logged in, return error
	if feedType == "personal" && !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": customerrors.UserUnauthorized,
		})
		return
	}

	// Else: if user is logged in and feed type is set to personal, get Personal GeneralFeedDTO
	postFeed, serviceErr, httpStatus := controller.feedService.GetPostsPersonalFeed(username, lastPostId, limit, username)
	if serviceErr != nil {
		c.JSON(httpStatus, gin.H{
			"error": serviceErr,
		})
		return
	}
	c.JSON(httpStatus, postFeed)
}

// GetPostsByHashtag is a controller function that gets posts by hashtag and can be called from router
func (controller *FeedController) GetPostsByHashtag(c *gin.Context) {
	// Check if user is logged in
	currentUsername, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": customerrors.UserUnauthorized,
		})
		return
	}

	// Read parameters from url
	hashtag := c.DefaultQuery("q", "")
	lastPostId := c.DefaultQuery("postId", "")
	limitQuery := c.DefaultQuery("limit", "10")

	limit, err := strconv.Atoi(limitQuery)
	if err != nil {
		limit = 10
	}

	feedDto, serviceErr, httpStatus := controller.feedService.GetPostsByHashtag(hashtag, lastPostId, limit, currentUsername.(string))
	if serviceErr != nil {
		c.JSON(httpStatus, gin.H{
			"error": serviceErr,
		})
		return
	}

	c.JSON(httpStatus, feedDto)
}
