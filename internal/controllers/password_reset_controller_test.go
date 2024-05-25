package controllers

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/wwi21seb-projekt/server-beta/internal/customerrors"
	"github.com/wwi21seb-projekt/server-beta/internal/models"
	"github.com/wwi21seb-projekt/server-beta/internal/repositories"
	"github.com/wwi21seb-projekt/server-beta/internal/services"
	"github.com/wwi21seb-projekt/server-beta/internal/utils"
	"gorm.io/gorm"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// TestPasswordResetSuccess tests if PasswordReset initiates the password reset process and returns 204 No Content
func TestPasswordResetSuccess(t *testing.T) {
	// Arrange
	mockUserRepo := new(repositories.MockUserRepository)
	mockPasswordResetRepo := new(repositories.MockPasswordResetRepository)
	mockMailService := new(services.MockMailService)
	mockValidator := new(utils.MockValidator)

	passwordResetService := services.NewPasswordResetService(mockUserRepo, mockPasswordResetRepo, mockMailService, mockValidator)
	passwordResetController := NewPasswordResetController(passwordResetService)

	username := "testUser"
	email := "test@example.com"

	user := models.User{
		Username: username,
		Email:    email,
	}

	// Mock expectations
	mockUserRepo.On("FindUserByUsername", username).Return(&user, nil)
	mockPasswordResetRepo.On("CreatePasswordResetToken", mock.AnythingOfType("*models.PasswordResetToken")).Return(nil)
	mockMailService.On("SendMail", email, "Password Reset Token", mock.AnythingOfType("string")).Return(nil)

	// Setup HTTP request
	url := "/password-reset"
	body := `{"username": "` + username + `"}`

	req, _ := http.NewRequest("POST", url, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/password-reset", passwordResetController.PasswordReset)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNoContent, w.Code)
	mockUserRepo.AssertExpectations(t)
	mockPasswordResetRepo.AssertExpectations(t)
	mockMailService.AssertExpectations(t)
}

// TestPasswordResetUserNotFound tests if PasswordReset returns 404 Not Found if the user does not exist
func TestPasswordResetUserNotFound(t *testing.T) {
	// Arrange
	mockUserRepo := new(repositories.MockUserRepository)
	mockPasswordResetRepo := new(repositories.MockPasswordResetRepository)
	mockMailService := new(services.MockMailService)
	mockValidator := new(utils.MockValidator)

	passwordResetService := services.NewPasswordResetService(mockUserRepo, mockPasswordResetRepo, mockMailService, mockValidator)
	passwordResetController := NewPasswordResetController(passwordResetService)

	username := "testUser"

	// Mock expectations
	mockUserRepo.On("FindUserByUsername", username).Return(nil, gorm.ErrRecordNotFound)

	// Setup HTTP request
	url := "/password-reset"
	body := `{"username": "` + username + `"}`

	req, _ := http.NewRequest("POST", url, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/password-reset", passwordResetController.PasswordReset)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
	var errorResponse customerrors.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)

	expectedCustomError := customerrors.UserNotFound
	assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
	assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)

	mockUserRepo.AssertExpectations(t)
	mockPasswordResetRepo.AssertExpectations(t)
	mockMailService.AssertExpectations(t)
}

// TestPasswordResetMailNotSent tests if PasswordReset returns 500 Internal Server Error if the mail could not be sent
func TestPasswordResetMailNotSent(t *testing.T) {
	// Arrange
	mockUserRepo := new(repositories.MockUserRepository)
	mockPasswordResetRepo := new(repositories.MockPasswordResetRepository)
	mockMailService := new(services.MockMailService)
	mockValidator := new(utils.MockValidator)

	passwordResetService := services.NewPasswordResetService(mockUserRepo, mockPasswordResetRepo, mockMailService, mockValidator)
	passwordResetController := NewPasswordResetController(passwordResetService)

	username := "testUser"
	email := "test@example.com"

	user := models.User{
		Username: username,
		Email:    email,
	}

	// Mock expectations
	mockUserRepo.On("FindUserByUsername", username).Return(&user, nil)
	mockPasswordResetRepo.On("CreatePasswordResetToken", mock.AnythingOfType("*models.PasswordResetToken")).Return(nil)
	mockMailService.On("SendMail", email, "Password Reset Token", mock.AnythingOfType("string")).Return(customerrors.EmailNotSent)

	// Setup HTTP request
	url := "/password-reset"
	body := `{"username": "` + username + `"}`

	req, _ := http.NewRequest("POST", url, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/password-reset", passwordResetController.PasswordReset)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	var errorResponse customerrors.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)

	expectedCustomError := customerrors.EmailNotSent
	assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
	assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)

	mockUserRepo.AssertExpectations(t)
	mockPasswordResetRepo.AssertExpectations(t)
	mockMailService.AssertExpectations(t)
}

// TestSetNewPasswordSuccess tests if SetNewPassword sets a new password for the user and returns 204 No Content
func TestSetNewPasswordSuccess(t *testing.T) {
	// Arrange
	mockUserRepo := new(repositories.MockUserRepository)
	mockPasswordResetRepo := new(repositories.MockPasswordResetRepository)
	mockMailService := new(services.MockMailService)
	mockValidator := new(utils.MockValidator)

	passwordResetService := services.NewPasswordResetService(mockUserRepo, mockPasswordResetRepo, mockMailService, mockValidator)
	passwordResetController := NewPasswordResetController(passwordResetService)

	username := "testUser"
	newPassword := "newPassword123"
	token := "123456"

	user := models.User{
		Username: username,
	}

	resetToken := models.PasswordResetToken{
		Token:          token,
		Username:       username,
		ExpirationTime: time.Now().Add(1 * time.Hour),
	}

	// Mock expectations
	mockValidator.On("ValidatePassword", newPassword).Return(true)
	mockPasswordResetRepo.On("FindPasswordResetToken", token).Return(&resetToken, nil)
	mockUserRepo.On("FindUserByUsername", username).Return(&user, nil)
	mockUserRepo.On("UpdateUser", mock.AnythingOfType("*models.User")).Return(nil)
	mockPasswordResetRepo.On("DeletePasswordResetToken", token).Return(nil)

	// Setup HTTP request
	url := "/set-new-password"
	body := `{"token": "` + token + `", "newPassword": "` + newPassword + `"}`

	req, _ := http.NewRequest("PATCH", url, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.PATCH("/set-new-password", passwordResetController.SetNewPassword)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNoContent, w.Code)
	mockValidator.AssertExpectations(t)
	mockPasswordResetRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

// TestSetNewPasswordInvalidToken tests if SetNewPassword returns 404 Not Found if the token is invalid
func TestSetNewPasswordInvalidToken(t *testing.T) {
	// Arrange
	mockUserRepo := new(repositories.MockUserRepository)
	mockPasswordResetRepo := new(repositories.MockPasswordResetRepository)
	mockMailService := new(services.MockMailService)
	mockValidator := new(utils.MockValidator)

	passwordResetService := services.NewPasswordResetService(mockUserRepo, mockPasswordResetRepo, mockMailService, mockValidator)
	passwordResetController := NewPasswordResetController(passwordResetService)

	token := "123456"
	newPassword := "newPassword123"

	// Mock expectations
	mockPasswordResetRepo.On("FindPasswordResetToken", token).Return(nil, gorm.ErrRecordNotFound)

	// Setup HTTP request
	url := "/set-new-password"
	body := `{"token": "` + token + `", "newPassword": "` + newPassword + `"}`

	req, _ := http.NewRequest("PATCH", url, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.PATCH("/set-new-password", passwordResetController.SetNewPassword)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
	var errorResponse customerrors.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)

	expectedCustomError := customerrors.InvalidToken
	assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
	assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)

	mockPasswordResetRepo.AssertExpectations(t)
}

// TestSetNewPasswordExpiredToken tests if SetNewPassword returns 401 Unauthorized if the token is expired
func TestSetNewPasswordExpiredToken(t *testing.T) {
	// Arrange
	mockUserRepo := new(repositories.MockUserRepository)
	mockPasswordResetRepo := new(repositories.MockPasswordResetRepository)
	mockMailService := new(services.MockMailService)
	mockValidator := new(utils.MockValidator)

	passwordResetService := services.NewPasswordResetService(mockUserRepo, mockPasswordResetRepo, mockMailService, mockValidator)
	passwordResetController := NewPasswordResetController(passwordResetService)

	username := "testUser"
	newPassword := "newPassword123"
	token := "123456"

	resetToken := models.PasswordResetToken{
		Token:          token,
		Username:       username,
		ExpirationTime: time.Now().Add(-1 * time.Hour), // expired token
	}

	// Mock expectations
	mockPasswordResetRepo.On("FindPasswordResetToken", token).Return(&resetToken, nil)

	// Setup HTTP request
	url := "/set-new-password"
	body := `{"token": "` + token + `", "newPassword": "` + newPassword + `"}`

	req, _ := http.NewRequest("PATCH", url, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.PATCH("/set-new-password", passwordResetController.SetNewPassword)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	var errorResponse customerrors.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)

	expectedCustomError := customerrors.ActivationTokenExpired
	assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
	assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)

	mockPasswordResetRepo.AssertExpectations(t)
}

// TestSetNewPasswordInvalidPassword tests if SetNewPassword returns 400 Bad Request if the new password is invalid
func TestSetNewPasswordInvalidPassword(t *testing.T) {
	// Arrange
	mockUserRepo := new(repositories.MockUserRepository)
	mockPasswordResetRepo := new(repositories.MockPasswordResetRepository)
	mockMailService := new(services.MockMailService)
	mockValidator := new(utils.MockValidator)

	passwordResetService := services.NewPasswordResetService(mockUserRepo, mockPasswordResetRepo, mockMailService, mockValidator)
	passwordResetController := NewPasswordResetController(passwordResetService)

	token := "123456"
	newPassword := "short"

	// Mock expectations
	mockValidator.On("ValidatePassword", newPassword).Return(false)

	// Setup HTTP request
	url := "/set-new-password"
	body := `{"token": "` + token + `", "newPassword": "` + newPassword + `"}`

	req, _ := http.NewRequest("PATCH", url, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.PATCH("/set-new-password", passwordResetController.SetNewPassword)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
	var errorResponse customerrors.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)

	expectedCustomError := customerrors.BadRequest
	assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
	assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)

	mockValidator.AssertExpectations(t)
}
