package controllers

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type ImprintControllerInterface interface {
	GetImprint(c *gin.Context)
}

type ImprintController struct {
}

// NewImprintController can be used as a constructor to return a new ImprintController "object"
func NewImprintController() *ImprintController {
	return &ImprintController{}
}

// GetImprint returns imprint information
func (controller ImprintController) GetImprint(c *gin.Context) {

	responseBody := struct {
		Text string `json:"text"`
	}{
		Text: "Das ist das Impressum...",
	}

	// Respond
	c.JSON(http.StatusOK, responseBody)
}
