package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/marcbudd/server-beta/internal/customerrors"
	"github.com/marcbudd/server-beta/internal/middleware"
	"github.com/marcbudd/server-beta/internal/models"
	"github.com/marcbudd/server-beta/internal/services"
	"net/http"
	"strconv"
)

type PostControllerInterface interface {
	CreatePost(c *gin.Context)
	GetPostsByUserUsername(c *gin.Context)
	GetPostFeed(c *gin.Context)
}

type PostController struct {
	postService services.PostServiceInterface
}

// NewPostController can be used as a constructor to create a PostController "object"
func NewPostController(postService services.PostServiceInterface) *PostController {
	return &PostController{postService: postService}
}

// CreatePost is a controller function that creates a post and can be called from router.go
func (controller *PostController) CreatePost(c *gin.Context) {
	// Get username from request that was set in middleware
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": customerrors.PreliminaryUserUnauthorized,
		})
		return
	}

	// Read body
	var postCreateRequestDTO models.PostCreateRequestDTO
	if c.Bind(&postCreateRequestDTO) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": customerrors.BadRequest,
		})
		return
	}
	// Create post
	postDto, serviceErr, httpStatus := controller.postService.CreatePost(&postCreateRequestDTO, username.(string))
	if serviceErr != nil {
		c.JSON(httpStatus, gin.H{
			"error": serviceErr,
		})
		return
	}

	c.JSON(httpStatus, postDto)
}

func (controller *PostController) GetPostsByUserUsername(c *gin.Context) {
	username := c.Param("username")
	offsetQuery := c.DefaultQuery("offset", "0")
	limitQuery := c.DefaultQuery("limit", "10")

	offset, err := strconv.Atoi(offsetQuery)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": customerrors.BadRequest,
		})
		return
	}

	limit, err := strconv.Atoi(limitQuery)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": customerrors.BadRequest,
		})
		return
	}

	feedDto, serviceErr, httpStatus := controller.postService.GetPostsByUsername(username, offset, limit)
	if serviceErr != nil {
		c.JSON(httpStatus, gin.H{
			"error": serviceErr,
		})
		return
	}

	c.JSON(httpStatus, feedDto)
}

// GetPostFeed is a controller function that gets a global or personal post feed and can be called from router
func (controller *PostController) GetPostFeed(c *gin.Context) {
	// Read query parameters for lastPostId and limit
	lastPostId := c.DefaultQuery("postId", "")
	limitStr := c.DefaultQuery("limit", "0")
	feedType := c.DefaultQuery("feedType", "global")

	if feedType != "global" && feedType != "personal" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": customerrors.BadRequest,
		})
		return
	}

	var limit int
	var err error

	// convert limit from string to int value
	limit, err = strconv.Atoi(limitStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": customerrors.BadRequest,
		})
		return
	}

	// Get username from request using middleware function
	username, ok := middleware.GetLoggedInUsername(c)

	// If feed type is set to global, get Global GeneralFeedDTO
	if feedType == "global" {
		postFeed, serviceErr, httpStatus := controller.postService.GetPostsGlobalFeed(lastPostId, limit)
		if serviceErr != nil {
			c.JSON(httpStatus, gin.H{
				"error": serviceErr,
			})
			return
		}
		c.JSON(http.StatusOK, postFeed)
		return
	}

	// If feed type is set to personal, but user is not logged in, return error
	if feedType == "personal" && !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": customerrors.PreliminaryUserUnauthorized,
		})
		return
	}

	// Else: if user is logged in and feed type is set to personal, get Personal GeneralFeedDTO
	postFeed, serviceErr, httpStatus := controller.postService.GetPostsPersonalFeed(username, lastPostId, limit)
	if serviceErr != nil {
		c.JSON(httpStatus, gin.H{
			"error": serviceErr,
		})
		return
	}
	c.JSON(http.StatusOK, postFeed)
}
