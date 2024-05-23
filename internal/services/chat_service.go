package services

import (
	"errors"
	"github.com/google/uuid"
	"github.com/wwi21seb-projekt/server-beta/internal/customerrors"
	"github.com/wwi21seb-projekt/server-beta/internal/models"
	"github.com/wwi21seb-projekt/server-beta/internal/repositories"
	"gorm.io/gorm"
	"net/http"
	"time"
)

type ChatServiceInterface interface {
	PostChat(postId, currentUsername string) (*customerrors.CustomError, int)
}

type ChatService struct {
	chatRepo    repositories.ChatRepositoryInterface
	userRepo    repositories.UserRepositoryInterface
	messageRepo repositories.MessageRepositoryInterface
}

// NewChatService can be used as a constructor to create a ChatService "object"
func NewChatService(
	chatRepo repositories.ChatRepositoryInterface,
	messageRepo repositories.MessageRepositoryInterface,
	userRepo repositories.UserRepositoryInterface) *ChatService {
	return &ChatService{chatRepo: chatRepo, messageRepo: messageRepo, userRepo: userRepo}
}

// PostChat creates a chat for a given post id and the current logged-in user
func (service *ChatService) PostChat(currentUsername string, chatPostRequestDTO models.ChatPostRequestDTO) (*customerrors.CustomError, int) {

	// Check if chat exists
	chat, err := service.chatRepo.GetChatByUsernames(currentUsername, chatPostRequestDTO.Username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return customerrors.PostNotFound, http.StatusNotFound
		}
		return customerrors.DatabaseError, http.StatusInternalServerError
	}

	currentUser, _ := service.userRepo.FindUserByUsername(currentUsername)
	otherUser, _ := service.userRepo.FindUserByUsername(chatPostRequestDTO.Username)

	// Create chat
	newChat := models.Chat{
		Id:        uuid.New(),
		Users:     []models.User{*currentUser, *otherUser},
		CreatedAt: time.Now(),
	}

	err = service.chatRepo.CreateChat(&newChat)
	if err != nil {
		return customerrors.DatabaseError, http.StatusInternalServerError
	}

	// Create message
	newMessage := models.Message{
		Id:        uuid.New(),
		ChatId:    chat.Id,
		Username:  currentUsername,
		Content:   chatPostRequestDTO.Content,
		CreatedAt: time.Now(),
	}

	err = service.messageRepo.CreateMessage(&newMessage)
	if err != nil {
		return customerrors.DatabaseError, http.StatusInternalServerError
	}

	// Return
	return nil, http.StatusNoContent
}
