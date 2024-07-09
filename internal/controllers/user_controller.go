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
	limitStr := c.DefaultQuery("limit", "10")
	offsetStr := c.DefaultQuery("offset", "0")

	// Convert limit and offset to int
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 10
	}
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		offset = 0
	}

	// Current username from middleware
	currentUsername, exists := c.Get("username")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": customerrors.Unauthorized,
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
			"error": customerrors.Unauthorized,
		})
		return
	}

	// Bind the JSON request body to a map
	var requestData map[string]interface{} // either nickname or status can be empty but ensure that the keys are present
	if err := c.ShouldBindJSON(&requestData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": customerrors.BadRequest,
		})
		return
	}

	// Check if the keys are present in the request
	_, nicknamePresent := requestData["nickname"]
	_, statusPresent := requestData["status"]
	if !nicknamePresent && !statusPresent {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": customerrors.BadRequest,
		})
		return
	}

	// Map the data to DTO
	var userUpdateRequestDTO models.UserInformationUpdateRequestDTO
	if nickname, ok := requestData["nickname"].(string); ok {
		userUpdateRequestDTO.Nickname = nickname
	}
	if status, ok := requestData["status"].(string); ok {
		userUpdateRequestDTO.Status = status
	}

	// Differentiate between nil and "" for picture
	picture, picturePresent := requestData["picture"] // Check if picture is present null
	pictureString, ok := picture.(string)             // convert picture to string
	if picture == nil || !picturePresent || !ok {     // if picture was not present, not nil or not a string, set dto picture to nil
		userUpdateRequestDTO.Picture = nil // nil is used to indicate that the picture should not be updated
	} else {
		userUpdateRequestDTO.Picture = &pictureString // "" is used to indicate that the picture should be removed, otherwise base64 string
	}

	// Update the user's information
	responseDTO, customErr, status := controller.userService.UpdateUserInformation(&userUpdateRequestDTO, username.(string))
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
			"error": customerrors.Unauthorized,
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
			"error": customerrors.Unauthorized,
		})
		return
	}

	// Get username from url
	username := c.Param("username")

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
