package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/wwi21seb-projekt/server-beta/internal/customerrors"
	"github.com/wwi21seb-projekt/server-beta/internal/services"
	"net/http"
)

type NotificationControllerInterface interface {
	GetNotifications(c *gin.Context)
	DeleteNotificationById(c *gin.Context)
}

type NotificationController struct {
	notificationService services.NotificationServiceInterface
}

// NewNotificationController can be used as a constructor to create a NotificationController "object"
func NewNotificationController(notificationService services.NotificationServiceInterface) *NotificationController {
	return &NotificationController{notificationService: notificationService}
}

// GetNotifications is a controller function that gets all notifications for the curent user and can be called from router.go
func (controller *NotificationController) GetNotifications(c *gin.Context) {
	// Get username from request that was set in middleware
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": customerrors.UserUnauthorized,
		})
		return
	}

	// Get notifications
	notifications, serviceErr, httpStatus := controller.notificationService.GetNotifications(username.(string))
	if serviceErr != nil {
		c.JSON(httpStatus, gin.H{
			"error": serviceErr,
		})
		return
	}

	c.JSON(http.StatusOK, notifications)
	return
}

// DeleteNotificationById is a controller function that deletes a notification by its id and can be called from router.go
func (controller *NotificationController) DeleteNotificationById(c *gin.Context) {
	// Get username from request that was set in middleware
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": customerrors.UserUnauthorized,
		})
		return
	}

	// Get notificationId from request
	notificationId := c.Param("notificationId")

	// Delete notification
	serviceErr, httpStatus := controller.notificationService.DeleteNotificationById(notificationId, username.(string))
	if serviceErr != nil {
		c.JSON(httpStatus, gin.H{
			"error": serviceErr,
		})
		return
	}

	c.JSON(httpStatus, gin.H{})
	return
}
