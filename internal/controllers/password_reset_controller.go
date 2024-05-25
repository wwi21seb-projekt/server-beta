package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/wwi21seb-projekt/server-beta/internal/customerrors"
	"github.com/wwi21seb-projekt/server-beta/internal/models"
	"github.com/wwi21seb-projekt/server-beta/internal/services"
	"net/http"
)

type PasswordResetControllerInterface interface {
	PasswordReset(c *gin.Context)
	SetNewPassword(c *gin.Context)
}

type PasswordResetController struct {
	passwordResetService services.PasswordResetServiceInterface
}

// NewPasswordResetController can be used as a constructor to return a new PasswordResetController "object"
func NewPasswordResetController(passwordResetService services.PasswordResetServiceInterface) *PasswordResetController {
	return &PasswordResetController{passwordResetService: passwordResetService}
}

// PasswordReset triggers a password reset process for the user
func (controller *PasswordResetController) PasswordReset(c *gin.Context) {
	// Read body
	var passwordResetRequestDTO models.PasswordResetRequestDTO

	if c.ShouldBindJSON(&passwordResetRequestDTO) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": customerrors.BadRequest,
		})
		return
	}

	// Initiate password reset
	response, serviceErr, httpStatus := controller.passwordResetService.PasswordReset(passwordResetRequestDTO.Username)
	if serviceErr != nil {
		c.JSON(httpStatus, gin.H{
			"error": serviceErr,
		})
		return
	}

	c.JSON(httpStatus, response)
}

// SetNewPassword sets a new password using the provided token
func (controller *PasswordResetController) SetNewPassword(c *gin.Context) {
	// Read body
	var setNewPasswordDTO models.SetNewPasswordDTO

	if c.ShouldBindJSON(&setNewPasswordDTO) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": customerrors.BadRequest,
		})
		return
	}

	// Read username from URL
	username := c.Param("username")

	// Set new password
	serviceErr, httpStatus := controller.passwordResetService.SetNewPassword(username, setNewPasswordDTO.Token, setNewPasswordDTO.NewPassword)
	if serviceErr != nil {
		c.JSON(httpStatus, gin.H{
			"error": serviceErr,
		})
		return
	}

	c.JSON(httpStatus, gin.H{})
}
