package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

	// MARC: Username aus dem Context von der Middleware holen
	// username, exists := c.Get("username")
	// if !exists {
	// 	c.JSON(http.StatusUnauthorized, gin.H{
	// 		"error": customerrors.PreliminaryUserUnauthorized,
	// 	})

	var req models.SubscriptionPostRequestDTO
	if err := c.Bind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": customerrors.BadRequest, // MARC: customerrors verwenden
		})
		return
	}

	// MARC: hier übergibst du als Username immer "exampleFollower" und "exampleFollowing"
	// du musst den request body übergeben und den usernamen den du aus der middleware hoslt
	response, serviceErr, httpStatus := controller.subscriptionService.PostSubscription("exampleFollower", "exampleFollowing")
	if serviceErr != nil {
		c.JSON(httpStatus, gin.H{
			"error": serviceErr,
		})
		return
	}

	c.JSON(httpStatus, response)
}

func (controller *SubscriptionController) DeleteSubscription(c *gin.Context) {

	// MARC: ich würde hier nicht unbedingt die Subscription ID zu UUID parsen, sondern die id als string verwende
	// und dann in der service methode das einfach als string zum filtern des requests verwenden
	// falls die subscribtion dann nicht existiert, ist das dann 404 Not Found
	// auch hier musst username des eingeloggten nutzers auslesen, an den service übergeben und überprüfen, ob der nutzer überhaupt berechtigt ist die subscription zu löschen
	subscriptionId, err := uuid.Parse(c.Param("subscriptionId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid subscription id"}) // MARC: customerrors verwenden
		return
	}

	response, serviceErr, httpStatus := controller.subscriptionService.DeleteSubscription(subscriptionId)
	if serviceErr != nil {
		c.JSON(httpStatus, gin.H{
			"error": serviceErr,
		})
		return
	}

	c.JSON(httpStatus, response) // MARC: no content ---> also kein response body
}
