package services

import (
	"errors"
	"github.com/google/uuid"
	"github.com/microcosm-cc/bluemonday"
	"github.com/wwi21seb-projekt/server-beta/internal/customerrors"
	"github.com/wwi21seb-projekt/server-beta/internal/models"
	"github.com/wwi21seb-projekt/server-beta/internal/repositories"
	"gorm.io/gorm"
	"net/http"
	"strings"
	"time"
)

type ChatServiceInterface interface {
	CreatePost(req *models.ChatCreateRequestDTO, currentUsername string) (*models.ChatCreateResponseDTO, *customerrors.CustomError, int)
}

type ChatService struct {
	chatRepo repositories.ChatRepositoryInterface
	userRepo repositories.UserRepositoryInterface
	policy   *bluemonday.Policy
}

// NewChatService can be used as a constructor to create a ChatService "object"
func NewChatService(
	chatRepo repositories.ChatRepositoryInterface,
	userRepo repositories.UserRepositoryInterface) *ChatService {
	return &ChatService{chatRepo: chatRepo, userRepo: userRepo, policy: bluemonday.UGCPolicy()}
}

// CreatePost creates a chat for a given post id, username and the current logged-in user
func (service *ChatService) CreatePost(req *models.ChatCreateRequestDTO, currentUsername string) (*models.ChatCreateResponseDTO, *customerrors.CustomError, int) {
	// Sanitize message content because it is a free text field
	req.Content = strings.Trim(req.Content, " ") // remove leading and trailing whitespaces
	req.Content = service.policy.Sanitize(req.Content)

	// Validate input
	if len(req.Content) <= 0 || len(req.Content) > 256 {
		return nil, customerrors.BadRequest, http.StatusBadRequest
	}

	// Check if user exists
	otherUser, err := service.userRepo.FindUserByUsername(req.Username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, customerrors.UserNotFound, http.StatusNotFound
		}
	}

	// Check if that chat already exists
	_, err = service.chatRepo.GetChatByUsernames(currentUsername, req.Username)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, customerrors.DatabaseError, http.StatusInternalServerError
	}
	if err == nil {
		return nil, customerrors.ChatAlreadyExists, http.StatusConflict
	}

	// Get current user
	currentUser, err := service.userRepo.FindUserByUsername(currentUsername)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, customerrors.UserUnauthorized, http.StatusUnauthorized
		}
		return nil, customerrors.DatabaseError, http.StatusInternalServerError
	}

	// Create chat with first message
	currentTime := time.Now()
	newChat := models.Chat{
		Id:        uuid.New(),
		Users:     []models.User{*currentUser, *otherUser},
		CreatedAt: currentTime,
	}

	firstMessage := models.Message{
		Id:        uuid.New(),
		ChatId:    newChat.Id,
		Username:  currentUsername,
		Content:   req.Content,
		CreatedAt: currentTime,
	}

	err = service.chatRepo.CreateChatWithFirstMessage(newChat, firstMessage)
	if err != nil {
		return nil, customerrors.DatabaseError, http.StatusInternalServerError
	}

	// Create response
	response := &models.ChatCreateResponseDTO{
		ChatId: newChat.Id.String(),
		Message: &models.FirstMessageResponseDTO{
			Content:      firstMessage.Content,
			Username:     firstMessage.Username,
			CreationDate: firstMessage.CreatedAt,
		},
	}

	return response, nil, http.StatusCreated
}
