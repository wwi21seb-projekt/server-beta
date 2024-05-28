package controllers

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/wwi21seb-projekt/server-beta/internal/customerrors"
	"github.com/wwi21seb-projekt/server-beta/internal/middleware"
	"github.com/wwi21seb-projekt/server-beta/internal/models"
	"github.com/wwi21seb-projekt/server-beta/internal/repositories"
	"github.com/wwi21seb-projekt/server-beta/internal/services"
	"github.com/wwi21seb-projekt/server-beta/internal/utils"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestGetAllChatsSuccess tests the GetAllChats function if it returns 200 OK after successfully retrieving chats
func TestGetAllChatsSuccess(t *testing.T) {
	// Arrange
	mockChatRepository := new(repositories.MockChatRepository)
	chatService := services.NewChatService(mockChatRepository)
	chatController := NewChatController(chatService)

	authenticationToken, err := utils.GenerateAccessToken("myuser")
	if err != nil {
		t.Fatal(err)
	}

	chats := []models.Chat{
		{
			Id: uuid.New(),
			Users: []models.User{
				{Username: "testuser1"},
				{Username: "testuser2"},
			},
			CreatedAt: time.Now(),
		},
	}

	// Mock expectations
	mockChatRepository.On("GetAllChats", "myuser").Return(chats, nil)

	// Setup HTTP request
	url := "/chats"
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+authenticationToken)
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.Use(middleware.AuthorizeUser)
	router.GET("/chats", chatController.GetAllChats)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code) // Expect 200 OK
	var response []models.ChatDTO
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response, 1)
	assert.Equal(t, chats[0].Id, response[0].Id)
	assert.Equal(t, len(chats[0].Users), len(response[0].Users))
	assert.WithinDuration(t, chats[0].CreatedAt, response[0].CreatedAt, time.Second)

	mockChatRepository.AssertExpectations(t)
}

// TestGetAllChatsUnauthorized tests the GetAllChats function if it returns 401 Unauthorized when the user is not authenticated
func TestGetAllChatsUnauthorized(t *testing.T) {
	// Arrange
	mockChatRepository := new(repositories.MockChatRepository)
	chatService := services.NewChatService(mockChatRepository)
	chatController := NewChatController(chatService)

	// Setup HTTP request
	url := "/chats"
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/chats", middleware.AuthorizeUser, chatController.GetAllChats)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code) // Expect 401 Unauthorized
	var errorResponse customerrors.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)

	expectedCustomError := customerrors.UserUnauthorized
	assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
	assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)

	mockChatRepository.AssertExpectations(t)
}

// TestGetChatMessagesSuccess tests the GetChatMessages function if it returns 200 OK after successfully retrieving messages
func TestGetChatMessagesSuccess(t *testing.T) {
	// Arrange
	mockChatRepository := new(repositories.MockChatRepository)
	chatService := services.NewChatService(mockChatRepository)
	chatController := NewChatController(chatService)

	authenticationToken, err := utils.GenerateAccessToken("myuser")
	if err != nil {
		t.Fatal(err)
	}

	chatId := uuid.New().String()
	messages := []models.Message{
		{
			Id:        uuid.New(),
			ChatId:    chatId,
			Username:  "myuser",
			Content:   "Test message 1",
			CreatedAt: time.Now(),
		},
		{
			Id:        uuid.New(),
			ChatId:    chatId,
			Username:  "myuser",
			Content:   "Test message 2",
			CreatedAt: time.Now(),
		},
	}

	// Mock expectations
	mockChatRepository.On("GetChatMessages", chatId, 0, 10).Return(messages, nil)

	// Setup HTTP request
	url := "/chats/" + chatId + "/messages?offset=0&limit=10"
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+authenticationToken)
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/chats/:chatId/messages", middleware.AuthorizeUser, chatController.GetChatMessages)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code) // Expect 200 OK
	var response struct {
		Records    []models.MessageRecordDTO `json:"records"`
		Pagination struct {
			Offset  int `json:"offset"`
			Limit   int `json:"limit"`
			Records int `json:"records"`
		} `json:"pagination"`
	}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Len(t, response.Records, 2)
	assert.Equal(t, 0, response.Pagination.Offset)
	assert.Equal(t, 10, response.Pagination.Limit)
	assert.Equal(t, len(messages), response.Pagination.Records)

	for i, message := range messages {
		assert.Equal(t, message.Id, response.Records[i].Id)
		assert.Equal(t, message.Content, response.Records[i].Content)
		assert.Equal(t, message.Username, response.Records[i].Username)
		assert.True(t, message.CreatedAt.Equal(response.Records[i].CreatedAt))
	}

	mockChatRepository.AssertExpectations(t)
}

// TestGetChatMessagesUnauthorized tests the GetChatMessages function if it returns 401 Unauthorized when the user is not authenticated
func TestGetChatMessagesUnauthorized(t *testing.T) {
	// Arrange
	mockChatRepository := new(repositories.MockChatRepository)
	chatService := services.NewChatService(mockChatRepository)
	chatController := NewChatController(chatService)

	chatId := uuid.New().String()

	// Setup HTTP request
	url := "/chats/" + chatId + "/messages"
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/chats/:chatId/messages", middleware.AuthorizeUser, chatController.GetChatMessages)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code) // Expect 401 Unauthorized
	var errorResponse customerrors.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)

	expectedCustomError := customerrors.UserUnauthorized
	assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
	assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)

	mockChatRepository.AssertExpectations(t)
}

// TestGetChatMessagesNotFound tests the GetChatMessages function if it returns 404 Not Found when the chat does not exist
func TestGetChatMessagesNotFound(t *testing.T) {
	// Arrange
	mockChatRepository := new(repositories.MockChatRepository)
	chatService := services.NewChatService(mockChatRepository)
	chatController := NewChatController(chatService)

	authenticationToken, err := utils.GenerateAccessToken("myuser")
	if err != nil {
		t.Fatal(err)
	}

	chatId := uuid.New().String()

	// Mock expectations
	mockChatRepository.On("GetChatMessages", chatId, 0, 10).Return([]models.Message, customerrors.ChatNotFound)

	// Setup HTTP request
	url := "/chats/" + chatId + "/messages?offset=0&limit=10"
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+authenticationToken)
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/chats/:chatId/messages", middleware.AuthorizeUser, chatController.GetChatMessages)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code) // Expect 404 Not Found
	var errorResponse customerrors.ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)

	expectedCustomError := customerrors.ChatNotFound
	assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
	assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)

	mockChatRepository.AssertExpectations(t)
}
