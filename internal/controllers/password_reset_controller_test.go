package controllers_test

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/wwi21seb-projekt/server-beta/internal/controllers"
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

// TestInitiatePasswordResetSuccess tests if the InitiatePasswordReset function returns 200-OK and censored email when successful
func TestInitiatePasswordResetSuccess(t *testing.T) {
	// Arrange
	mockUserRepo := new(repositories.MockUserRepository)
	mockPasswordResetRepo := new(repositories.MockPasswordResetRepository)
	mockMailService := new(services.MockMailService)
	mockValidator := new(utils.MockValidator)

	passwordResetService := services.NewPasswordResetService(mockUserRepo, mockPasswordResetRepo, mockMailService, mockValidator)
	passwordResetController := controllers.NewPasswordResetController(passwordResetService)

	username := "testUser"
	email := "test@example.com"

	censoredMail := "tes*@example.com"

	user := models.User{
		Username: username,
		Email:    email,
	}

	// Mock expectations
	mockUserRepo.On("FindUserByUsername", username).Return(&user, nil)
	mockPasswordResetRepo.On("DeletePasswordResetTokensByUsername", username).Return(nil)

	var capturedToken *models.PasswordResetToken
	mockPasswordResetRepo.On("CreatePasswordResetToken", mock.AnythingOfType("*models.PasswordResetToken")).
		Run(func(args mock.Arguments) {
			capturedToken = args.Get(0).(*models.PasswordResetToken)
		}).Return(nil)

	var capturedEmailBody string
	mockMailService.On("SendMail", email, "Reset your password", mock.AnythingOfType("string")).
		Run(func(args mock.Arguments) {
			capturedEmailBody = args.String(2)
		}).Return(nil)

	// Setup HTTP request
	url := "/users/" + username + "/reset-password"

	req, _ := http.NewRequest("POST", url, nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/users/:username/reset-password", passwordResetController.InitiatePasswordReset)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockUserRepo.AssertExpectations(t)
	mockPasswordResetRepo.AssertExpectations(t)
	mockMailService.AssertExpectations(t)

	assert.NotNil(t, capturedToken)
	assert.NotEmpty(t, capturedToken.Id)
	assert.Equal(t, username, capturedToken.Username)
	assert.NotEmpty(t, capturedToken.Token)
	assert.WithinDuration(t, time.Now().Add(2*time.Hour), capturedToken.ExpirationTime, time.Minute)

	var responseBody models.InitiatePasswordResetResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &responseBody)
	assert.NoError(t, err)
	assert.Equal(t, censoredMail, responseBody.Email)

	assert.Contains(t, capturedEmailBody, capturedToken.Token)
}

// TestInitiatePasswordResetUserNotFound tests if the TestInitiatePassword function returns 400 Bad Request when the user to the given username cannot be found
func TestInitiatePasswordResetUserNotFound(t *testing.T) {
	// Arrange
	mockUserRepo := new(repositories.MockUserRepository)
	mockPasswordResetRepo := new(repositories.MockPasswordResetRepository)
	mockMailService := new(services.MockMailService)
	mockValidator := new(utils.MockValidator)

	passwordResetService := services.NewPasswordResetService(mockUserRepo, mockPasswordResetRepo, mockMailService, mockValidator)
	passwordResetController := controllers.NewPasswordResetController(passwordResetService)

	username := "testUser"

	// Mock expectations
	mockUserRepo.On("FindUserByUsername", username).Return(&models.User{}, gorm.ErrRecordNotFound)

	// Setup HTTP request
	url := "/users/" + username + "/reset-password"

	req, _ := http.NewRequest("POST", url, nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/users/:username/reset-password", passwordResetController.InitiatePasswordReset)
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

// TestInitiatePasswordResetMailNotSent tests if the TestInitiatePassword function returns 500 Internal Server Error when the mail could not be sent
func TestInitiatePasswordResetMailNotSent(t *testing.T) {
	// Arrange
	mockUserRepo := new(repositories.MockUserRepository)
	mockPasswordResetRepo := new(repositories.MockPasswordResetRepository)
	mockMailService := new(services.MockMailService)
	mockValidator := new(utils.MockValidator)

	passwordResetService := services.NewPasswordResetService(mockUserRepo, mockPasswordResetRepo, mockMailService, mockValidator)
	passwordResetController := controllers.NewPasswordResetController(passwordResetService)

	username := "testUser"
	email := "test@example.com"

	user := models.User{
		Username: username,
		Email:    email,
	}

	// Mock expectations
	mockUserRepo.On("FindUserByUsername", username).Return(&user, nil)
	mockPasswordResetRepo.On("DeletePasswordResetTokensByUsername", username).Return(nil)
	mockPasswordResetRepo.On("CreatePasswordResetToken", mock.AnythingOfType("*models.PasswordResetToken")).Return(nil)
	mockMailService.On("SendMail", email, "Reset your password", mock.AnythingOfType("string")).Return(customerrors.EmailNotSent)

	// Setup HTTP request
	url := "/users/" + username + "/reset-password"

	req, _ := http.NewRequest("POST", url, nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/users/:username/reset-password", passwordResetController.InitiatePasswordReset)
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

// TestResetPasswordSuccess tests the ResetPassword function if it returns 200 OK if a correct token and strong password is given
func TestResetPasswordSuccess(t *testing.T) {
	// Arrange
	mockUserRepo := new(repositories.MockUserRepository)
	mockPasswordResetRepo := new(repositories.MockPasswordResetRepository)
	mockMailService := new(services.MockMailService)
	validator := new(utils.Validator)

	passwordResetService := services.NewPasswordResetService(mockUserRepo, mockPasswordResetRepo, mockMailService, validator)
	passwordResetController := controllers.NewPasswordResetController(passwordResetService)

	username := "testUser"
	newPassword := "newPassword123!"
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
	mockPasswordResetRepo.On("FindPasswordResetToken", username, token).Return(&resetToken, nil)
	mockUserRepo.On("FindUserByUsername", username).Return(&user, nil)

	var capturedUpdatedUser *models.User
	mockUserRepo.On("UpdateUser", mock.AnythingOfType("*models.User")).
		Run(func(args mock.Arguments) {
			capturedUpdatedUser = args.Get(0).(*models.User)
		}).Return(nil)
	mockPasswordResetRepo.On("DeletePasswordResetTokenById", resetToken.Id.String()).Return(nil)

	// Setup HTTP request
	url := "/users/" + username + "/reset-password"
	body := `{"token": "` + token + `", "newPassword": "` + newPassword + `"}`

	req, _ := http.NewRequest("PATCH", url, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.PATCH("/users/:username/reset-password", passwordResetController.ResetPassword)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNoContent, w.Code)
	mockPasswordResetRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)

	check := utils.CheckPassword(newPassword, capturedUpdatedUser.PasswordHash)
	assert.True(t, check)
}

// TestResetPasswordBadRequest tests the ResetPassword function if it returns 400 Bad Request if the body or password does not meet specifications
func TestResetPasswordBadRequest(t *testing.T) {
	invalidBodies := []string{
		`{}`,                   // Empty body
		`{"token"": "123456"}`, // Not complete
		`{"token":"", "newPassword":"ValidPassword123!"}`,
		`{"token": "1234562", "newPassword":"InvalidPassword123"}`, // New password does not match password policy
	}

	for _, body := range invalidBodies {

		// Arrange
		mockUserRepo := new(repositories.MockUserRepository)
		mockPasswordResetRepo := new(repositories.MockPasswordResetRepository)
		mockMailService := new(services.MockMailService)
		validator := new(utils.Validator)

		passwordResetService := services.NewPasswordResetService(mockUserRepo, mockPasswordResetRepo, mockMailService, validator)
		passwordResetController := controllers.NewPasswordResetController(passwordResetService)

		username := "testUser"

		// Setup HTTP request
		url := "/users/" + username + "/reset-password"
		req, _ := http.NewRequest("PATCH", url, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		// Act
		gin.SetMode(gin.TestMode)
		router := gin.Default()
		router.PATCH("/users/:username/reset-password", passwordResetController.ResetPassword)
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, w.Code)
		var errorResponse customerrors.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
		assert.NoError(t, err)

		expectedCustomError := customerrors.BadRequest
		assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
		assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)
	}
}

// TestResetPasswordInvalidToken tests the ResetPassword function if it returns 403 Forbidden if the token cannot be found
func TestResetPasswordInvalidToken(t *testing.T) {
	// Arrange
	mockUserRepo := new(repositories.MockUserRepository)
	mockPasswordResetRepo := new(repositories.MockPasswordResetRepository)
	mockMailService := new(services.MockMailService)
	mockValidator := new(utils.MockValidator)

	passwordResetService := services.NewPasswordResetService(mockUserRepo, mockPasswordResetRepo, mockMailService, mockValidator)
	passwordResetController := controllers.NewPasswordResetController(passwordResetService)

	username := "testUser"
	token := "123456"
	newPassword := "newPassword123!"

	user := models.User{
		Username: username,
	}

	// Mock expectations
	mockUserRepo.On("FindUserByUsername", username).Return(&user, nil)
	mockPasswordResetRepo.On("FindPasswordResetToken", username, token).Return(&models.PasswordResetToken{}, gorm.ErrRecordNotFound)

	// Setup HTTP request
	url := "/users/" + username + "/reset-password"
	body := `{"token": "` + token + `", "newPassword": "` + newPassword + `"}`

	req, _ := http.NewRequest("PATCH", url, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.PATCH("/users/:username/reset-password", passwordResetController.ResetPassword)
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

// TestResetPasswordExpiredToken tests the ResetPassword function if it returns 403 Forbidden if the token is expired
func TestResetPasswordExpiredToken(t *testing.T) {
	// Arrange
	mockUserRepo := new(repositories.MockUserRepository)
	mockPasswordResetRepo := new(repositories.MockPasswordResetRepository)
	mockMailService := new(services.MockMailService)
	mockValidator := new(utils.MockValidator)

	passwordResetService := services.NewPasswordResetService(mockUserRepo, mockPasswordResetRepo, mockMailService, mockValidator)
	passwordResetController := controllers.NewPasswordResetController(passwordResetService)

	username := "testUser"
	newPassword := "newPassword123!"
	token := "123456"

	resetToken := models.PasswordResetToken{
		Id:             uuid.New(),
		Token:          token,
		Username:       username,
		ExpirationTime: time.Now().Add(-24 * time.Hour), // expired token
	}

	user := models.User{
		Username: username,
	}

	// Mock expectations
	mockUserRepo.On("FindUserByUsername", username).Return(&user, nil)
	mockPasswordResetRepo.On("FindPasswordResetToken", username, token).Return(&resetToken, nil)

	// Setup HTTP request
	url := "/users/" + username + "/reset-password"
	body := `{"token": "` + token + `", "newPassword": "` + newPassword + `"}`

	req, _ := http.NewRequest("PATCH", url, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.PATCH("/users/:username/reset-password", passwordResetController.ResetPassword)
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

// TestResetPasswordUserNotFound tests the ResetPassword function if it returns 404 Not Found if the user to the given username cannot be found
func TestResetPasswordUserNotFound(t *testing.T) {
	// Arrange
	mockUserRepo := new(repositories.MockUserRepository)
	mockPasswordResetRepo := new(repositories.MockPasswordResetRepository)
	mockMailService := new(services.MockMailService)
	mockValidator := new(utils.MockValidator)

	passwordResetService := services.NewPasswordResetService(mockUserRepo, mockPasswordResetRepo, mockMailService, mockValidator)
	passwordResetController := controllers.NewPasswordResetController(passwordResetService)

	username := "testUser"
	newPassword := "newPassword123!"
	token := "123456"

	// Mock expectations
	mockUserRepo.On("FindUserByUsername", username).Return(&models.User{}, gorm.ErrRecordNotFound)

	// Setup HTTP request
	url := "/users/" + username + "/reset-password"
	body := `{"token": "` + token + `", "newPassword": "` + newPassword + `"}`

	req, _ := http.NewRequest("PATCH", url, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.PATCH("/users/:username/reset-password", passwordResetController.ResetPassword)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code)
	var errorResponse customerrors.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)

	expectedCustomError := customerrors.UserNotFound
	assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
	assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)

	mockPasswordResetRepo.AssertExpectations(t)
}
