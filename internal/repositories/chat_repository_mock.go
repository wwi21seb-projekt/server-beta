package repositories

import (
	"github.com/stretchr/testify/mock"
	"github.com/wwi21seb-projekt/server-beta/internal/models"
)

type MockChatRepository struct {
	mock.Mock
}

func (m *MockChatRepository) CreateChat(chat *models.Chat) error {
	args := m.Called(chat)
	return args.Error(0)
}

func (m *MockChatRepository) GetChatMessages(username string, offset, limit int) ([]models.Chat, error) {
	args := m.Called(username, offset, limit)
	return args.Get(0).([]models.Chat), args.Error(1)
}

func (m *MockChatRepository) GetAllChats(username string) ([]models.Chat, error) {
	args := m.Called(username)
	return args.Get(0).([]models.Chat), args.Error(1)
}
