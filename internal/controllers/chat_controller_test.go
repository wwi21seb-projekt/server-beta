package controllers_test

import (
	"bytes"
	"encoding/json"
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
	"os"
	"strings"
	"testing"
	"time"
)

// TestCreateChatSuccess tests the CreateChat function if it returns 201 Created after creating a chat successfully
func TestCreateChatSuccess(t *testing.T) {
	// Arrange
	mockUserRepo := new(repositories.MockUserRepository)
	mockChatRepo := new(repositories.MockChatRepository)
	mockNotificationRepo := new(repositories.MockNotificationRepository)
	mockPushSubscriptionRepo := new(repositories.MockPushSubscriptionRepository)
	pushSubscriptionService := services.NewPushSubscriptionService(mockPushSubscriptionRepo)
	notificationService := services.NewNotificationService(mockNotificationRepo, pushSubscriptionService)
	chatService := services.NewChatService(mockChatRepo, mockUserRepo, notificationService)
	chatController := controllers.NewChatController(chatService)

	currentUser := &models.User{
		Username: "testUser",
	}
	otherUser := &models.User{
		Username: "testUser2",
	}

	chatCreateRequest := models.ChatCreateRequestDTO{
		Username: "testUser2",
		Content:  "Hello",
	}

	authenticationToken, err := utils.GenerateAccessToken(currentUser.Username)
	if err != nil {
		t.Fatal(err)
	}

	// Mock expectations
	var capturedChat models.Chat
	var capturedMessage models.Message
	var capturedNotification *models.Notification
	mockUserRepo.On("FindUserByUsername", chatCreateRequest.Username).Return(otherUser, nil)
	mockUserRepo.On("FindUserByUsername", currentUser.Username).Return(currentUser, nil)
	mockChatRepo.On("GetChatByUsernames", currentUser.Username, chatCreateRequest.Username).Return(models.Chat{}, gorm.ErrRecordNotFound)
	mockChatRepo.On("CreateChatWithFirstMessage", mock.AnythingOfType("models.Chat"), mock.AnythingOfType("models.Message")).
		Run(func(args mock.Arguments) {
			capturedChat = args.Get(0).(models.Chat)
			capturedMessage = args.Get(1).(models.Message)
		}).Return(nil)
	mockNotificationRepo.On("CreateNotification", mock.AnythingOfType("*models.Notification")).
		Run(func(args mock.Arguments) {
			capturedNotification = args.Get(0).(*models.Notification)
		}).Return(nil)
	mockNotificationRepo.On("GetNotificationById", mock.AnythingOfType("string")).Return(models.Notification{}, nil)
	mockPushSubscriptionRepo.On("GetPushSubscriptionsByUsername", otherUser.Username).Return([]models.PushSubscription{}, nil)

	// Setup HTTP request
	requestBody, err := json.Marshal(chatCreateRequest)
	if err != nil {
		t.Fatal(err)
	}
	req, _ := http.NewRequest("POST", "/chats", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+authenticationToken)
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/chats", middleware.AuthorizeUser, chatController.CreateChat)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusCreated, w.Code) // Expect 201 Created
	var response models.ChatCreateResponseDTO
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.NotNil(t, capturedChat)
	assert.NotEmpty(t, capturedChat.Id)
	assert.NotEmpty(t, capturedChat.CreatedAt)

	assert.NotNil(t, capturedMessage)
	assert.NotEmpty(t, capturedMessage.Id)
	assert.Equal(t, capturedChat.Id, capturedMessage.ChatId)
	assert.Equal(t, currentUser.Username, capturedMessage.Username)
	assert.Equal(t, chatCreateRequest.Content, capturedMessage.Content)

	// Check if users are in chat list
	assert.Len(t, capturedChat.Users, 2)
	assert.Contains(t, capturedChat.Users, *currentUser)
	assert.Contains(t, capturedChat.Users, *otherUser)

	assert.Equal(t, capturedChat.Id.String(), response.ChatId)
	assert.Equal(t, chatCreateRequest.Content, response.Message.Content)
	assert.Equal(t, currentUser.Username, response.Message.Username)
	assert.True(t, response.Message.CreationDate.Equal(capturedMessage.CreatedAt))

	assert.NotNil(t, capturedNotification)
	assert.NotEmpty(t, capturedNotification.Id)
	assert.Equal(t, "message", capturedNotification.NotificationType)
	assert.Equal(t, otherUser.Username, capturedNotification.ForUsername)
	assert.Equal(t, currentUser.Username, capturedNotification.FromUsername)
	assert.NotEmpty(t, capturedNotification.Timestamp)

	mockChatRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
	mockPushSubscriptionRepo.AssertExpectations(t)
	mockNotificationRepo.AssertExpectations(t)
}

// TestCreateChatBadRequest tests the CreateChat function if it returns 400 Bad Request if the request body is invalid
func TestCreateChatBadRequest(t *testing.T) {
	invalidBodies := []string{
		`{}`,                        // Empty body
		`{"username": "testUser2"}`, // Missing content
		`{"content": "Hello"}`,      // Missing username
		`{"username": "testUser2", "content": "` + strings.Repeat("A", 257) + `"}`, // content too long
	}

	for _, body := range invalidBodies {
		// Arrange
		mockUserRepo := new(repositories.MockUserRepository)
		mockChatRepo := new(repositories.MockChatRepository)
		mockNotificationRepo := new(repositories.MockNotificationRepository)
		mockPushSubscriptionRepo := new(repositories.MockPushSubscriptionRepository)
		pushSubscriptionService := services.NewPushSubscriptionService(mockPushSubscriptionRepo)
		notificationService := services.NewNotificationService(mockNotificationRepo, pushSubscriptionService)
		chatService := services.NewChatService(mockChatRepo, mockUserRepo, notificationService)
		chatController := controllers.NewChatController(chatService)

		currentUser := &models.User{
			Username: "testUser",
		}

		authenticationToken, err := utils.GenerateAccessToken(currentUser.Username)
		if err != nil {
			t.Fatal(err)
		}

		// Setup HTTP request
		req, _ := http.NewRequest("POST", "/chats", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+authenticationToken)
		w := httptest.NewRecorder()

		// Act
		gin.SetMode(gin.TestMode)
		router := gin.Default()
		router.POST("/chats", middleware.AuthorizeUser, chatController.CreateChat)
		router.ServeHTTP(w, req)

		// Assert
		assert.Equal(t, http.StatusBadRequest, w.Code) // Expect 400 Bad Request
		var errorResponse customerrors.ErrorResponse
		err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
		assert.NoError(t, err)

		expectedCustomError := customerrors.BadRequest
		assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
		assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)

		mockChatRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
		mockPushSubscriptionRepo.AssertExpectations(t)
		mockNotificationRepo.AssertExpectations(t)
	}
}

// TestCreateChatUnauthorized tests the CreateChat function if it returns 401 Unauthorized if the user is not logged in
func TestCreateChatUnauthorized(t *testing.T) {
	// Arrange
	mockUserRepo := new(repositories.MockUserRepository)
	mockChatRepo := new(repositories.MockChatRepository)
	mockNotificationRepo := new(repositories.MockNotificationRepository)
	mockPushSubscriptionRepo := new(repositories.MockPushSubscriptionRepository)
	pushSubscriptionService := services.NewPushSubscriptionService(mockPushSubscriptionRepo)
	notificationService := services.NewNotificationService(mockNotificationRepo, pushSubscriptionService)
	chatService := services.NewChatService(mockChatRepo, mockUserRepo, notificationService)
	chatController := controllers.NewChatController(chatService)

	// Setup HTTP request
	req, _ := http.NewRequest("POST", "/chats", nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/chats", middleware.AuthorizeUser, chatController.CreateChat)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code) // Expect 401 Unauthorized
	var errorResponse customerrors.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)

	expectedCustomError := customerrors.UserUnauthorized
	assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
	assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)

	mockChatRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
	mockPushSubscriptionRepo.AssertExpectations(t)
	mockNotificationRepo.AssertExpectations(t)
}

// TestCreateChatUserNotFound tests the CreateChat function if it returns 404 Not Found if the other user does not exist
func TestCreateChatUserNotFound(t *testing.T) {
	// Arrange
	mockUserRepo := new(repositories.MockUserRepository)
	mockChatRepo := new(repositories.MockChatRepository)
	mockNotificationRepo := new(repositories.MockNotificationRepository)
	mockPushSubscriptionRepo := new(repositories.MockPushSubscriptionRepository)
	pushSubscriptionService := services.NewPushSubscriptionService(mockPushSubscriptionRepo)
	notificationService := services.NewNotificationService(mockNotificationRepo, pushSubscriptionService)
	chatService := services.NewChatService(mockChatRepo, mockUserRepo, notificationService)
	chatController := controllers.NewChatController(chatService)

	currentUser := &models.User{
		Username: "testUser",
	}

	chatCreateRequest := models.ChatCreateRequestDTO{
		Username: "testUser2",
		Content:  "Hello",
	}

	authenticationToken, err := utils.GenerateAccessToken(currentUser.Username)
	if err != nil {
		t.Fatal(err)
	}

	// Mock expectations
	mockUserRepo.On("FindUserByUsername", chatCreateRequest.Username).Return(&models.User{}, gorm.ErrRecordNotFound)

	// Setup HTTP request
	requestBody, err := json.Marshal(chatCreateRequest)
	if err != nil {
		t.Fatal(err)
	}
	req, _ := http.NewRequest("POST", "/chats", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+authenticationToken)
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/chats", middleware.AuthorizeUser, chatController.CreateChat)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, w.Code) // Expect 404 Not Found
	var errorResponse customerrors.ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)

	expectedCustomError := customerrors.UserNotFound
	assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
	assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)

	mockChatRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
	mockPushSubscriptionRepo.AssertExpectations(t)
	mockNotificationRepo.AssertExpectations(t)
}

// TestCreateChatChatAlreadyExists tests the CreateChat function if it returns 409 Conflict if the chat already exists
func TestCreateChatChatAlreadyExists(t *testing.T) {
	// Arrange
	mockUserRepo := new(repositories.MockUserRepository)
	mockChatRepo := new(repositories.MockChatRepository)
	mockNotificationRepo := new(repositories.MockNotificationRepository)
	mockPushSubscriptionRepo := new(repositories.MockPushSubscriptionRepository)
	pushSubscriptionService := services.NewPushSubscriptionService(mockPushSubscriptionRepo)
	notificationService := services.NewNotificationService(mockNotificationRepo, pushSubscriptionService)
	chatService := services.NewChatService(mockChatRepo, mockUserRepo, notificationService)
	chatController := controllers.NewChatController(chatService)

	currentUser := &models.User{
		Username: "testUser",
	}
	otherUser := &models.User{
		Username: "testUser2",
	}

	chatCreateRequest := models.ChatCreateRequestDTO{
		Username: "testUser2",
		Content:  "Hello",
	}

	authenticationToken, err := utils.GenerateAccessToken(currentUser.Username)
	if err != nil {
		t.Fatal(err)
	}

	// Mock expectations
	mockUserRepo.On("FindUserByUsername", chatCreateRequest.Username).Return(otherUser, nil)
	mockChatRepo.On("GetChatByUsernames", currentUser.Username, chatCreateRequest.Username).Return(models.Chat{}, nil)

	// Setup HTTP request
	requestBody, err := json.Marshal(chatCreateRequest)
	if err != nil {
		t.Fatal(err)
	}
	req, _ := http.NewRequest("POST", "/chats", bytes.NewBuffer(requestBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+authenticationToken)
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.POST("/chats", middleware.AuthorizeUser, chatController.CreateChat)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusConflict, w.Code) // Expect 409 Conflict
	var errorResponse customerrors.ErrorResponse
	err = json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)

	expectedCustomError := customerrors.ChatAlreadyExists
	assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
	assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)

	mockChatRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
	mockPushSubscriptionRepo.AssertExpectations(t)
	mockNotificationRepo.AssertExpectations(t)
}

// TestGetChatsSuccess tests the GetChats function if it returns 200 OK after successfully retrieving chats
func TestGetChatsSuccess(t *testing.T) {
	// Arrange
	mockUserRepo := new(repositories.MockUserRepository)
	mockChatRepo := new(repositories.MockChatRepository)
	mockNotificationRepo := new(repositories.MockNotificationRepository)
	mockPushSubscriptionRepo := new(repositories.MockPushSubscriptionRepository)
	pushSubscriptionService := services.NewPushSubscriptionService(mockPushSubscriptionRepo)
	notificationService := services.NewNotificationService(mockNotificationRepo, pushSubscriptionService)
	chatService := services.NewChatService(mockChatRepo, mockUserRepo, notificationService)
	chatController := controllers.NewChatController(chatService)

	currentUsername := "myUser"
	authenticationToken, err := utils.GenerateAccessToken(currentUsername)
	if err != nil {
		t.Fatal(err)
	}

	imageId := uuid.New()
	imageFormat := "png"
	err = os.Setenv("SERVER_URL", "https://example.com")
	if err != nil {
		t.Fatal(err)
	}
	expectedImageUrl := os.Getenv("SERVER_URL") + "/api/images/" + imageId.String() + "." + imageFormat

	chats := []models.Chat{
		{
			Id: uuid.New(),
			Users: []models.User{
				{
					Username: "testUser2",
					Nickname: "Test User 2",
					ImageId:  &imageId,
					Image: models.Image{
						Id:     imageId,
						Format: imageFormat,
						Width:  100,
						Height: 200,
						Tag:    time.Now().UTC(),
					},
				},
			},
			CreatedAt: time.Now().UTC(),
		},
		{
			Id: uuid.New(),
			Users: []models.User{
				{
					Username: "testUser3",
					Nickname: "Test User 3",
				},
			},
			CreatedAt: time.Now().UTC(),
		},
	}

	// Mock expectations
	mockChatRepo.On("GetChatsByUsername", currentUsername).Return(chats, nil)

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
	router.GET("/chats", chatController.GetChats)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code) // Expect 200 OK
	var response models.ChatsResponseDTO
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Len(t, response.Records, 2)
	for i, chat := range chats {
		assert.Equal(t, chat.Id.String(), response.Records[i].ChatId)
		assert.Equal(t, chat.Users[0].Username, response.Records[i].User.Username)
		assert.Equal(t, chat.Users[0].Nickname, response.Records[i].User.Nickname)

		if chat.Users[0].ImageId != nil {
			assert.NotNil(t, response.Records[i].User.Picture)
			assert.Equal(t, expectedImageUrl, response.Records[i].User.Picture.Url)
			assert.Equal(t, chat.Users[0].Image.Width, response.Records[i].User.Picture.Width)
			assert.Equal(t, chat.Users[0].Image.Height, response.Records[i].User.Picture.Height)
			assert.True(t, chat.Users[0].Image.Tag.Equal(response.Records[i].User.Picture.Tag))
		} else {
			assert.Nil(t, response.Records[i].User.Picture)
		}
	}

	mockChatRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
	mockPushSubscriptionRepo.AssertExpectations(t)
	mockNotificationRepo.AssertExpectations(t)
}

// TestGetChatsUnauthorized tests the GetChats function if it returns 401 Unauthorized when the user is not authenticated
func TestGetChatsUnauthorized(t *testing.T) {
	// Arrange
	mockUserRepo := new(repositories.MockUserRepository)
	mockChatRepo := new(repositories.MockChatRepository)
	mockNotificationRepo := new(repositories.MockNotificationRepository)
	mockPushSubscriptionRepo := new(repositories.MockPushSubscriptionRepository)
	pushSubscriptionService := services.NewPushSubscriptionService(mockPushSubscriptionRepo)
	notificationService := services.NewNotificationService(mockNotificationRepo, pushSubscriptionService)
	chatService := services.NewChatService(mockChatRepo, mockUserRepo, notificationService)
	chatController := controllers.NewChatController(chatService)

	// Setup HTTP request
	url := "/chats"
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Act
	gin.SetMode(gin.TestMode)
	router := gin.Default()
	router.GET("/chats", middleware.AuthorizeUser, chatController.GetChats)
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code) // Expect 401 Unauthorized
	var errorResponse customerrors.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)

	expectedCustomError := customerrors.UserUnauthorized
	assert.Equal(t, expectedCustomError.Message, errorResponse.Error.Message)
	assert.Equal(t, expectedCustomError.Code, errorResponse.Error.Code)

	mockChatRepo.AssertExpectations(t)
	mockUserRepo.AssertExpectations(t)
	mockPushSubscriptionRepo.AssertExpectations(t)
	mockNotificationRepo.AssertExpectations(t)
}
