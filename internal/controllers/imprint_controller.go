package controllers

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// GetImprint returns imprint information
func GetImprint(c *gin.Context) {

	responseBody := struct {
		Text string `json:"text"`
	}{
		Text: "Das ist das Impressum...",
	}

	// Respond
	c.JSON(http.StatusOK, responseBody)
}
