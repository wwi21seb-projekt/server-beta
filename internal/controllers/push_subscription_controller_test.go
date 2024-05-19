package controllers_test

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/wwi21seb-projekt/server-beta/internal/controllers"
	"github.com/wwi21seb-projekt/server-beta/internal/customerrors"
	"github.com/wwi21seb-projekt/server-beta/internal/middleware"
	"github.com/wwi21seb-projekt/server-beta/internal/models"
	"github.com/wwi21seb-projekt/server-beta/internal/repositories"
	"github.com/wwi21seb-projekt/server-beta/internal/services"
	"github.com/wwi21seb-projekt/server-beta/internal/utils"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestGetVapidKeySuccess tests if the function GetVapidKey returns a token if user is authorized
func TestGetVapidKeySuccess(t *testing.T) {
	// Arrange
	pushSubscriptionService := services.NewPushSubscriptionService(nil)
	pushSubscriptionController := controllers.NewPushSubscriptionController(pushSubscriptionService)

	authorizationToken, err := utils.GenerateAccessToken("testUser")
	if err != nil {
		t.Fatal(err)
	}

	// Setup HTTP request and recorder
	req, _ := http.NewRequest(http.MethodGet, "/push/vapid", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+authorizationToken)
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/push/vapid", middleware.AuthorizeUser, pushSubscriptionController.GetVapidKey)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code) // Expect HTTP 200 OK

	var vapidKeyResponse models.VapidKeyResponseDTO
	err = json.Unmarshal(w.Body.Bytes(), &vapidKeyResponse)
	assert.NoError(t, err)
}

// TestGetVapidKeyUnauthorized tests if the function GetVapidKey returns an error if user is not authorized
func TestGetVapidKeyUnauthorized(t *testing.T) {
	// Arrange
	pushSubscriptionController := controllers.NewPushSubscriptionController(nil)

	// Setup HTTP request and recorder
	req, _ := http.NewRequest(http.MethodGet, "/push/vapid", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/push/vapid", middleware.AuthorizeUser, pushSubscriptionController.GetVapidKey)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code) // Expect HTTP 401 Unauthorized

	var errorResponse customerrors.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)

	expectedCustomError := customerrors.UserUnauthorized
	assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
	assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)

}

// TestCreatePushSubscriptionSuccess tests if the function CreatePushSubscription saves a new push subscription key to the database if user is authorized and req body is correct
func TestCreatePushSubscriptionSuccess(t *testing.T) {
	// Arrange
	mockPushSubscriptionRepo := new(repositories.MockPushSubscriptionRepository)

	pushSubscriptionService := services.NewPushSubscriptionService(mockPushSubscriptionRepo)
	pushSubscriptionController := controllers.NewPushSubscriptionController(pushSubscriptionService)

	testUsername := "testUser"
	authorizationToken, err := utils.GenerateAccessToken(testUsername)
	if err != nil {
		t.Fatal(err)
	}

	pushSubscriptionCreateRequest := models.PushSubscriptionRequestDTO{
		Type: "web",
		SubscriptionInfo: models.SubscriptionInfo{
			Endpoint: "https://example.com",
			P256dh:   "p256dh",
			Auth:     "auth",
		},
	}

	// Mock expectations
	var capturedPushSubscription models.PushSubscription
	mockPushSubscriptionRepo.On("CreatePushSubscription", mock.AnythingOfType("*models.PushSubscription")).
		Run(func(args mock.Arguments) {
			capturedPushSubscription = *args.Get(0).(*models.PushSubscription)
		}).Return(nil)

	// Setup HTTP request and recorder
	requestBody, err := json.Marshal(pushSubscriptionCreateRequest)
	if err != nil {
		t.Fatal(err)
	}
	req, _ := http.NewRequest(http.MethodPost, "/push/register", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+authorizationToken)
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/push/register", middleware.AuthorizeUser, pushSubscriptionController.CreatePushSubscription)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusCreated, w.Code) // Expect HTTP 201 Created

	var responseObject models.PushSubscription
	err = json.Unmarshal(w.Body.Bytes(), &responseObject)
	assert.NoError(t, err)

	assert.NotEqual(t, responseObject.Id, "")
	assert.Equal(t, responseObject.Username, testUsername)
	assert.Equal(t, pushSubscriptionCreateRequest.Type, responseObject.Type)
	assert.Equal(t, pushSubscriptionCreateRequest.SubscriptionInfo.Endpoint, responseObject.Endpoint)
	assert.Equal(t, pushSubscriptionCreateRequest.SubscriptionInfo.P256dh, responseObject.P256dh)
	assert.Equal(t, pushSubscriptionCreateRequest.SubscriptionInfo.Auth, responseObject.Auth)

	assert.Equal(t, capturedPushSubscription.Id, responseObject.Id)
	assert.Equal(t, capturedPushSubscription.Username, testUsername)
	assert.Equal(t, capturedPushSubscription.Type, pushSubscriptionCreateRequest.Type)
	assert.Equal(t, capturedPushSubscription.Endpoint, pushSubscriptionCreateRequest.SubscriptionInfo.Endpoint)
	assert.Equal(t, capturedPushSubscription.P256dh, pushSubscriptionCreateRequest.SubscriptionInfo.P256dh)
	assert.Equal(t, capturedPushSubscription.Auth, pushSubscriptionCreateRequest.SubscriptionInfo.Auth)

	mockPushSubscriptionRepo.AssertExpectations(t)
}

// TestCreatePushSubscriptionUnauthorized tests if the function CreatePushSubscription returns an error if user is not authorized
func TestCreatePushSubscriptionUnauthorized(t *testing.T) {
	// Arrange
	pushSubscriptionController := controllers.NewPushSubscriptionController(nil)

	// Setup HTTP request and recorder
	req, _ := http.NewRequest(http.MethodPost, "/push/register", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/push/register", middleware.AuthorizeUser, pushSubscriptionController.CreatePushSubscription)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code) // Expect HTTP 401 Unauthorized

	var errorResponse customerrors.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)

	expectedCustomError := customerrors.UserUnauthorized
	assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
	assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)
}

// TestCreatePushSubscriptionBadRequest tests if the function CreatePushSubscription returns an error if req body is incorrect
func TestCreatePushSubscriptionBadRequest(t *testing.T) {
	invalidBodies := []string{
		`{"invalidField": "value"}`,
		``,
		`{type: "invalid, subscription: {}}"`, // invalid type, only "web" or "expo" allowed
	}

	for _, body := range invalidBodies {
		// Arrange
		pushSubscriptionService := services.NewPushSubscriptionService(nil)
		pushSubscriptionController := controllers.NewPushSubscriptionController(pushSubscriptionService)

		authenticationToken, err := utils.GenerateAccessToken("testUser")
		if err != nil {
			t.Fatal(err)
		}

		// Setup HTTP request and recorder
		req, _ := http.NewRequest(http.MethodPost, "/push/register", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+authenticationToken)
		w := httptest.NewRecorder()

		// Act
		gin.SetMode(gin.TestMode)
		router := gin.Default()
		router.POST("/push/register", middleware.AuthorizeUser, pushSubscriptionController.CreatePushSubscription)
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, w.Code) // Expect HTTP 400 Bad Request

		var errorResponse customerrors.ErrorResponse
		err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
		assert.NoError(t, err)

		expectedCustomError := customerrors.BadRequest
		assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
		assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)
	}
}