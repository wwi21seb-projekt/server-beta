package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/marcbudd/server-beta/internal/errors"
	"github.com/marcbudd/server-beta/internal/models"
	"github.com/marcbudd/server-beta/internal/services"
	"net/http"
)

type UserControllerInterface interface {
	CreateUser(c *gin.Context)
	Login(c *gin.Context)
	ActivateUser(c *gin.Context)
	ResendActivationToken(c *gin.Context)
	ValidateLogin(c *gin.Context)
}

type UserController struct {
	userService services.UserServiceInterface
}

// NewUserController can be used as a constructor to return a new UserController "object"
func NewUserController(userService services.UserServiceInterface) *UserController {
	return &UserController{userService: userService}
}

// CreateUser creates a new user
func (controller *UserController) CreateUser(c *gin.Context) {
	// Read body
	var userCreateRequestDTO models.UserCreateRequestDTO

	if c.Bind(&userCreateRequestDTO) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errors.BadRequest,
		})
		return
	}

	// Create userDto
	userDto, serviceErr, httpStatus := controller.userService.CreateUser(userCreateRequestDTO)
	if serviceErr != nil {
		c.JSON(httpStatus, gin.H{
			"error": serviceErr,
		})
		return
	}

	c.JSON(httpStatus, userDto)

}

// Login validates password of user and creates a jwt token
func (controller *UserController) Login(c *gin.Context) {

	// Read body
	var userLoginRequestDTO models.UserLoginRequestDTO

	if c.Bind(&userLoginRequestDTO) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errors.BadRequest,
		})
		return
	}

	// Lookup requested user
	loginResponseDto, serviceErr, httpStatus := controller.userService.LoginUser(userLoginRequestDTO)
	if serviceErr != nil {
		c.JSON(httpStatus, gin.H{
			"error": serviceErr,
		})
		return
	}

	c.JSON(httpStatus, loginResponseDto)

}

// ActivateUser verifies user with given six-digit code and resends a new token, if it is expired
func (controller *UserController) ActivateUser(c *gin.Context) {

	// Read body
	var verificationTokenRequestDTO models.ActivationTokenRequestDTO

	if c.Bind(&verificationTokenRequestDTO) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errors.BadRequest,
		})
		return
	}

	// Read username from url
	username := c.Param("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errors.BadRequest,
		})
		return
	}

	// Activate user
	serviceErr, httpStatus := controller.userService.ActivateUser(username, verificationTokenRequestDTO.Token)
	if serviceErr != nil {
		c.JSON(httpStatus, gin.H{
			"error": serviceErr,
		})
		return
	}

	c.JSON(httpStatus, gin.H{})
}

// ResendActivationToken sends a new six-digit verification code to the user
func (controller *UserController) ResendActivationToken(c *gin.Context) {
	// Read username from url
	username := c.Param("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": errors.BadRequest,
		})
		return
	}

	// Resend code
	serviceErr, httpStatus := controller.userService.ResendActivationToken(username)
	if serviceErr != nil {
		c.JSON(httpStatus, gin.H{
			"error": serviceErr,
		})
		return
	}

	c.JSON(httpStatus, gin.H{})
}

// ValidateLogin is a test function to see whether the user is logged in and returns username
func (controller *UserController) ValidateLogin(c *gin.Context) {
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"username": username,
	})
}
