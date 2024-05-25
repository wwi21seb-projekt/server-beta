package services

import (
	"errors"
	"github.com/wwi21seb-projekt/server-beta/internal/customerrors"
	"github.com/wwi21seb-projekt/server-beta/internal/models"
	"github.com/wwi21seb-projekt/server-beta/internal/repositories"
	"net/http"
)

type ChatServiceInterface interface {
	CreateChat(username string, chatMessage models.Chat) (*customerrors.CustomError, int)
	GetChatMessages(username string, offset int, limit int) ([]models.Chat, *customerrors.CustomError, int)
	GetAllChats(username string) ([]models.Chat, *customerrors.CustomError, int)
}

type ChatService struct {
	chatRepository repositories.ChatRepositoryInterface
}

func NewChatService(chatRepository repositories.ChatRepositoryInterface) *ChatService {
	return &ChatService{chatRepository: chatRepository}
}

func (s *ChatService) CreateChat(string, models.Chat) (*customerrors.CustomError, int) {
	err := s.chatRepository.CreateChat(&models.Chat{})
	if err != nil {
		return customerrors.DatabaseError, http.StatusInternalServerError
	}
	return nil, http.StatusCreated
}

func (s *ChatService) GetChatMessages(username string, offset int, limit int) ([]models.Chat, *customerrors.CustomError, int) {
	chats, err := s.chatRepository.GetChatMessages(username, offset, limit)
	if err != nil {
		if errors.Is(err, customerrors.PostNotFound) {
			return nil, customerrors.PostNotFound, http.StatusNotFound
		}
		return nil, customerrors.DatabaseError, http.StatusInternalServerError
	}
	return chats, nil, http.StatusOK
}

func (s *ChatService) GetAllChats(username string) ([]models.Chat, *customerrors.CustomError, int) {
	chats, err := s.chatRepository.GetAllChats(username)
	if err != nil {
		if errors.Is(err, customerrors.PostNotFound) {
			return nil, customerrors.PostNotFound, http.StatusNotFound
		}
		return nil, customerrors.DatabaseError, http.StatusInternalServerError
	}
	return chats, nil, http.StatusOK
}
