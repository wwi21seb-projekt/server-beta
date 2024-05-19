package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/wwi21seb-projekt/server-beta/internal/customerrors"
	"github.com/wwi21seb-projekt/server-beta/internal/models"
	"github.com/wwi21seb-projekt/server-beta/internal/services"
	"net/http"
	"strconv"
)

type CommentControllerInterface interface {
	CreateComment(c *gin.Context)
	GetCommentsByPostId(c *gin.Context)
}

type CommentController struct {
	commentService services.CommentServiceInterface
}

// NewCommentController can be used as a constructor to create a CommentController "object"
func NewCommentController(commentService services.CommentServiceInterface) *CommentController {
	return &CommentController{commentService: commentService}
}

// CreateComment is a controller function that creates a comment and can be called from router.go
func (controller *CommentController) CreateComment(c *gin.Context) {
	// Get username from request that was set in middleware
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": customerrors.UserUnauthorized,
		})
		return
	}

	// Read body
	var commentCreateRequestDTO models.CommentCreateRequestDTO
	if c.ShouldBindJSON(&commentCreateRequestDTO) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": customerrors.BadRequest,
		})
		return
	}

	// Get post id from URL
	postId := c.Param("postId")

	// Create comment
	commentDto, serviceErr, httpStatus := controller.commentService.CreateComment(&commentCreateRequestDTO, postId, username.(string))
	if serviceErr != nil {
		c.JSON(httpStatus, gin.H{
			"error": serviceErr,
		})
		return
	}

	c.JSON(httpStatus, commentDto)
}

// GetCommentsByPostId is a controller function that retrieves comments by post id and can be called from router.go
func (controller *CommentController) GetCommentsByPostId(c *gin.Context) {
	// Get pagination information
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

	// Get post id from URL
	postId := c.Param("postId")

	// Check if user is logged in
	_, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": customerrors.UserUnauthorized,
		})
		return
	}

	// Get comments by post id
	commentFeedDto, serviceErr, httpStatus := controller.commentService.GetCommentsByPostId(postId, offset, limit)
	if serviceErr != nil {
		c.JSON(httpStatus, gin.H{
			"error": serviceErr,
		})
		return
	}

	c.JSON(httpStatus, commentFeedDto)
}
