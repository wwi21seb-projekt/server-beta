package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/marcbudd/server-beta/internal/models"
	"github.com/marcbudd/server-beta/internal/services"
	"net/http"
)

type SubscriptionControllerInterface interface {
	PostSubscription(c *gin.Context)
	DeleteSubscription(c *gin.Context)
}

type SubscriptionController struct {
	subscriptionService services.SubscriptionServiceInterface
}

func NewSubscriptionController(subscriptionService services.SubscriptionServiceInterface) *SubscriptionController {
	return &SubscriptionController{subscriptionService: subscriptionService}
}

func (controller *SubscriptionController) PostSubscription(c *gin.Context) {
	var req models.SubscriptionPostRequestDTO
	if err := c.Bind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// Hier Annahme, dass Follower und Following aus dem Request oder dem Context stammen
	response, serviceErr, httpStatus := controller.subscriptionService.PostSubscription("exampleFollower", "exampleFollowing")
	if serviceErr != nil {
		c.JSON(httpStatus, gin.H{"error": serviceErr})
		return
	}

	c.JSON(httpStatus, response)
}

func (controller *SubscriptionController) DeleteSubscription(c *gin.Context) {
	subscriptionId, err := uuid.Parse(c.Param("subscriptionId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid subscription id"})
		return
	}

	response, serviceErr, httpStatus := controller.subscriptionService.DeleteSubscription(subscriptionId)
	if serviceErr != nil {
		c.JSON(httpStatus, gin.H{"error": serviceErr})
		return
	}

	c.JSON(httpStatus, response)
}
