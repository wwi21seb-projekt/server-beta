package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/marcbudd/server-beta/internal/errors"
	"github.com/marcbudd/server-beta/internal/models"
	"github.com/marcbudd/server-beta/internal/services"
	"net/http"
)

// CreateUser creates a new user
func CreateUser(c *gin.Context) {
	// Read body
	var userCreateRequestDTO models.UserCreateRequestDTO

	if c.Bind(&userCreateRequestDTO) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errors.INVALID_REQ_BODY,
		})
		return
	}

	// Create userDto
	userDto, serviceErr, httpStatus := services.CreateUser(userCreateRequestDTO)
	if serviceErr != nil {
		c.JSON(httpStatus, gin.H{
			"error": serviceErr,
		})
		return
	}

	c.JSON(httpStatus, userDto)

}

// Login validates password of user and creates a jwt token
func Login(c *gin.Context) {

	// Read body
	var userLoginRequestDTO models.UserLoginRequestDTO

	if c.Bind(&userLoginRequestDTO) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errors.INVALID_REQ_BODY,
		})
		return
	}

	// Lookup requested user
	loginResponseDto, serviceErr, httpStatus := services.LoginUser(userLoginRequestDTO)
	if serviceErr != nil {
		c.JSON(httpStatus, gin.H{
			"error": serviceErr,
		})
		return
	}

	c.JSON(httpStatus, loginResponseDto)

}

// VerifyUser verifies user with given six-digit code and resends a new token, if it is expired
func VerifyUser(context *gin.Context) {

	// Read body
	var verificationTokenRequestDTO models.VerificationTokenRequestDTO

	if context.Bind(&verificationTokenRequestDTO) != nil {
		context.JSON(http.StatusBadRequest, gin.H{
			"error": errors.INVALID_REQ_BODY,
		})
		return
	}

	// Read username from url
	username := context.Param("username")
	if username == "" {
		context.JSON(http.StatusBadRequest, gin.H{
			"error": errors.INVALID_URL_PARAMETER,
		})
		return
	}

	// Activate user
	serviceErr, httpStatus := services.VerifyUser(username, verificationTokenRequestDTO.Token)
	if serviceErr != nil {
		context.JSON(httpStatus, gin.H{
			"error": serviceErr,
		})
		return
	}

	context.JSON(httpStatus, gin.H{})
}

// ResendCode sends a new six-digit verification code to the user
func ResendCode(context *gin.Context) {
	// Read username from url
	username := context.Param("username")
	if username == "" {
		context.JSON(http.StatusBadRequest, gin.H{
			"error": errors.INVALID_URL_PARAMETER,
		})
		return
	}

	// Resend code
	serviceErr, httpStatus := services.ResendVerificationToken(username)
	if serviceErr != nil {
		context.JSON(httpStatus, gin.H{
			"error": serviceErr,
		})
		return
	}

	context.JSON(httpStatus, gin.H{})
}

func ValidateLogin(context *gin.Context) {
	username, exists := context.Get("username")
	if !exists {
		context.JSON(http.StatusUnauthorized, gin.H{})
		return
	}

	context.JSON(http.StatusOK, gin.H{
		"username": username,
	})
}
