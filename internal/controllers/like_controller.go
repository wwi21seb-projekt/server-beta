package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/wwi21seb-projekt/server-beta/internal/customerrors"
	"github.com/wwi21seb-projekt/server-beta/internal/models"
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

func (controller *LikeController) PostLike(c *gin.Context) {

	// Get current user from middleware
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": customerrors.UserUnauthorized,
		})
		return
	}

	var req models.LikePostRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": customerrors.BadRequest,
		})
		return
	}

	// Create like
	response, serviceErr, httpStatus := controller.likeService.PostLike(&req, username.(string))
	if serviceErr != nil {
		c.JSON(httpStatus, gin.H{
			"error": serviceErr,
		})
		return
	}

	c.JSON(httpStatus, response)
}

func (controller *LikeController) DeleteLike(c *gin.Context) {

	likeId := c.Param("likeId")

	// Get current user from middleware
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": customerrors.UserUnauthorized,
		})
		return
	}

	// Delete like
	serviceErr, httpStatus := controller.likeService.DeleteLike(likeId, username.(string))
	if serviceErr != nil {
		c.JSON(httpStatus, gin.H{
			"error": serviceErr,
		})
		return
	}

	c.JSON(httpStatus, gin.H{})
}
