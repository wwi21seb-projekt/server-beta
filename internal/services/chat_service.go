package services

import (
	"errors"
	"github.com/wwi21seb-projekt/server-beta/internal/customerrors"
	"github.com/wwi21seb-projekt/server-beta/internal/models"
	"github.com/wwi21seb-projekt/server-beta/internal/repositories"
	"net/http"
)

type ChatServiceInterface interface {
	GetChatMessages(chatId string, username string, offset int, limit int) ([]models.MessageRecordDTO, *customerrors.CustomError, int)
	GetAllChats(username string) ([]models.ChatDTO, *customerrors.CustomError, int)
}

type ChatService struct {
	chatRepository repositories.ChatRepositoryInterface
}

func NewChatService(chatRepository repositories.ChatRepositoryInterface) *ChatService {
	return &ChatService{chatRepository: chatRepository}
}

func (s *ChatService) GetChatMessages(chatId string, username string, offset int, limit int) ([]models.MessageRecordDTO, *customerrors.CustomError, int) {
	chat, err := s.chatRepository.GetChatMessages(chatId, offset, limit)
	if err != nil {
		if errors.Is(err, customerrors.ChatNotFound) {
			return nil, customerrors.ChatNotFound, http.StatusNotFound
		}
		return nil, customerrors.DatabaseError, http.StatusInternalServerError
	}

	isParticipant := false
	for _, user := range chat {
		if user.Username == username {
			isParticipant = true
			break
		}
	}
	if !isParticipant {
		return nil, customerrors.UserUnauthorized, http.StatusUnauthorized
	}

	messages, err := s.chatRepository.GetChatMessages(chatId, offset, limit)
	if err != nil {
		if errors.Is(err, customerrors.ChatNotFound) {
			return nil, customerrors.ChatNotFound, http.StatusNotFound
		}
		return nil, customerrors.DatabaseError, http.StatusInternalServerError
	}

	var messageDTOs []models.MessageRecordDTO
	for _, message := range messages {
		messageDTO := models.MessageRecordDTO{
			Id:        message.Id,
			Content:   message.Content,
			Username:  message.Username,
			CreatedAt: message.CreatedAt,
		}
		messageDTOs = append(messageDTOs, messageDTO)
	}

	return messageDTOs, nil, http.StatusOK
}

func (s *ChatService) GetAllChats(username string) ([]models.ChatDTO, *customerrors.CustomError, int) {
	chats, err := s.chatRepository.GetAllChats(username)
	if err != nil {
		if errors.Is(err, customerrors.ChatNotFound) {
			return nil, customerrors.ChatNotFound, http.StatusNotFound
		}
		return nil, customerrors.DatabaseError, http.StatusInternalServerError
	}

	var chatDTOs []models.ChatDTO
	for _, chat := range chats {
		var users []string
		for _, user := range chat.Users {
			users = append(users, user.Username)
		}
		chatDTO := models.ChatDTO{
			Id:        chat.Id,
			Users:     users,
			CreatedAt: chat.CreatedAt,
		}
		chatDTOs = append(chatDTOs, chatDTO)
	}

	return chatDTOs, nil, http.StatusOK
}
