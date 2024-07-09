package services

import (
	"errors"
	"github.com/google/uuid"
	"github.com/microcosm-cc/bluemonday"
	"github.com/wwi21seb-projekt/server-beta/internal/customerrors"
	"github.com/wwi21seb-projekt/server-beta/internal/models"
	"github.com/wwi21seb-projekt/server-beta/internal/repositories"
	"github.com/wwi21seb-projekt/server-beta/internal/utils"
	"gorm.io/gorm"
	"net/http"
	"strings"
	"time"
)

type ChatServiceInterface interface {
	CreateChat(req *models.ChatCreateRequestDTO, currentUsername string) (*models.ChatCreateResponseDTO, *customerrors.CustomError, int)
	GetChatsByUsername(username string) (*models.ChatsResponseDTO, *customerrors.CustomError, int)
}

type ChatService struct {
	chatRepo            repositories.ChatRepositoryInterface
	userRepo            repositories.UserRepositoryInterface
	notificationService NotificationServiceInterface
	policy              *bluemonday.Policy
}

// NewChatService can be used as a constructor to create a ChatService "object"
func NewChatService(
	chatRepo repositories.ChatRepositoryInterface,
	userRepo repositories.UserRepositoryInterface,
	notificationService NotificationServiceInterface) *ChatService {
	return &ChatService{chatRepo: chatRepo, userRepo: userRepo, notificationService: notificationService, policy: bluemonday.UGCPolicy()}
}

// CreateChat creates a chat for a given post id, username and the current logged-in user
func (service *ChatService) CreateChat(req *models.ChatCreateRequestDTO, currentUsername string) (*models.ChatCreateResponseDTO, *customerrors.CustomError, int) {
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
			return nil, customerrors.Unauthorized, http.StatusUnauthorized
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

	// Send notification to other user
	_ = service.notificationService.CreateNotification("message", req.Username, currentUsername) // ignore creation/sending error for current user

	// Create response
	response := &models.ChatCreateResponseDTO{
		ChatId: newChat.Id.String(),
		Message: &models.MessageRecordDTO{
			Content:      firstMessage.Content,
			Username:     firstMessage.Username,
			CreationDate: firstMessage.CreatedAt,
		},
	}

	return response, nil, http.StatusCreated
}

// GetChatsByUsername retrieves all chats of a user by its username
func (service *ChatService) GetChatsByUsername(username string) (*models.ChatsResponseDTO, *customerrors.CustomError, int) {
	// Get Chats by username
	chats, err := service.chatRepo.GetChatsByUsername(username)
	if err != nil {
		return nil, customerrors.DatabaseError, http.StatusInternalServerError
	}

	// Create response
	chatDTOs := make([]models.ChatRecordDTO, 0)
	for _, chat := range chats {

		// Currently, we only have two users in a chat --> find the other user
		var chatUserDto models.UserDTO
		for _, user := range chat.Users {
			var imageDto *models.ImageMetadataDTO
			if user.ImageId != nil {
				imageDto = &models.ImageMetadataDTO{
					Url:    utils.FormatImageUrl(user.ImageId.String(), user.Image.Format),
					Width:  user.Image.Width,
					Height: user.Image.Height,
					Tag:    user.Image.Tag,
				}
			}
			if user.Username != username {
				chatUserDto = models.UserDTO{
					Username: user.Username,
					Nickname: user.Nickname,
					Picture:  imageDto,
				}
				break
			}
		}

		chatDTO := models.ChatRecordDTO{
			ChatId: chat.Id.String(),
			User:   &chatUserDto,
		}

		chatDTOs = append(chatDTOs, chatDTO)
	}

	response := models.ChatsResponseDTO{
		Records: chatDTOs,
	}

	return &response, nil, http.StatusOK
}
