package controllers

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
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

	var capturedToken *models.PasswordResetToken
	mockPasswordResetRepo.On("CreatePasswordResetToken", mock.AnythingOfType("*models.PasswordResetToken")).
		Run(func(args mock.Arguments) {
			capturedToken = args.Get(0).(*models.PasswordResetToken)
		}).Return(nil)

	var capturedEmail, capturedSubject, capturedBody string
	mockMailService.On("SendMail", email, "Password Reset Token", mock.AnythingOfType("string")).
		Run(func(args mock.Arguments) {
			capturedEmail = args.String(0)
			capturedSubject = args.String(1)
			capturedBody = args.String(2)
		}).Return(nil)

	// Setup HTTP request
	url := "/password-reset/" + username

	req, _ := http.NewRequest("POST", url, nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/password-reset/:username", passwordResetController.PasswordReset)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockUserRepo.AssertExpectations(t)
	mockPasswordResetRepo.AssertExpectations(t)
	mockMailService.AssertExpectations(t)

	// Additional assertions
	assert.NotNil(t, capturedToken)
	assert.Equal(t, username, capturedToken.Username)
	assert.NotEmpty(t, capturedToken.Token)
	assert.WithinDuration(t, time.Now().Add(2*time.Hour), capturedToken.ExpirationTime, time.Minute)

	assert.Equal(t, email, capturedEmail)
	assert.Equal(t, "Password Reset Token", capturedSubject)
	assert.Contains(t, capturedBody, capturedToken.Token)
}

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
	url := "/password-reset/" + username

	req, _ := http.NewRequest("POST", url, nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/password-reset/:username", passwordResetController.PasswordReset)
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
	url := "/password-reset/" + username

	req, _ := http.NewRequest("POST", url, nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/password-reset/:username", passwordResetController.PasswordReset)
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
		Id:             uuid.New(),
		Token:          token,
		Username:       username,
		ExpirationTime: time.Now().Add(1 * time.Hour),
	}

	// Mock expectations
	mockValidator.On("ValidatePassword", newPassword).Return(true)
	mockPasswordResetRepo.On("FindPasswordResetToken", username, token).Return(&resetToken, nil)
	mockUserRepo.On("FindUserByUsername", username).Return(&user, nil)
	mockUserRepo.On("UpdateUser", mock.AnythingOfType("*models.User")).Return(nil)
	mockPasswordResetRepo.On("DeletePasswordResetToken", resetToken.Id.String()).Return(nil)

	// Setup HTTP request
	url := "/users/" + username + "/set-new-password"
	body := `{"token": "` + token + `", "newPassword": "` + newPassword + `"}`

	req, _ := http.NewRequest("PATCH", url, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.PATCH("/users/:username/set-new-password", passwordResetController.SetNewPassword)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNoContent, w.Code)
	mockValidator.AssertExpectations(t)
	mockPasswordResetRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

func TestSetNewPasswordInvalidToken(t *testing.T) {
	// Arrange
	mockUserRepo := new(repositories.MockUserRepository)
	mockPasswordResetRepo := new(repositories.MockPasswordResetRepository)
	mockMailService := new(services.MockMailService)
	mockValidator := new(utils.MockValidator)

	passwordResetService := services.NewPasswordResetService(mockUserRepo, mockPasswordResetRepo, mockMailService, mockValidator)
	passwordResetController := NewPasswordResetController(passwordResetService)

	username := "testUser"
	token := "123456"
	newPassword := "newPassword123"

	// Mock expectations
	mockPasswordResetRepo.On("FindPasswordResetToken", username, token).Return(nil, gorm.ErrRecordNotFound)

	// Setup HTTP request
	url := "/users/" + username + "/set-new-password"
	body := `{"token": "` + token + `", "newPassword": "` + newPassword + `"}`

	req, _ := http.NewRequest("PATCH", url, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.PATCH("/users/:username/set-new-password", passwordResetController.SetNewPassword)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusForbidden, w.Code)
	var errorResponse customerrors.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)

	expectedCustomError := customerrors.PasswordResetTokenInvalid
	assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
	assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)

	mockPasswordResetRepo.AssertExpectations(t)
}

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
		Id:             uuid.New(),
		Token:          token,
		Username:       username,
		ExpirationTime: time.Now().Add(-1 * time.Hour), // expired token
	}

	// Mock expectations
	mockPasswordResetRepo.On("FindPasswordResetToken", username, token).Return(&resetToken, nil)

	// Setup HTTP request
	url := "/users/" + username + "/set-new-password"
	body := `{"token": "` + token + `", "newPassword": "` + newPassword + `"}`

	req, _ := http.NewRequest("PATCH", url, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.PATCH("/users/:username/set-new-password", passwordResetController.SetNewPassword)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusForbidden, w.Code)
	var errorResponse customerrors.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)

	expectedCustomError := customerrors.PasswordResetTokenInvalid
	assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
	assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)

	mockPasswordResetRepo.AssertExpectations(t)
}

func TestSetNewPasswordInvalidPassword(t *testing.T) {
	// Arrange
	mockUserRepo := new(repositories.MockUserRepository)
	mockPasswordResetRepo := new(repositories.MockPasswordResetRepository)
	mockMailService := new(services.MockMailService)
	mockValidator := new(utils.MockValidator)

	passwordResetService := services.NewPasswordResetService(mockUserRepo, mockPasswordResetRepo, mockMailService, mockValidator)
	passwordResetController := NewPasswordResetController(passwordResetService)

	username := "testUser"
	token := "123456"
	newPassword := "short"

	// Mock expectations
	mockValidator.On("ValidatePassword", newPassword).Return(false)

	// Setup HTTP request
	url := "/users/" + username + "/set-new-password"
	body := `{"token": "` + token + `", "newPassword": "` + newPassword + `"}`

	req, _ := http.NewRequest("PATCH", url, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.PATCH("/users/:username/set-new-password", passwordResetController.SetNewPassword)
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
