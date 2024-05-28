package controllers_test

import (
	"encoding/json"
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
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"
)

// TestGetMessagesByChatIdSuccess tests the GetMessagesByChatId function if it returns 200 OK after successfully retrieving messages
func TestGetMessagesByChatIdSuccess(t *testing.T) {
	// Arrange
	mockChatRepository := new(repositories.MockChatRepository)
	mockMessageRepository := new(repositories.MockMessageRepository)
	messageService := services.NewMessageService(mockMessageRepository, mockChatRepository)
	messageController := controllers.NewMessageController(messageService)

	currentUsername := "myUser"
	authenticationToken, err := utils.GenerateAccessToken(currentUsername)
	if err != nil {
		t.Fatal(err)
	}
	otherUsername := "otherUser"

	chatId := uuid.New()
	chat := models.Chat{
		Id: chatId,
		Users: []models.User{
			{Username: currentUsername},
			{Username: otherUsername},
		},
	}
	messages := []models.Message{
		{
			Id:        uuid.New(),
			ChatId:    chatId,
			Username:  currentUsername,
			Content:   "Test message 1",
			CreatedAt: time.Now(),
		},
		{
			Id:       uuid.New(),
			ChatId:   chatId,
			Username: otherUsername,
			Content:  "Test message 2",
		},
	}

	offset := 3
	limit := 2
	totalRecords := int64(200)

	// Mock expectations
	mockChatRepository.On("GetChatById", chatId.String()).Return(chat, nil)
	mockMessageRepository.On("GetMessagesByChatId", chatId.String(), offset, limit).Return(messages, totalRecords, nil)

	// Setup HTTP request
	url := "/chats/" + chatId.String() + "?offset=" + strconv.Itoa(offset) + "&limit=" + strconv.Itoa(limit)
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+authenticationToken)
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/chats/:chatId", middleware.AuthorizeUser, messageController.GetMessagesByChatId)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code) // Expect 200 OK
	var response models.MessagesResponseDTO
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Len(t, response.Records, 2)
	assert.Equal(t, offset, response.Pagination.Offset)
	assert.Equal(t, limit, response.Pagination.Limit)
	assert.Equal(t, totalRecords, response.Pagination.Records)

	for i, message := range messages {
		assert.Equal(t, message.Content, response.Records[i].Content)
		assert.Equal(t, message.Username, response.Records[i].Username)
		assert.True(t, message.CreatedAt.Equal(response.Records[i].CreationDate))
	}

	mockMessageRepository.AssertExpectations(t)
	mockChatRepository.AssertExpectations(t)
}

// TestGetMessagesByChatIdUnauthorized tests the GetMessagesByChatId function if it returns 401 Unauthorized when the user is not authenticated
func TestGetMessagesByChatIdUnauthorized(t *testing.T) {
	// Arrange
	mockChatRepository := new(repositories.MockChatRepository)
	mockMessageRepository := new(repositories.MockMessageRepository)
	messageService := services.NewMessageService(mockMessageRepository, mockChatRepository)
	messageController := controllers.NewMessageController(messageService)

	chatId := uuid.New().String()

	// Setup HTTP request
	url := "/chats/" + chatId
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/chats/:chatId", middleware.AuthorizeUser, messageController.GetMessagesByChatId)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code) // Expect 401 Unauthorized
	var errorResponse customerrors.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)

	expectedCustomError := customerrors.UserUnauthorized
	assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
	assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)

	mockMessageRepository.AssertExpectations(t)
	mockChatRepository.AssertExpectations(t)
}

// TestGetMessagesByChatIdForbidden tests the GetMessagesByChatId function if it returns 403 Forbidden when the user is not part of the chat
func TestGetMessagesByChatIdForbidden(t *testing.T) {
	// Arrange
	mockChatRepository := new(repositories.MockChatRepository)
	mockMessageRepository := new(repositories.MockMessageRepository)
	messageService := services.NewMessageService(mockMessageRepository, mockChatRepository)
	messageController := controllers.NewMessageController(messageService)

	currentUsername := "myUser"
	authenticationToken, err := utils.GenerateAccessToken(currentUsername)
	if err != nil {
		t.Fatal(err)
	}

	chatId := uuid.New()
	chat := models.Chat{
		Id: chatId,
		Users: []models.User{
			{
				Username: "testUser2",
			},
			{
				Username: "testUser3",
			},
		},
	}

	// Mock expectations
	mockChatRepository.On("GetChatById", chatId.String()).Return(chat, nil)

	// Setup HTTP request
	url := "/chats/" + chatId.String()
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+authenticationToken)
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/chats/:chatId", middleware.AuthorizeUser, messageController.GetMessagesByChatId)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusForbidden, w.Code) // Expect 403 Forbidden
	var errorResponse customerrors.ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)

	expectedCustomError := customerrors.NotChatParticipant
	assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
	assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)

	mockMessageRepository.AssertExpectations(t)
	mockChatRepository.AssertExpectations(t)
}
