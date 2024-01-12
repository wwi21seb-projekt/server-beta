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

func (controller *PostController) GetPostFeed(c *gin.Context) {
	// Read query parameters for lastPostId and limit
	lastPostId := c.DefaultQuery("lastPostId", "")
	limitStr := c.DefaultQuery("limit", "0")
	feedType := c.DefaultQuery("feedType", "global")

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

	// Get username from request that was set in middleware
	_, ok := middleware.GetLoggedInUsername(c)

	// If feed type is set to global, get Global PostFeed
	if feedType == "global" {
		// Get Global PostFeed
		postFeed, serviceErr, httpStatus := controller.postService.GetPostsGlobalFeed(lastPostId, limit)
		if serviceErr != nil {
			c.JSON(httpStatus, gin.H{
				"error": serviceErr,
			})
			return
		}

		c.JSON(http.StatusOK, postFeed)
	}

	// If feed type is set to personal, but user is not logged in, return error
	if feedType == "personal" && !ok {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": customerrors.PreliminaryUserUnauthorized,
		})
	}

	// Else: if user is logged in and feed type is set to personal, get Personal PostFeed
	// TODO: add personal feed
}
