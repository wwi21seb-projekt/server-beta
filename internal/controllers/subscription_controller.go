package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/wwi21seb-projekt/server-beta/internal/customerrors"
	"github.com/wwi21seb-projekt/server-beta/internal/models"
	"github.com/wwi21seb-projekt/server-beta/internal/services"
	"net/http"
	"strconv"
)

type SubscriptionControllerInterface interface {
	PostSubscription(c *gin.Context)
	DeleteSubscription(c *gin.Context)
	GetSubscriptions(c *gin.Context)
}

type SubscriptionController struct {
	subscriptionService services.SubscriptionServiceInterface
}

func NewSubscriptionController(subscriptionService services.SubscriptionServiceInterface) *SubscriptionController {
	return &SubscriptionController{subscriptionService: subscriptionService}
}

func (controller *SubscriptionController) PostSubscription(c *gin.Context) {

	// Get current user from middleware
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": customerrors.Unauthorized,
		})
		return
	}

	var req models.SubscriptionPostRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": customerrors.BadRequest,
		})
		return
	}

	// Create subscription
	response, serviceErr, httpStatus := controller.subscriptionService.PostSubscription(&req, username.(string))
	if serviceErr != nil {
		c.JSON(httpStatus, gin.H{
			"error": serviceErr,
		})
		return
	}

	c.JSON(httpStatus, response)
}

func (controller *SubscriptionController) DeleteSubscription(c *gin.Context) {

	subscriptionId := c.Param("subscriptionId")

	// Get current user from middleware
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": customerrors.Unauthorized,
		})
		return
	}

	// Delete subscription
	serviceErr, httpStatus := controller.subscriptionService.DeleteSubscription(subscriptionId, username.(string))
	if serviceErr != nil {
		c.JSON(httpStatus, gin.H{
			"error": serviceErr,
		})
		return
	}

	c.JSON(httpStatus, gin.H{})
}

func (controller *SubscriptionController) GetSubscriptions(c *gin.Context) {

	// Read information from url
	queryType := c.DefaultQuery("type", "followers")
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")
	username := c.Param("username")

	// Convert limit and offset to int
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 10
	}
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		offset = 0
	}

	if queryType != "followers" && queryType != "following" {
		queryType = "followers" // set default value
	}

	// Get current user from middleware
	currentUsername, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": customerrors.Unauthorized,
		})
		return
	}

	// Search user
	subscriptionsDto, serviceErr, httpStatus := controller.subscriptionService.GetSubscriptions(queryType, limit, offset, username, currentUsername.(string))
	if serviceErr != nil {
		c.JSON(httpStatus, gin.H{
			"error": serviceErr,
		})
		return
	}
	c.JSON(httpStatus, subscriptionsDto)
}
