package controllers_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

// TestPostSubscriptionSuccess tests if PostSubscription returns 201-Created and correct response body when subscription is created successfully
func TestPostSubscriptionSuccess(t *testing.T) {
	// Arrange
	mockSubscriptionRepo := new(repositories.MockSubscriptionRepository)
	mockUserRepo := new(repositories.MockUserRepository)

	subscriptionService := services.NewSubscriptionService(mockSubscriptionRepo, mockUserRepo)
	subscriptionController := controllers.NewSubscriptionController(subscriptionService)

	currentUsername := "testUser"
	authenticationToken, err := utils.GenerateAccessToken(currentUsername)
	if err != nil {
		t.Error(err)
	}

	subscriptionCreateRequest := models.SubscriptionPostRequestDTO{
		Following: "testUser2",
	}

	// Mock expectations
	var capturedSubscription models.Subscription
	mockUserRepo.On("FindUserByUsername", subscriptionCreateRequest.Following).Return(&models.User{}, nil)                                                             // Expect user to be found
	mockSubscriptionRepo.On("GetSubscriptionByUsernames", currentUsername, subscriptionCreateRequest.Following).Return(&models.Subscription{}, gorm.ErrRecordNotFound) // Expect user is not following already
	mockSubscriptionRepo.On("CreateSubscription", mock.AnythingOfType("*models.Subscription")).
		Run(func(args mock.Arguments) {
			capturedSubscription = *args.Get(0).(*models.Subscription) // Save argument to captor
		}).Return(nil) // Expect subscription to be created

	// Setup HTTP request and recorder
	requestBody, err := json.Marshal(subscriptionCreateRequest)
	if err != nil {
		t.Fatal(err)
	}
	req, _ := http.NewRequest(http.MethodPost, "/subscriptions", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+authenticationToken)
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/subscriptions", middleware.AuthorizeUser, subscriptionController.PostSubscription)
	router.ServeHTTP(w, req)

	// Assert Response
	assert.Equal(t, http.StatusCreated, w.Code) // Expect HTTP 201 Created status
	var responseSubscription models.SubscriptionPostResponseDTO
	err = json.Unmarshal(w.Body.Bytes(), &responseSubscription)
	assert.NoError(t, err)

	assert.Equal(t, subscriptionCreateRequest.Following, capturedSubscription.FollowingUsername)
	assert.Equal(t, currentUsername, capturedSubscription.FollowerUsername)
	assert.NotNil(t, capturedSubscription.Id)
	assert.NotNil(t, capturedSubscription.SubscriptionDate)
	assert.Equal(t, capturedSubscription.Id, responseSubscription.SubscriptionId)
	assert.Equal(t, currentUsername, responseSubscription.Follower)
	assert.Equal(t, subscriptionCreateRequest.Following, responseSubscription.Following)
	assert.True(t, capturedSubscription.SubscriptionDate.Equal(responseSubscription.SubscriptionDate))

	mockSubscriptionRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)

}

// TestPostSubscriptionBadRequest tests if PostSubscription returns 400-Bad Request when request body is invalid
func TestPostSubscriptionBadRequest(t *testing.T) {
	invalidBodies := []string{
		"",                          // empty body
		`{"invalidField": "value"}`, // invalid body
		`{"following": ""}`,         // empty following
	}

	for _, body := range invalidBodies {
		// Arrange
		mockSubscriptionRepo := new(repositories.MockSubscriptionRepository)
		mockUserRepo := new(repositories.MockUserRepository)

		subscriptionService := services.NewSubscriptionService(mockSubscriptionRepo, mockUserRepo)
		subscriptionController := controllers.NewSubscriptionController(subscriptionService)

		currentUsername := "testUser"
		authenticationToken, err := utils.GenerateAccessToken(currentUsername)
		if err != nil {
			t.Error(err)
		}

		// Setup HTTP request and recorder
		req, _ := http.NewRequest(http.MethodPost, "/subscriptions", bytes.NewBuffer([]byte(body)))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+authenticationToken)
		w := httptest.NewRecorder()

		// Act
		gin.SetMode(gin.TestMode)
		router := gin.Default()
		router.POST("/subscriptions", middleware.AuthorizeUser, subscriptionController.PostSubscription)
		router.ServeHTTP(w, req)

		// Assert Response
		assert.Equal(t, http.StatusBadRequest, w.Code) // Expect HTTP 400 Bad Request status

		var errorResponse customerrors.ErrorResponse
		err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
		assert.NoError(t, err)

		expectedCustomError := customerrors.BadRequest
		assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
		assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)

		mockSubscriptionRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	}
}

// TestPostSubscriptionUnauthorized tests if PostSubscription returns 401-Unauthorized when user is not authenticated
func TestPostSubscriptionUnauthorized(t *testing.T) {
	invalidTokens := []string{
		"",
		"invalid",
	}

	for _, token := range invalidTokens {
		// Arrange
		mockSubscriptionRepo := new(repositories.MockSubscriptionRepository)
		mockUserRepo := new(repositories.MockUserRepository)

		subscriptionService := services.NewSubscriptionService(mockSubscriptionRepo, mockUserRepo)
		subscriptionController := controllers.NewSubscriptionController(subscriptionService)

		// Setup HTTP request and recorder
		req, _ := http.NewRequest(http.MethodPost, "/subscriptions", nil)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		// Act
		gin.SetMode(gin.TestMode)
		router := gin.Default()
		router.POST("/subscriptions", middleware.AuthorizeUser, subscriptionController.PostSubscription)
		router.ServeHTTP(w, req)

		// Assert Response
		assert.Equal(t, http.StatusUnauthorized, w.Code) // Expect HTTP 401 Unauthorized status

		var errorResponse customerrors.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
		assert.NoError(t, err)

		expectedCustomError := customerrors.UserUnauthorized
		assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
		assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)

		mockSubscriptionRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	}
}

// TestPostSubscriptionAlreadyExists tests if PostSubscription returns 409-Conflict when user is already following
func TestPostSubscriptionAlreadyExists(t *testing.T) {
	// Arrange
	mockSubscriptionRepo := new(repositories.MockSubscriptionRepository)
	mockUserRepo := new(repositories.MockUserRepository)

	subscriptionService := services.NewSubscriptionService(mockSubscriptionRepo, mockUserRepo)
	subscriptionController := controllers.NewSubscriptionController(subscriptionService)

	currentUsername := "testUser"
	authenticationToken, err := utils.GenerateAccessToken(currentUsername)
	if err != nil {
		t.Error(err)
	}

	subscriptionCreateRequest := models.SubscriptionPostRequestDTO{
		Following: "testUser2",
	}

	// Mock expectations
	mockUserRepo.On("FindUserByUsername", subscriptionCreateRequest.Following).Return(&models.User{}, nil)                                          // Expect user to be found
	mockSubscriptionRepo.On("GetSubscriptionByUsernames", currentUsername, subscriptionCreateRequest.Following).Return(&models.Subscription{}, nil) // Expect user is already following

	// Setup HTTP request and recorder
	requestBody, err := json.Marshal(subscriptionCreateRequest)
	if err != nil {
		t.Fatal(err)
	}
	req, _ := http.NewRequest(http.MethodPost, "/subscriptions", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+authenticationToken)
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/subscriptions", middleware.AuthorizeUser, subscriptionController.PostSubscription)
	router.ServeHTTP(w, req)

	// Assert Response
	assert.Equal(t, http.StatusConflict, w.Code) // Expect HTTP 409 Conflict status

	var errorResponse customerrors.ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)

	expectedCustomError := customerrors.SubscriptionAlreadyExists
	assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
	assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)

	mockUserRepo.AssertExpectations(t)
	mockSubscriptionRepo.AssertExpectations(t)
}

// TestPostSubscriptionSelfFollow tests if PostSubscription returns 406-Not Acceptable when user tries to follow himself
func TestPostSubscriptionSelfFollow(t *testing.T) {
	// Arrange
	mockSubscriptionRepo := new(repositories.MockSubscriptionRepository)
	mockUserRepo := new(repositories.MockUserRepository)

	subscriptionService := services.NewSubscriptionService(mockSubscriptionRepo, mockUserRepo)
	subscriptionController := controllers.NewSubscriptionController(subscriptionService)

	currentUsername := "testUser"
	authenticationToken, err := utils.GenerateAccessToken(currentUsername)
	if err != nil {
		t.Error(err)
	}

	subscriptionCreateRequest := models.SubscriptionPostRequestDTO{
		Following: currentUsername,
	}

	// Setup HTTP request and recorder
	requestBody, err := json.Marshal(subscriptionCreateRequest)
	if err != nil {
		t.Fatal(err)
	}
	req, _ := http.NewRequest(http.MethodPost, "/subscriptions", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+authenticationToken)
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/subscriptions", middleware.AuthorizeUser, subscriptionController.PostSubscription)
	router.ServeHTTP(w, req)

	// Assert Response
	assert.Equal(t, http.StatusNotAcceptable, w.Code) // Expect HTTP 406 Not Acceptable status

	var errorResponse customerrors.ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)

	expectedCustomError := customerrors.SelfFollow
	assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
	assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)

	mockSubscriptionRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

// TestDeleteSubscriptionSuccess tests if DeleteSubscription returns 204-No Content when subscription is deleted successfully
func TestDeleteSubscriptionSuccess(t *testing.T) {
	// Arrange
	mockSubscriptionRepo := new(repositories.MockSubscriptionRepository)
	mockUserRepo := new(repositories.MockUserRepository)

	subscriptionService := services.NewSubscriptionService(mockSubscriptionRepo, mockUserRepo)
	subscriptionController := controllers.NewSubscriptionController(subscriptionService)

	currentUsername := "testUser"
	authenticationToken, err := utils.GenerateAccessToken(currentUsername)
	if err != nil {
		t.Error(err)
	}

	subscription := models.Subscription{
		Id:                uuid.New(),
		SubscriptionDate:  time.Now(),
		FollowerUsername:  currentUsername,
		FollowingUsername: "testUser2",
	}

	// Mock expectations
	mockSubscriptionRepo.On("GetSubscriptionById", subscription.Id.String()).Return(&subscription, nil) // Expect subscription to be found
	mockSubscriptionRepo.On("DeleteSubscription", subscription.Id.String()).Return(nil)                 // Expect subscription to be deleted

	// Setup HTTP request and recorder
	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/subscriptions/%s", subscription.Id), nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authenticationToken))
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.DELETE("/subscriptions/:subscriptionId", middleware.AuthorizeUser, subscriptionController.DeleteSubscription)
	router.ServeHTTP(w, req)

	// Assert Response
	assert.Equal(t, http.StatusNoContent, w.Code) // Expect HTTP 204 No Content status

	mockSubscriptionRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

// TestDeleteSubscriptionUnauthorized tests if DeleteSubscription returns 401-Unauthorized when user is not authenticated
func TestDeleteSubscriptionUnauthorized(t *testing.T) {
	invalidTokens := []string{
		"",
		"invalid",
	}

	for _, token := range invalidTokens {
		// Arrange
		mockSubscriptionRepo := new(repositories.MockSubscriptionRepository)
		mockUserRepo := new(repositories.MockUserRepository)

		subscriptionService := services.NewSubscriptionService(mockSubscriptionRepo, mockUserRepo)
		subscriptionController := controllers.NewSubscriptionController(subscriptionService)

		// Setup HTTP request and recorder
		req, _ := http.NewRequest(http.MethodDelete, "/subscriptions/123", nil)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		// Act
		gin.SetMode(gin.TestMode)
		router := gin.Default()
		router.DELETE("/subscriptions/:subscriptionId", middleware.AuthorizeUser, subscriptionController.DeleteSubscription)
		router.ServeHTTP(w, req)

		// Assert Response
		assert.Equal(t, http.StatusUnauthorized, w.Code) // Expect HTTP 401 Unauthorized status

		var errorResponse customerrors.ErrorResponse
		err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
		assert.NoError(t, err)

		expectedCustomError := customerrors.UserUnauthorized
		assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
		assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)

		mockSubscriptionRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	}
}

// TestDeleteSubscriptionNotFound tests if DeleteSubscription returns 404-Not Found when subscription is not found
func TestDeleteSubscriptionNotFound(t *testing.T) {
	// Arrange
	mockSubscriptionRepo := new(repositories.MockSubscriptionRepository)
	mockUserRepo := new(repositories.MockUserRepository)

	subscriptionService := services.NewSubscriptionService(mockSubscriptionRepo, mockUserRepo)
	subscriptionController := controllers.NewSubscriptionController(subscriptionService)

	currentUsername := "testUser"
	authenticationToken, err := utils.GenerateAccessToken(currentUsername)
	if err != nil {
		t.Error(err)
	}

	subscriptionId := uuid.New()

	// Mock expectations
	mockSubscriptionRepo.On("GetSubscriptionById", subscriptionId.String()).Return(&models.Subscription{}, gorm.ErrRecordNotFound) // Expect subscription to not be found

	// Setup HTTP request and recorder
	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/subscriptions/%s", subscriptionId), nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authenticationToken))
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.DELETE("/subscriptions/:subscriptionId", middleware.AuthorizeUser, subscriptionController.DeleteSubscription)
	router.ServeHTTP(w, req)

	// Assert Response
	assert.Equal(t, http.StatusNotFound, w.Code) // Expect HTTP 404 Not Found status

	var errorResponse customerrors.ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)

	expectedCustomError := customerrors.SubscriptionNotFound
	assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
	assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)

	mockSubscriptionRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
}

// TestDeleteSubscriptionForbidden tests if DeleteSubscription returns 403-PostDeleteForbidden when user tries to delete a subscription of another user
func TestDeleteSubscriptionForbidden(t *testing.T) {
	// Arrange
	mockSubscriptionRepo := new(repositories.MockSubscriptionRepository)
	mockUserRepo := new(repositories.MockUserRepository)

	subscriptionService := services.NewSubscriptionService(mockSubscriptionRepo, mockUserRepo)
	subscriptionController := controllers.NewSubscriptionController(subscriptionService)

	currentUsername := "testUser"
	authenticationToken, err := utils.GenerateAccessToken(currentUsername)
	if err != nil {
		t.Fatal(err)
	}

	subscription := models.Subscription{
		Id:                uuid.New(),
		SubscriptionDate:  time.Now(),
		FollowerUsername:  "testUser2",
		FollowingUsername: "testUser3",
	}

	// Mock expectations
	mockSubscriptionRepo.On("GetSubscriptionById", subscription.Id.String()).Return(&subscription, nil) // Expect subscription to be found

	// Setup HTTP request and recorder
	req, _ := http.NewRequest(http.MethodDelete, fmt.Sprintf("/subscriptions/%s", subscription.Id), nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authenticationToken))
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.DELETE("/subscriptions/:subscriptionId", middleware.AuthorizeUser, subscriptionController.DeleteSubscription)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusForbidden, w.Code) // Expect HTTP 403 PostDeleteForbidden status

	var errorResponse customerrors.ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)

	expectedCustomError := customerrors.SubscriptionDeleteNotAuthorized
	assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
	assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)

	mockSubscriptionRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)

}

func TestGetSubscriptionsFollowerSuccess(t *testing.T) {
	// Setup Mocks
	// Arrange
	mockSubscriptionRepo := new(repositories.MockSubscriptionRepository)
	mockUserRepo := new(repositories.MockUserRepository)

	subscriptionService := services.NewSubscriptionService(mockSubscriptionRepo, mockUserRepo)
	subscriptionController := controllers.NewSubscriptionController(subscriptionService)

	currentUsername := "testUser"
	authenticationToken, err := utils.GenerateAccessToken(currentUsername)
	if err != nil {
		t.Error(err)
	}

	id1 := uuid.New()
	id2 := uuid.New()
	foundSubscriptions := []models.SubscriptionSearchRecordDTO{
		{
			SubscriptionId:   id1,
			SubscriptionDate: time.Date(2024, time.January, 24, 15, 42, 58, 0, time.UTC),
			User: models.UserSearchRecordDTO{
				Username:          "theo",
				Nickname:          "theotester",
				ProfilePictureUrl: "",
			},
		},
		{
			SubscriptionId:   id2,
			SubscriptionDate: time.Date(2024, time.January, 23, 15, 42, 58, 0, time.UTC),
			User: models.UserSearchRecordDTO{
				Username:          "tina",
				Nickname:          "tinatester",
				ProfilePictureUrl: "",
			},
		},
	}

	ftype := "follower"
	limit := 10
	offset := 0

	// Mock Erwartungen
	mockSubscriptionRepo.On("GetFollower", limit, offset, currentUsername).Return(foundSubscriptions, int64(len(foundSubscriptions)), nil)

	// Setup HTTP request und recorder
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("/subscriptions/%s?type=%s&limit=%d&offset=%d", currentUsername, ftype, limit, offset), nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", authenticationToken))
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/subscriptions/:username", subscriptionController.GetSubscriptions)
	router.ServeHTTP(w, req)

	// Assert Response
	assert.Equal(t, http.StatusOK, w.Code) // Erwarte HTTP 200 OK Status
	var responseDto models.SubscriptionSearchResponseDTO
	err = json.Unmarshal(w.Body.Bytes(), &responseDto)
	assert.NoError(t, err)

	// Hier weitere Überprüfungen der responseDto Daten ...

	mockSubscriptionRepo.AssertExpectations(t)
}
