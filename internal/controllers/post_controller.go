package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/wwi21seb-projekt/server-beta/internal/customerrors"
	"github.com/wwi21seb-projekt/server-beta/internal/models"
	"github.com/wwi21seb-projekt/server-beta/internal/services"
	"net/http"
)

type PostControllerInterface interface {
	CreatePost(c *gin.Context)
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
		return
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
		return
	}

	// If ContentType is neither application/json nor multipart/form-data, return bad request
	c.JSON(http.StatusBadRequest, gin.H{
		"error": customerrors.BadRequest,
	})
}

func (controller *PostController) DeletePost(c *gin.Context) {
	postId := c.Param("postId")

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

	c.Status(httpStatus)
}
