package services

import (
	"errors"
	"fmt"
	"github.com/wwi21seb-projekt/server-beta/internal/customerrors"
	"github.com/wwi21seb-projekt/server-beta/internal/models"
	"github.com/wwi21seb-projekt/server-beta/internal/repositories"
	"gorm.io/gorm"
	"net/http"
)

type MessageServiceInterface interface {
	GetMessagesByChatId(chatId, currentUsername string, offset, limit int) (*models.MessagesResponseDTO, *customerrors.CustomError, int)
}

type MessageService struct {
	messageRepo repositories.MessageRepositoryInterface
	chatRepo    repositories.ChatRepositoryInterface
}

// NewMessageService can be used as a constructor to create a MessageService "object"
func NewMessageService(messageRepo repositories.MessageRepositoryInterface, chatRepo repositories.ChatRepositoryInterface) *MessageService {
	return &MessageService{messageRepo: messageRepo, chatRepo: chatRepo}
}

// GetMessagesByChatId retrieves all messages of a chat by its chatId
func (service *MessageService) GetMessagesByChatId(chatId, currentUsername string, offset, limit int) (*models.MessagesResponseDTO, *customerrors.CustomError, int) {
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

	// Get messages by chatId
	messages, totalCount, err := service.messageRepo.GetMessagesByChatId(chatId, offset, limit)
	if err != nil {
		fmt.Println("error: ", err)
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
