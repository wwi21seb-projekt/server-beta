package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/marcbudd/server-beta/internal/models"
	"github.com/marcbudd/server-beta/internal/services"
	"github.com/marcbudd/server-beta/internal/utils"
	"net/http"
)

// CreateUser creates a new user
func CreateUser(c *gin.Context) {
	// Read body
	var userCreateRequestDTO models.UserCreateRequestDTO

	if c.Bind(&userCreateRequestDTO) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "failed to read body",
		})
		return
	}

	// Create user
	user, serviceErr := services.CreateUser(userCreateRequestDTO)
	if serviceErr != nil {
		c.JSON(serviceErr.HTTPStatusCode(), gin.H{
			"error": serviceErr.Error(),
		})
		return
	}

	// Generate a jwt token
	tokenString, err := utils.GenerateJWTToken(user.Username)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create token",
		})

		return
	}

	// Respond
	c.SetSameSite(http.SameSiteNoneMode)
	c.SetCookie("Authorization", tokenString, 3600, "/", "", true, true)
	c.JSON(http.StatusCreated, user)

}

// Login validates password of user and creates a jwt token
func Login(c *gin.Context) {

	// Read body
	var userLoginRequestDTO models.UserLoginRequestDTO

	if c.Bind(&userLoginRequestDTO) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read body",
		})
		return
	}

	// Look up requested user
	user, serviceErr := services.LoginUser(userLoginRequestDTO)
	if serviceErr != nil {
		c.JSON(serviceErr.HTTPStatusCode(), gin.H{
			"error": serviceErr.Error(),
		})
	}

	// Generate a jwt token
	tokenString, err := utils.GenerateJWTToken(user.Username)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to create token",
		})
		return
	}

	// Respond
	c.SetSameSite(http.SameSiteNoneMode)
	c.SetCookie("Authorization", tokenString, 3600*24, "/", "", true, true)
	c.JSON(http.StatusNoContent, gin.H{})

}
