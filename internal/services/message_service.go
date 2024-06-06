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

type MessageServiceInterface interface {
	GetChatById(chatId string, currentUsername string) (*models.Chat, *customerrors.CustomError, int)
	GetMessagesByChatId(chatId, currentUsername string, offset, limit int) (*models.MessagesResponseDTO, *customerrors.CustomError, int)
	CreateMessage(chatId, currentUsername string, req *models.MessageCreateRequestDTO, connectedParticipants []string) (*models.MessageRecordDTO, *customerrors.CustomError, int)
}

type MessageService struct {
	messageRepo         repositories.MessageRepositoryInterface
	chatRepo            repositories.ChatRepositoryInterface
	notificationService NotificationServiceInterface
	policy              *bluemonday.Policy
}

// NewMessageService can be used as a constructor to create a MessageService "object"
func NewMessageService(messageRepo repositories.MessageRepositoryInterface, chatRepo repositories.ChatRepositoryInterface, notificationService NotificationServiceInterface) *MessageService {
	return &MessageService{messageRepo: messageRepo, chatRepo: chatRepo, notificationService: notificationService, policy: bluemonday.UGCPolicy()}
}

// GetChatById retrieves a chat by its chatId and checks if the current user is a participant of the chat
func (service *MessageService) GetChatById(chatId string, currentUsername string) (*models.Chat, *customerrors.CustomError, int) {
	// Get chat by chatId
	chat, err := service.chatRepo.GetChatById(chatId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, customerrors.ChatNotFound, http.StatusNotFound
		}
		return nil, customerrors.DatabaseError, http.StatusInternalServerError
	}

	// Check if current user is a participant of the chat
	isParticipant := false
	for _, user := range chat.Users {
		if user.Username == currentUsername {
			isParticipant = true
			break
		}
	}
	if !isParticipant {
		return nil, customerrors.ChatNotFound, http.StatusNotFound // if user is not a participant of the chat, send 404
	}

	return &chat, nil, http.StatusOK
}

// GetMessagesByChatId retrieves all messages of a chat by its chatId
func (service *MessageService) GetMessagesByChatId(chatId, currentUsername string, offset, limit int) (*models.MessagesResponseDTO, *customerrors.CustomError, int) {
	// Get chat by chatId, also checks if current user is a participant of the chat
	_, serviceErr, httpStatus := service.GetChatById(chatId, currentUsername)
	if serviceErr != nil {
		return nil, serviceErr, httpStatus
	}

	// Get messages by chatId
	messages, totalCount, err := service.messageRepo.GetMessagesByChatId(chatId, offset, limit)
	if err != nil {
		return nil, customerrors.DatabaseError, http.StatusInternalServerError
	}

	// Create response DTO
	records := make([]models.MessageRecordDTO, 0)
	for _, message := range messages {
		record := models.MessageRecordDTO{
			Content:      message.Content,
			Username:     message.Username,
			CreationDate: message.CreatedAt,
		}
		records = append(records, record)
	}

	response := models.MessagesResponseDTO{
		Records: records,
		Pagination: &models.MessagePaginationDTO{
			Offset:  offset,
			Limit:   limit,
			Records: totalCount,
		},
	}

	return &response, nil, http.StatusOK
}

// CreateMessage creates a new message for a given chatId and username
func (service *MessageService) CreateMessage(chatId, currentUsername string, req *models.MessageCreateRequestDTO, connectedParticipants []string) (*models.MessageRecordDTO, *customerrors.CustomError, int) {
	// Sanitize message content because it is a free text field
	req.Content = strings.Trim(req.Content, " ") // remove leading and trailing whitespaces
	req.Content = service.policy.Sanitize(req.Content)

	// Validate input
	if len(req.Content) <= 0 || len(req.Content) > 256 {
		return nil, customerrors.BadRequest, http.StatusBadRequest
	}

	// Get chat by chatId
	chat, err := service.chatRepo.GetChatById(chatId)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, customerrors.ChatNotFound, http.StatusNotFound
		}
		return nil, customerrors.DatabaseError, http.StatusInternalServerError
	}

	// Check if current user is a participant of the chat
	isParticipant := false
	for _, user := range chat.Users {
		if user.Username == currentUsername {
			isParticipant = true
			break
		}
	}
	if !isParticipant { // maybe not reachable because user cannot create websocket connection to chat if not participant
		return nil, customerrors.ChatNotFound, http.StatusNotFound // if user is not a participant of the chat, send 404
	}

	// Create message
	message := models.Message{
		Id:        uuid.New(),
		ChatId:    chat.Id,
		Username:  currentUsername,
		Content:   req.Content,
		CreatedAt: time.Now(),
	}

	// Save message
	err = service.messageRepo.CreateMessage(&message)
	if err != nil {
		return nil, customerrors.DatabaseError, http.StatusInternalServerError
	}

	// Send notifications to other chat participants that have no active websocket connection
	for _, user := range chat.Users {
		if user.Username != currentUsername && !contains(connectedParticipants, user.Username) {
			_ = service.notificationService.CreateNotification("message", user.Username, currentUsername) // ignore creation/sending error for current user
		}
	}

	// Create response DTO
	response := models.MessageRecordDTO{
		Content:      message.Content,
		Username:     message.Username,
		CreationDate: message.CreatedAt,
	}

	return &response, nil, http.StatusCreated
}

// contains checks if a slice contains a specific string
func contains(slice []string, item string) bool {
	for _, a := range slice {
		if a == item {
			return true
		}
	}
	return false
}
