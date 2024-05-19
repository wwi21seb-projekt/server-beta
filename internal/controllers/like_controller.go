package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/wwi21seb-projekt/server-beta/internal/customerrors"
	"github.com/wwi21seb-projekt/server-beta/internal/services"
	"net/http"
)

type LikeControllerInterface interface {
	PostLike(c *gin.Context)
	DeleteLike(c *gin.Context)
}

type LikeController struct {
	likeService services.LikeServiceInterface
}

func NewLikeController(likeService services.LikeServiceInterface) *LikeController {
	return &LikeController{likeService: likeService}
}

// PostLike creates a like for a given post id and the current logged-in user
func (controller *LikeController) PostLike(c *gin.Context) {

	// Get current user from middleware
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": customerrors.UserUnauthorized,
		})
		return
	}

	// Read post id from request
	postId := c.Param("postId")

	// Create like
	serviceErr, httpStatus := controller.likeService.PostLike(postId, username.(string))
	if serviceErr != nil {
		c.JSON(httpStatus, gin.H{
			"error": serviceErr,
		})
		return
	}

	c.JSON(httpStatus, gin.H{})
}

// DeleteLike deletes a like for a given post id and the current logged-in user
func (controller *LikeController) DeleteLike(c *gin.Context) {
	// Get current user from middleware
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": customerrors.UserUnauthorized,
		})
		return
	}

	// Read post id from request
	postId := c.Param("postId")

	// Delete like
	serviceErr, httpStatus := controller.likeService.DeleteLike(postId, username.(string))
	if serviceErr != nil {
		c.JSON(httpStatus, gin.H{
			"error": serviceErr,
		})
		return
	}

	c.JSON(httpStatus, gin.H{})
}
