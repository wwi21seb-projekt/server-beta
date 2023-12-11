package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/marcbudd/server-beta/internal/services"
	"net/http"
)

func ActivateUser(context *gin.Context) {

	// Read body
	var activateRequestDTO struct {
		Token string `json:"token" binding:"required"`
	}

	if context.Bind(&activateRequestDTO) != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read body",
		})
		return
	}

	// Read username from url
	username := context.Param("username")
	if username == "" {
		context.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read username from url",
		})
		return
	}

	// Activate user
	serviceErr := services.VerifyUser(username, activateRequestDTO.Token)
	if serviceErr != nil {
		context.JSON(serviceErr.HTTPStatusCode(), gin.H{
			"error": serviceErr.Error(),
		})
		return
	}

	context.JSON(http.StatusNoContent, gin.H{})
}

func ResendCode(context *gin.Context) {
	// Read username from url
	username := context.Param("username")
	if username == "" {
		context.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read username from url",
		})
		return
	}

	// Resend code
	serviceErr := services.ResendVerificationCode(username)
	if serviceErr != nil {
		context.JSON(serviceErr.HTTPStatusCode(), gin.H{
			"error": serviceErr.Error(),
		})
		return
	}

	context.JSON(http.StatusNoContent, gin.H{})
}
