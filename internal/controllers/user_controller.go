package controllers

import (
	"github.com/gin-gonic/gin"
	"github.com/wwi21seb-projekt/server-beta/internal/customerrors"
	"github.com/wwi21seb-projekt/server-beta/internal/models"
	"github.com/wwi21seb-projekt/server-beta/internal/services"
	"net/http"
	"strconv"
)

type UserControllerInterface interface {
	CreateUser(c *gin.Context)
	Login(c *gin.Context)
	ActivateUser(c *gin.Context)
	ResendActivationToken(c *gin.Context)
	RefreshToken(c *gin.Context)
	SearchUser(c *gin.Context)
	UpdateUserInformation(c *gin.Context)
	ChangeUserPassword(c *gin.Context)
	GetUserProfile(c *gin.Context)
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

	if c.ShouldBindJSON(&userCreateRequestDTO) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": customerrors.BadRequest,
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

	if c.ShouldBindJSON(&userLoginRequestDTO) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": customerrors.BadRequest,
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
	var verificationTokenRequestDTO models.UserActivationRequestDTO

	if c.ShouldBindJSON(&verificationTokenRequestDTO) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": customerrors.BadRequest,
		})
		return
	}

	// Read username from url
	username := c.Param("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": customerrors.BadRequest,
		})
		return
	}

	// Activate user
	loginResponse, serviceErr, httpStatus := controller.userService.ActivateUser(username, verificationTokenRequestDTO.Token)
	if serviceErr != nil {
		c.JSON(httpStatus, gin.H{
			"error": serviceErr,
		})
		return
	}

	c.JSON(httpStatus, loginResponse)
}

// ResendActivationToken sends a new six-digit verification code to the user
func (controller *UserController) ResendActivationToken(c *gin.Context) {
	// Read username from url
	username := c.Param("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": customerrors.BadRequest,
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

// RefreshToken creates a new jwt token with the refresh token
func (controller *UserController) RefreshToken(c *gin.Context) {
	// Read body
	var refreshTokenRequestDTO models.UserRefreshTokenRequestDTO
	if c.ShouldBindJSON(&refreshTokenRequestDTO) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": customerrors.BadRequest,
		})
		return
	}

	// Resend code
	loginResponse, serviceErr, httpStatus := controller.userService.RefreshToken(&refreshTokenRequestDTO)
	if serviceErr != nil {
		c.JSON(httpStatus, gin.H{
			"error": serviceErr,
		})
		return
	}

	c.JSON(httpStatus, loginResponse)
}

// SearchUser searches for a user by username given in url
func (controller *UserController) SearchUser(c *gin.Context) {
	// Read information from url
	username := c.DefaultQuery("username", "")
	limitStr := c.DefaultQuery("limit", "0")
	offsetStr := c.DefaultQuery("offset", "0")

	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": customerrors.BadRequest,
		})
		return
	}

	// Convert limit and offset to int
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": customerrors.BadRequest,
		})
		return
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": customerrors.BadRequest,
		})
		return
	}

	// Current username from middleware
	currentUsername, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": customerrors.UserUnauthorized,
		})
		return
	}

	// Search user
	userDto, serviceErr, httpStatus := controller.userService.SearchUser(username, limit, offset, currentUsername.(string))
	if serviceErr != nil {
		c.JSON(httpStatus, gin.H{
			"error": serviceErr,
		})
		return
	}
	c.JSON(httpStatus, userDto)
}

// UpdateUserInformation updates the user's nickname and status
func (controller *UserController) UpdateUserInformation(c *gin.Context) {
	// Extract the username from the context
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": customerrors.UserUnauthorized,
		})
		return
	}

	// Bind the JSON request body to the struct
	var userUpdateResponseDTO models.UserInformationUpdateDTO
	if err := c.ShouldBindJSON(&userUpdateResponseDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": customerrors.BadRequest,
		})
		return
	}

	// Update the user's information
	responseDTO, customErr, status := controller.userService.UpdateUserInformation(&userUpdateResponseDTO, username.(string))
	if customErr != nil {
		c.JSON(status, gin.H{
			"error": customErr,
		})
		return
	}

	c.JSON(status, responseDTO)
}

// ChangeUserPassword changes the user's password
func (controller *UserController) ChangeUserPassword(c *gin.Context) {
	// Extract the username from the context
	username, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": customerrors.UserUnauthorized,
		})
		return
	}

	// Bind the JSON request body to the struct
	var userPasswordChangeDTO models.ChangePasswordDTO
	if err := c.ShouldBindJSON(&userPasswordChangeDTO); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": customerrors.BadRequest,
		})
		return
	}

	// Update the user's password
	customErr, status := controller.userService.ChangeUserPassword(&userPasswordChangeDTO, username.(string))
	if customErr != nil {
		c.JSON(status, gin.H{
			"error": customErr,
		})
		return
	}

	c.JSON(status, gin.H{})
}

// GetUserProfile returns the user's profile
func (controller *UserController) GetUserProfile(c *gin.Context) {
	// Get logged-in username from middleware
	currentUsername, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": customerrors.UserUnauthorized,
		})
		return
	}

	// Get username from url
	username := c.Param("username")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": customerrors.BadRequest,
		})
		return
	}

	// Get user profile
	userProfileDTO, customErr, status := controller.userService.GetUserProfile(username, currentUsername.(string))
	if customErr != nil {
		c.JSON(status, gin.H{
			"error": customErr,
		})
		return
	}

	c.JSON(status, userProfileDTO)
}
