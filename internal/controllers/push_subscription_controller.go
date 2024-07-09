package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/wwi21seb-projekt/server-beta/internal/customerrors"
	"github.com/wwi21seb-projekt/server-beta/internal/models"
	"github.com/wwi21seb-projekt/server-beta/internal/services"
	"net/http"
)

type PushSubscriptionControllerInterface interface {
	GetVapidKey(c *gin.Context)
	CreatePushSubscription(c *gin.Context)
}

type PushSubscriptionController struct {
	pushSubscriptionService services.PushSubscriptionServiceInterface
}

// NewPushSubscriptionController can be used as a constructor to create a PushSubscriptionController "object"
func NewPushSubscriptionController(pushSubscriptionService services.PushSubscriptionServiceInterface) *PushSubscriptionController {
	return &PushSubscriptionController{pushSubscriptionService: pushSubscriptionService}
}

// GetVapidKey is a controller function that returns a VAPID key for clients to register for push notifications
func (controller *PushSubscriptionController) GetVapidKey(c *gin.Context) {
	// Check if user is authorized
	_, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": customerrors.Unauthorized,
		})
		return
	}

	// Get token object from service
	vapidToken, serviceErr, httpStatus := controller.pushSubscriptionService.GetVapidKey()
	if serviceErr != nil {
		c.JSON(httpStatus, gin.H{
			"error": serviceErr,
		})
		return
	}

	c.JSON(httpStatus, vapidToken)
	return
}

// CreatePushSubscription is a controller function that saves a new push subscription key to the database to send notifications to the client
func (controller *PushSubscriptionController) CreatePushSubscription(c *gin.Context) {
	// Get username from request that was set in middleware
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, customerrors.Unauthorized)
		return
	}

	// Read body
	var dto models.PushSubscriptionRequestDTO
	if c.ShouldBindJSON(&dto) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": customerrors.BadRequest,
		})
		return
	}

	// Create push subscription
	responseDto, serviceErr, httpStatus := controller.pushSubscriptionService.CreatePushSubscription(&dto, username.(string))
	if serviceErr != nil {
		c.JSON(httpStatus, gin.H{
			"error": serviceErr,
		})
		return
	}

	c.JSON(httpStatus, responseDto)
	return
}
