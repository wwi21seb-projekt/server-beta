package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/wwi21seb-projekt/server-beta/internal/customerrors"
	"github.com/wwi21seb-projekt/server-beta/internal/middleware"
	"github.com/wwi21seb-projekt/server-beta/internal/models"
	"github.com/wwi21seb-projekt/server-beta/internal/services"
	"net/http"
	"strconv"
)

type PostControllerInterface interface {
	CreatePost(c *gin.Context)
	GetPostsByUserUsername(c *gin.Context)
	GetPostFeed(c *gin.Context)
	DeletePost(c *gin.Context)
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
			"error": customerrors.UserUnauthorized,
		})
		return
	}

	// If ContentType is application/json, read body and continue only with text
	if c.ContentType() == "application/json" {
		// Read body
		var postCreateRequestDTO models.PostCreateRequestDTO
		if c.ShouldBindJSON(&postCreateRequestDTO) != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": customerrors.BadRequest,
			})
			return
		}

		// Create post
		postDto, serviceErr, httpStatus := controller.postService.CreatePost(&postCreateRequestDTO, nil, username.(string))
		if serviceErr != nil {
			c.JSON(httpStatus, gin.H{
				"error": serviceErr,
			})
			return
		}

		c.JSON(httpStatus, postDto)
	}

	// If ContentType is multipart/form-data, continue with image and (optional) text
	if c.ContentType() == "multipart/form-data" {
		// Read content
		content := c.PostForm("content")
		postCreateRequestDTO := models.PostCreateRequestDTO{
			Content: content,
		}

		// Read image
		file, err := c.FormFile("image")
		if err != nil {
			file = nil
		}

		// If no file is present and content is empty, return bad request
		if file == nil && content == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": customerrors.BadRequest,
			})
			return
		}

		// Create post
		postDto, serviceErr, httpStatus := controller.postService.CreatePost(&postCreateRequestDTO, file, username.(string))
		if serviceErr != nil {
			c.JSON(httpStatus, gin.H{
				"error": serviceErr,
			})
			return
		}

		c.JSON(httpStatus, postDto)

	}

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
			"error": customerrors.UserUnauthorized,
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

func (controller *PostController) DeletePost(c *gin.Context) {
	postIdStr := c.Param("postId")
	postId, err := uuid.Parse(postIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": customerrors.BadRequest,
		})
		return
	}

	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": customerrors.UserUnauthorized,
		})
		return
	}

	serviceErr, httpStatus := controller.postService.DeletePost(postId, username.(string))
	if serviceErr != nil {
		c.JSON(httpStatus, gin.H{
			"error": serviceErr,
		})
		return
	}

	c.Status(http.StatusOK)
}
