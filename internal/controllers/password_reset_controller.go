package controllers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/wwi21seb-projekt/server-beta/internal/customerrors"
	"github.com/wwi21seb-projekt/server-beta/internal/models"
	"github.com/wwi21seb-projekt/server-beta/internal/services"
	"net/http"
)

type PasswordResetControllerInterface interface {
	InitiatePasswordReset(c *gin.Context)
	ResetPassword(c *gin.Context)
}

type PasswordResetController struct {
	passwordResetService services.PasswordResetServiceInterface
}

// NewPasswordResetController can be used as a constructor to return a new PasswordResetController "object"
func NewPasswordResetController(passwordResetService services.PasswordResetServiceInterface) *PasswordResetController {
	return &PasswordResetController{passwordResetService: passwordResetService}
}

// InitiatePasswordReset triggers a password reset process for the user and can be called from the router
func (controller *PasswordResetController) InitiatePasswordReset(c *gin.Context) {
	fmt.Println("test")

	// Read username from URL
	username := c.Param("username")

	// Initiate password reset
	response, serviceErr, httpStatus := controller.passwordResetService.InitiatePasswordReset(username)
	if serviceErr != nil {
		c.JSON(httpStatus, gin.H{
			"error": serviceErr,
		})
		return
	}

	c.JSON(httpStatus, response)
}

// ResetPassword sets a new password using the provided token and can be called from the router
func (controller *PasswordResetController) ResetPassword(c *gin.Context) {
	// Read body
	var setNewPasswordDTO models.ResetPasswordRequestDTO

	if c.ShouldBindJSON(&setNewPasswordDTO) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": customerrors.BadRequest,
		})
		return
	}

	// Read username from URL
	username := c.Param("username")

	// Set new password
	serviceErr, httpStatus := controller.passwordResetService.ResetPassword(username, &setNewPasswordDTO)
	if serviceErr != nil {
		c.JSON(httpStatus, gin.H{
			"error": serviceErr,
		})
		return
	}

	c.JSON(httpStatus, gin.H{})
}
