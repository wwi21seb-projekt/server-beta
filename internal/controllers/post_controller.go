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
			"error": customerrors.Unauthorized,
		})
		return
	}

	var postCreateRequestDTO models.PostCreateRequestDTO
	if c.ShouldBindJSON(&postCreateRequestDTO) != nil {
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
	return

}

func (controller *PostController) DeletePost(c *gin.Context) {
	postId := c.Param("postId")

	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": customerrors.Unauthorized,
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
