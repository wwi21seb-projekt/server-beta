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
	mockNotificationRepository := new(repositories.MockNotificationRepository)
	mockPushSubscriptionRepository := new(repositories.MockPushSubscriptionRepository)
	pushSubscriptionService := services.NewPushSubscriptionService(mockPushSubscriptionRepository)
	notificationService := services.NewNotificationService(mockNotificationRepository, pushSubscriptionService)
	messageService := services.NewMessageService(mockMessageRepository, mockChatRepository, notificationService)
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
	mockPushSubscriptionRepository.AssertExpectations(t)
	mockNotificationRepository.AssertExpectations(t)
}

// TestGetMessagesByChatIdUnauthorized tests the GetMessagesByChatId function if it returns 401 Unauthorized when the user is not authenticated
func TestGetMessagesByChatIdUnauthorized(t *testing.T) {
	// Arrange
	mockChatRepository := new(repositories.MockChatRepository)
	mockMessageRepository := new(repositories.MockMessageRepository)
	mockNotificationRepository := new(repositories.MockNotificationRepository)
	mockPushSubscriptionRepository := new(repositories.MockPushSubscriptionRepository)
	pushSubscriptionService := services.NewPushSubscriptionService(mockPushSubscriptionRepository)
	notificationService := services.NewNotificationService(mockNotificationRepository, pushSubscriptionService)
	messageService := services.NewMessageService(mockMessageRepository, mockChatRepository, notificationService)
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
	mockPushSubscriptionRepository.AssertExpectations(t)
	mockNotificationRepository.AssertExpectations(t)
}

// TestGetMessagesByChatIdNoParticipant tests the GetMessagesByChatId function if it returns 404 Not Found when the user is not part of the chat
func TestGetMessagesByChatIdNoParticipant(t *testing.T) {
	// Arrange
	mockChatRepository := new(repositories.MockChatRepository)
	mockMessageRepository := new(repositories.MockMessageRepository)
	mockNotificationRepository := new(repositories.MockNotificationRepository)
	mockPushSubscriptionRepository := new(repositories.MockPushSubscriptionRepository)
	pushSubscriptionService := services.NewPushSubscriptionService(mockPushSubscriptionRepository)
	notificationService := services.NewNotificationService(mockNotificationRepository, pushSubscriptionService)
	messageService := services.NewMessageService(mockMessageRepository, mockChatRepository, notificationService)
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
	assert.Equal(t, http.StatusNotFound, w.Code) // Expect 404 Not Found
	var errorResponse customerrors.ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)

	expectedCustomError := customerrors.ChatNotFound
	assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
	assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)

	mockMessageRepository.AssertExpectations(t)
	mockChatRepository.AssertExpectations(t)
	mockPushSubscriptionRepository.AssertExpectations(t)
	mockNotificationRepository.AssertExpectations(t)
}

//// TestHandleWebSocketSuccess tests the HandleWebSocket establishes a connection and is able to send and receive messages
//func TestHandleWebSocketSuccess(t *testing.T) {
//	// Arrange
//	mockChatRepository := new(repositories.MockChatRepository)
//	mockMessageRepository := new(repositories.MockMessageRepository)
//	mockNotificationRepository := new(repositories.MockNotificationRepository)
//	mockPushSubscriptionRepository := new(repositories.MockPushSubscriptionRepository)
//	pushSubscriptionService := services.NewPushSubscriptionService(mockPushSubscriptionRepository)
//	notificationService := services.NewNotificationService(mockNotificationRepository, pushSubscriptionService)
//	messageService := services.NewMessageService(mockMessageRepository, mockChatRepository, notificationService)
//	messageController := controllers.NewMessageController(messageService)
//
//	currentUsername := "myUser"
//	authenticationToken, err := utils.GenerateAccessToken(currentUsername)
//	if err != nil {
//		t.Fatal(err)
//	}
//	otherUsername := "otherUser"
//
//	chat := models.Chat{
//		Id: uuid.New(),
//		Users: []models.User{
//			{Username: currentUsername},
//			{Username: otherUsername},
//		},
//	}
//
//	// Mock expectations
//	var capturedMessage models.Message
//	mockChatRepository.On("GetChatById", chat.Id.String()).Return(chat, nil) // Expect chat to be found
//	mockMessageRepository.On("GetMessagesByChatId", chat.Id.String(), 0, 1).Return([]models.Message{}, int64(0), nil)
//	mockMessageRepository.On("CreateMessage", mock.AnythingOfType("*models.Message")).
//		Run(func(args mock.Arguments) {
//			capturedMessage = args.Get(0).(models.Message)
//		}).Return(nil)
//
//	// Create test server
//	gin.SetMode(gin.TestMode)
//	router := gin.Default()
//	router.GET("/chat", messageController.HandleWebSocket)
//	server := httptest.NewServer(router)
//	defer server.Close()
//
//	// Create WebSocket connection
//	url := "ws" + server.URL[4:] + "/chat?chatId=" + chat.Id.String()
//	headers := http.Header{"Sec-WebSocket-Protocol": authenticationToken}
//	ws, _, err := websocket.DefaultDialer.Dial(url, headers)
//	assert.NoError(t, err)
//
//	// Send message
//	message := models.MessageCreateRequestDTO{
//		Content: "Test message",
//	}
//	messageJSON, err := json.Marshal(message)
//	assert.NoError(t, err)
//
//	err = ws.WriteMessage(websocket.TextMessage, messageJSON)
//	assert.NoError(t, err)
//
//	// Read message
//	_, receivedMessage, err := ws.ReadMessage()
//	assert.NoError(t, err)
//	var response models.MessageRecordDTO
//	err = json.Unmarshal(receivedMessage, &response)
//	assert.NoError(t, err)
//
//	// Assert
//	assert.Equal(t, message.Content, response.Content)
//	assert.Equal(t, currentUsername, response.Username)
//	assert.NotEmpty(t, response.CreationDate)
//
//	assert.Equal(t, message.Content, capturedMessage.Content)
//	assert.Equal(t, currentUsername, capturedMessage.Username)
//	assert.Equal(t, chat.Id, capturedMessage.ChatId)
//	assert.NotEmpty(t, capturedMessage.Id)
//	assert.True(t, capturedMessage.CreatedAt.Equal(response.CreationDate))
//
//	mockMessageRepository.AssertExpectations(t)
//	mockChatRepository.AssertExpectations(t)
//
//}
