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
	"testing"
	"time"
)

// TestGetChatsSuccess tests the GetChats function if it returns 200 OK after successfully retrieving chats
func TestGetChatsSuccess(t *testing.T) {
	// Arrange
	mockChatRepository := new(repositories.MockChatRepository)
	chatService := services.NewChatService(mockChatRepository)
	chatController := controllers.NewChatController(chatService)

	currentUsername := "myUser"
	authenticationToken, err := utils.GenerateAccessToken(currentUsername)
	if err != nil {
		t.Fatal(err)
	}

	chats := []models.Chat{
		{
			Id: uuid.New(),
			Users: []models.User{
				{Username: "testuser2", Nickname: "Test User 2", ProfilePictureUrl: "https://example.com/testuser2.jpg"},
			},
			CreatedAt: time.Now(),
		},
		{
			Id: uuid.New(),
			Users: []models.User{
				{Username: "testuser3", Nickname: "Test User 3", ProfilePictureUrl: "https://example.com/testuser3.jpg"},
			},
			CreatedAt: time.Now(),
		},
	}

	// Mock expectations
	mockChatRepository.On("GetChatsByUsername", currentUsername).Return(chats, nil)

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
		assert.Equal(t, chat.Users[0].ProfilePictureUrl, response.Records[i].User.ProfilePictureUrl)
	}

	mockChatRepository.AssertExpectations(t)
}

// TestGetChatsUnauthorized tests the GetChats function if it returns 401 Unauthorized when the user is not authenticated
func TestGetChatsUnauthorized(t *testing.T) {
	// Arrange
	mockChatRepository := new(repositories.MockChatRepository)
	chatService := services.NewChatService(mockChatRepository)
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

	mockChatRepository.AssertExpectations(t)
}
