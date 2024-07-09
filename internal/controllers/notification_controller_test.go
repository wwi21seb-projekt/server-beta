package controllers_test

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/wwi21seb-projekt/server-beta/internal/controllers"
	"github.com/wwi21seb-projekt/server-beta/internal/customerrors"
	"github.com/wwi21seb-projekt/server-beta/internal/middleware"
	"github.com/wwi21seb-projekt/server-beta/internal/models"
	"github.com/wwi21seb-projekt/server-beta/internal/repositories"
	"github.com/wwi21seb-projekt/server-beta/internal/services"
	"github.com/wwi21seb-projekt/server-beta/internal/utils"
	"gorm.io/gorm"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestGetNotificationsSuccess tests if the GetNotifications returns a list of notifications
func TestGetNotificationsSuccess(t *testing.T) {
	// Arrange
	mockNotificationRepo := new(repositories.MockNotificationRepository)

	notificationService := services.NewNotificationService(mockNotificationRepo, nil)
	notificationController := controllers.NewNotificationController(notificationService)

	currentUsername := "testUser"
	authenticationToken, err := utils.GenerateAccessToken(currentUsername)
	if err != nil {
		t.Error(err)
	}

	otherUser := models.User{
		Username: "test2",
		Nickname: "nick",
		Image: models.Image{
			Width:  100,
			Height: 101,
		},
	}

	foundNotifications := []models.Notification{
		{
			Id:               uuid.New(),
			FromUser:         otherUser,
			ForUsername:      currentUsername,
			NotificationType: "follow",
			Timestamp:        time.Now().UTC(),
		},
		{
			Id:               uuid.New(),
			FromUser:         otherUser,
			ForUsername:      currentUsername,
			NotificationType: "repost",
			Timestamp:        time.Now().UTC(),
		},
	}

	// Mock expectations
	mockNotificationRepo.On("GetNotificationsByUsername", currentUsername).Return(foundNotifications, nil)

	// Setup HTTP request and recorder
	req, _ := http.NewRequest(http.MethodGet, "/notifications", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authenticationToken))
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/notifications", middleware.AuthorizeUser, notificationController.GetNotifications)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code) // Expect HTTP 200 OK
	mockNotificationRepo.AssertExpectations(t)

	var responseDto models.NotificationsResponseDTO
	err = json.Unmarshal(w.Body.Bytes(), &responseDto)
	if err != nil {
		t.Error(err)
	}

	assert.Equal(t, len(foundNotifications), len(responseDto.Records))
	for i, notification := range foundNotifications {
		assert.Equal(t, notification.Id.String(), responseDto.Records[i].NotificationId)
		assert.Equal(t, notification.NotificationType, responseDto.Records[i].NotificationType)
		assert.Equal(t, notification.FromUsername, responseDto.Records[i].User.Username)
		assert.Equal(t, notification.FromUser.Nickname, responseDto.Records[i].User.Nickname)
		assert.True(t, notification.Timestamp.Equal(responseDto.Records[i].Timestamp))
	}
}

// TestGetNotificationsUnauthorized tests if the GetNotifications returns an unauthorized error
func TestGetNotificationsUnauthorized(t *testing.T) {
	// Arrange
	mockNotificationRepo := new(repositories.MockNotificationRepository)

	notificationService := services.NewNotificationService(mockNotificationRepo, nil)
	notificationController := controllers.NewNotificationController(notificationService)

	// Setup HTTP request and recorder
	req, _ := http.NewRequest(http.MethodGet, "/notifications", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/notifications", middleware.AuthorizeUser, notificationController.GetNotifications)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code) // Expect HTTP 401 Unauthorized

	var errorResponse customerrors.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)

	expectedCustomError := customerrors.Unauthorized
	assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
	assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)

}

// TestDeleteNotificationByIdSuccess tests if the DeleteNotificationById returns a success message
func TestDeleteNotificationByIdSuccess(t *testing.T) {
	// Arrange
	mockNotificationRepo := new(repositories.MockNotificationRepository)

	notificationService := services.NewNotificationService(mockNotificationRepo, nil)
	notificationController := controllers.NewNotificationController(notificationService)

	currentUsername := "testUser"
	authenticationToken, err := utils.GenerateAccessToken(currentUsername)
	if err != nil {
		t.Error(err)
	}

	notificationId := uuid.New()
	notification := models.Notification{
		Id:               notificationId,
		ForUsername:      currentUsername,
		NotificationType: "follow",
	}

	// Mock expectations
	mockNotificationRepo.On("GetNotificationById", notificationId.String()).Return(notification, nil)
	mockNotificationRepo.On("DeleteNotificationById", notificationId.String()).Return(nil)

	// Setup HTTP request and recorder
	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/notifications/%s", notificationId.String()), nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authenticationToken))
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.DELETE("/notifications/:notificationId", middleware.AuthorizeUser, notificationController.DeleteNotificationById)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNoContent, w.Code) // Expect HTTP 204 Not Content
	mockNotificationRepo.AssertExpectations(t)

	assert.Empty(t, w.Body.String())
}

// TestDeleteNotificationByIdUnauthorized tests if the DeleteNotificationById returns an unauthorized error if the user is not authorized
func TestDeleteNotificationByIdUnauthorized(t *testing.T) {
	// Arrange
	mockNotificationRepo := new(repositories.MockNotificationRepository)

	notificationService := services.NewNotificationService(mockNotificationRepo, nil)
	notificationController := controllers.NewNotificationController(notificationService)

	notificationId := uuid.New()

	// Setup HTTP request and recorder
	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/notifications/%s", notificationId.String()), nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.DELETE("/notifications/:notificationId", middleware.AuthorizeUser, notificationController.DeleteNotificationById)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code) // Expect HTTP 401 Unauthorized

	var errorResponse customerrors.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)

	expectedCustomError := customerrors.Unauthorized
	assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
	assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)
}

// TestDeleteNotificationByIdNotFound tests if the DeleteNotificationById returns a not found error if the notification does not exist
func TestDeleteNotificationByIdNotFound(t *testing.T) {
	// Arrange
	mockNotificationRepo := new(repositories.MockNotificationRepository)

	notificationService := services.NewNotificationService(mockNotificationRepo, nil)
	notificationController := controllers.NewNotificationController(notificationService)

	currentUsername := "testUser"
	authenticationToken, err := utils.GenerateAccessToken(currentUsername)
	if err != nil {
		t.Error(err)
	}

	notificationId := uuid.New()

	// Mock expectations
	mockNotificationRepo.On("GetNotificationById", notificationId.String()).Return(models.Notification{}, gorm.ErrRecordNotFound)

	// Setup HTTP request and recorder
	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/notifications/%s", notificationId.String()), nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authenticationToken))
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.DELETE("/notifications/:notificationId", middleware.AuthorizeUser, notificationController.DeleteNotificationById)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code) // Expect HTTP 404 Not Found
	mockNotificationRepo.AssertExpectations(t)

	var errorResponse customerrors.ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)

	expectedCustomError := customerrors.NotificationNotFound
	assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
	assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)
}

// TestDeleteNotificationByIdForbidden tests if the DeleteNotificationById returns a forbidden error if the notification does not belong to the user
func TestDeleteNotificationByIdForbidden(t *testing.T) {
	// Arrange
	mockNotificationRepo := new(repositories.MockNotificationRepository)

	notificationService := services.NewNotificationService(mockNotificationRepo, nil)
	notificationController := controllers.NewNotificationController(notificationService)

	currentUsername := "testUser"
	authenticationToken, err := utils.GenerateAccessToken(currentUsername)
	if err != nil {
		t.Error(err)
	}

	notificationId := uuid.New()
	notification := models.Notification{
		Id:               notificationId,
		ForUsername:      "otherUser",
		NotificationType: "follow",
	}

	// Mock expectations
	mockNotificationRepo.On("GetNotificationById", notificationId.String()).Return(notification, nil)

	// Setup HTTP request and recorder
	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/notifications/%s", notificationId.String()), nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authenticationToken))
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.DELETE("/notifications/:notificationId", middleware.AuthorizeUser, notificationController.DeleteNotificationById)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusForbidden, w.Code) // Expect HTTP 403 Forbidden
	mockNotificationRepo.AssertExpectations(t)

	var errorResponse customerrors.ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)

	expectedCustomError := customerrors.DeleteNotificationForbidden
	assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
	assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)
}
