package controllers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/marcbudd/server-beta/internal/customerrors"
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

	// Get current user from middleware
	username, exists := c.Get("username")
	if !exists {
		fmt.Println("test")
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": customerrors.PreliminaryUserUnauthorized,
		})
		return
	}

	var req models.SubscriptionPostRequestDTO
	if err := c.Bind(&req); err != nil {
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
	if subscriptionId == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": customerrors.BadRequest,
		})
		return
	}

	// Get current user from middleware
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": customerrors.PreliminaryUserUnauthorized,
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